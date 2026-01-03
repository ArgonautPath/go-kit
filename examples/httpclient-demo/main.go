package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/httpclient"
)

func main() {
	ctx := context.Background()

	// Example 1: Create a generic HTTP client
	fmt.Println("=== Example 1: Basic HTTP Client ===")
	client, err := httpclient.NewGeneric(httpclient.Config{
		BaseURL:        "https://jsonplaceholder.typicode.com",
		DefaultTimeout: 30 * time.Second,
		DefaultHeaders: map[string]string{
			"User-Agent": "go-kit-example/1.0",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 2: GET request with type-safe response
	fmt.Println("\n=== Example 2: GET Request with Type-Safe Response ===")
	type Post struct {
		UserID int    `json:"userId"`
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	resp, err := httpclient.Get[Post](client, ctx, "/posts/1")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Post ID: %d\n", resp.Body.ID)
		fmt.Printf("Title: %s\n", resp.Body.Title)
		fmt.Printf("User ID: %d\n", resp.Body.UserID)
	}

	// Example 3: GET request with query parameters
	fmt.Println("\n=== Example 3: GET Request with Query Parameters ===")
	type Comment struct {
		PostID int    `json:"postId"`
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Body   string `json:"body"`
	}

	comments, err := httpclient.Get[[]Comment](client, ctx, "/comments",
		httpclient.WithQuery(map[string]string{
			"postId": "1",
		}),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", comments.StatusCode)
		fmt.Printf("Found %d comments\n", len(comments.Body))
		if len(comments.Body) > 0 {
			fmt.Printf("First comment: %s\n", comments.Body[0].Name)
		}
	}

	// Example 4: POST request with body
	fmt.Println("\n=== Example 4: POST Request with Body ===")
	type CreatePost struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}

	newPost := CreatePost{
		Title:  "Example Post",
		Body:   "This is an example post created with go-kit httpclient",
		UserID: 1,
	}

	createdPost, err := httpclient.Post[Post](client, ctx, "/posts",
		httpclient.WithBody(newPost),
		httpclient.WithHeader("Content-Type", "application/json"),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", createdPost.StatusCode)
		fmt.Printf("Created Post ID: %d\n", createdPost.Body.ID)
		fmt.Printf("Title: %s\n", createdPost.Body.Title)
	}

	// Example 5: PUT request
	fmt.Println("\n=== Example 5: PUT Request ===")
	updatePost := CreatePost{
		Title:  "Updated Post",
		Body:   "This post has been updated",
		UserID: 1,
	}

	updatedPost, err := httpclient.Put[Post](client, ctx, "/posts/1",
		httpclient.WithBody(updatePost),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", updatedPost.StatusCode)
		fmt.Printf("Updated Post Title: %s\n", updatedPost.Body.Title)
	}

	// Example 6: DELETE request
	fmt.Println("\n=== Example 6: DELETE Request ===")
	deleteResp, err := httpclient.Delete[map[string]interface{}](client, ctx, "/posts/1")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", deleteResp.StatusCode)
		fmt.Println("Post deleted successfully")
	}

	// Example 7: Request with custom timeout
	fmt.Println("\n=== Example 7: Request with Custom Timeout ===")
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fastResp, err := httpclient.Get[Post](client, timeoutCtx, "/posts/1",
		httpclient.WithTimeout(2 * time.Second),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d (completed within timeout)\n", fastResp.StatusCode)
	}

	// Example 8: Request with custom headers
	fmt.Println("\n=== Example 8: Request with Custom Headers ===")
	customResp, err := httpclient.Get[Post](client, ctx, "/posts/1",
		httpclient.WithHeaders(map[string]string{
			"X-Custom-Header": "custom-value",
			"Accept":          "application/json",
		}),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Status: %d\n", customResp.StatusCode)
		fmt.Printf("Response Headers: %v\n", customResp.Headers)
	}

	// Example 9: Error handling
	fmt.Println("\n=== Example 9: Error Handling ===")
	_, err = httpclient.Get[Post](client, ctx, "/nonexistent/999")
	if err != nil {
		if httpErr, ok := err.(*httpclient.HTTPError); ok {
			fmt.Printf("HTTP Error: Status %d %s\n", httpErr.StatusCode, httpErr.Status)
		} else if reqErr, ok := err.(*httpclient.RequestError); ok {
			fmt.Printf("Request Error: %v\n", reqErr)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println("\n=== All examples completed ===")
}

