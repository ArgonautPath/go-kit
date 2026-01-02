package logger

import (
	"context"
	"sync"
	"sync/atomic"
)

// AsyncWriter wraps a Writer to provide asynchronous, non-blocking logging.
//
// Why do we need AsyncWriter?
//
// Without async logging, every log call performs synchronous I/O operations that can block
// the calling goroutine. This causes several problems:
//
//  1. Blocking under load: When logging to files or stdout/stderr, slow disk I/O or
//     full buffers can cause log calls to block for seconds, freezing your application.
//
//  2. Cascading delays: In high-concurrency scenarios (e.g., HTTP handlers), if many
//     goroutines try to log simultaneously, they all block waiting for I/O, creating
//     a bottleneck that slows down the entire application.
//
//  3. Deadlock risk: If logging blocks and your application has limited goroutine pools,
//     you can exhaust all goroutines waiting on log writes, causing deadlocks.
//
//  4. Poor user experience: In request handlers, blocking on log writes directly impacts
//     response times, making your service appear slow or unresponsive.
//
// How AsyncWriter solves this:
//
// AsyncWriter decouples log generation from log writing by:
//
//   - Queuing entries in a buffered channel instead of writing immediately
//   - Using a dedicated background goroutine to perform all I/O operations
//   - Making log calls return immediately (non-blocking) after queuing
//   - Dropping entries when the buffer is full (fail-fast) to prevent blocking
//
// This ensures that logging never blocks your application code, even when the underlying
// writer is slow or unavailable.
//
// Example problem scenario:
//
//	Without AsyncWriter:
//	  func handleRequest(w http.ResponseWriter, r *http.Request) {
//	      logger.Info(ctx, "processing") // BLOCKS here if disk is slow
//	      // Handler is frozen, user waits...
//	  }
//
//	With AsyncWriter:
//	  func handleRequest(w http.ResponseWriter, r *http.Request) {
//	      logger.Info(ctx, "processing") // Returns immediately, never blocks
//	      // Handler continues immediately, user gets fast response
//	  }
//
// Trade-offs:
//
//   - Memory usage: Buffered channel uses memory proportional to buffer size
//   - Potential data loss: Entries are dropped when buffer is full (but this prevents blocking)
//   - Eventual consistency: Logs may be written slightly after the log call returns
//
// When to use:
//
//   - High-throughput applications where logging must not impact performance
//   - Services with many concurrent goroutines that log frequently
//   - Production environments where blocking on I/O is unacceptable
//   - When you can tolerate occasional log drops under extreme load
type AsyncWriter struct {
	writer  Writer             // The underlying writer that performs actual I/O
	queue   chan *LogEntry     // Buffered channel queue for log entries
	ctx     context.Context    // Context for graceful shutdown
	cancel  context.CancelFunc // Cancel function to stop the worker
	wg      sync.WaitGroup     // WaitGroup to wait for worker goroutine
	dropped uint64             // Counter for dropped entries (atomic, thread-safe)
	mu      sync.RWMutex       // Mutex to protect closed flag
	closed  bool               // Flag indicating if writer is closed
}

// NewAsyncWriter creates a new async writer that wraps the given writer.
// bufferSize determines the size of the internal queue. When full, entries are dropped.
func NewAsyncWriter(writer Writer, bufferSize int) *AsyncWriter {
	ctx, cancel := context.WithCancel(context.Background())
	aw := &AsyncWriter{
		writer: writer,
		queue:  make(chan *LogEntry, bufferSize),
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background worker
	aw.wg.Add(1)
	go aw.worker()

	return aw
}

// Write queues a log entry for asynchronous writing. If the queue is full, the entry
// is dropped and the method returns immediately without blocking.
func (aw *AsyncWriter) Write(entry *LogEntry) error {
	aw.mu.RLock()
	closed := aw.closed
	aw.mu.RUnlock()

	if closed {
		// If closed, write synchronously as fallback
		return aw.writer.Write(entry)
	}

	// Non-blocking send - drop if channel is full
	select {
	case aw.queue <- entry:
		return nil
	default:
		// Channel is full, drop the entry and increment counter
		atomic.AddUint64(&aw.dropped, 1)
		return nil
	}
}

// worker processes log entries from the queue in the background.
func (aw *AsyncWriter) worker() {
	defer aw.wg.Done()

	for {
		select {
		case <-aw.ctx.Done():
			// Context cancelled, drain remaining entries
			aw.drain()
			return
		case entry := <-aw.queue:
			// Write entry synchronously (this is in background goroutine, so it's OK)
			_ = aw.writer.Write(entry)
		}
	}
}

// drain writes all remaining entries in the queue before shutdown.
func (aw *AsyncWriter) drain() {
	for {
		select {
		case entry := <-aw.queue:
			_ = aw.writer.Write(entry)
		default:
			// Queue is empty
			return
		}
	}
}

// Close gracefully shuts down the async writer. It stops accepting new entries,
// drains the queue, and waits for the worker goroutine to finish.
func (aw *AsyncWriter) Close() error {
	aw.mu.Lock()
	if aw.closed {
		aw.mu.Unlock()
		return nil
	}
	aw.closed = true
	aw.mu.Unlock()

	// Cancel context to stop worker
	aw.cancel()

	// Wait for worker to finish
	aw.wg.Wait()

	return nil
}

// DroppedCount returns the number of entries that were dropped due to a full buffer.
func (aw *AsyncWriter) DroppedCount() uint64 {
	return atomic.LoadUint64(&aw.dropped)
}
