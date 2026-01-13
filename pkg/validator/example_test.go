package validator_test

import (
	"fmt"

	"github.com/ArgonautPath/go-kit/pkg/validator"
)

func ExampleValidateStruct() {
	type User struct {
		Email    string `validate:"required,email"`
		Age      int    `validate:"required,range=18,120"`
		Password string `validate:"required,min=8"`
	}

	user := User{
		Email:    "test@example.com",
		Age:      25,
		Password: "password123",
	}

	err := validator.ValidateStruct(user)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation passed")
	// Output: Validation passed
}

func ExampleValidateStruct_multipleErrors() {
	type User struct {
		Email    string `validate:"required,email"`
		Age      int    `validate:"required,range=18,120"`
		Password string `validate:"required,min=8"`
	}

	user := User{
		Email:    "invalid-email",
		Age:      15,
		Password: "short",
	}

	err := validator.ValidateStruct(user)
	if err != nil {
		if validationErrs, ok := validator.AsValidationErrors(err); ok {
			for _, e := range validationErrs.Errors() {
				fmt.Printf("Field %s: %s\n", e.Field, e.Message)
			}
		}
	}

	// Output:
	// Field Email: invalid email format
	// Field Age: value 15 is less than minimum 18
	// Field Password: length 5 is less than minimum 8
}

func ExampleValidateStruct_nested() {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
		Zip    string `validate:"required,numeric,len=5"`
	}

	type User struct {
		Name    string  `validate:"required"`
		Email   string  `validate:"required,email"`
		Address Address
	}

	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
			Zip:    "10001",
		},
	}

	err := validator.ValidateStruct(user)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation passed")
	// Output: Validation passed
}

func ExampleValidateStruct_oneOf() {
	type Config struct {
		Environment string `validate:"required,oneof=development,staging,production"`
		LogLevel    string `validate:"required,oneof=debug,info,warn,error"`
	}

	config := Config{
		Environment: "production",
		LogLevel:    "info",
	}

	err := validator.ValidateStruct(config)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation passed")
	// Output: Validation passed
}

func ExampleValidateStruct_regex() {
	type User struct {
		Username string `validate:"required,regex=^[a-zA-Z0-9_]+$"`
		Phone    string `validate:"required,regex=^[0-9]{10}$"`
	}

	user := User{
		Username: "john_doe123",
		Phone:    "1234567890",
	}

	err := validator.ValidateStruct(user)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation passed")
	// Output: Validation passed
}

func ExampleRegisterValidator() {
	// Register a custom validator
	validator.RegisterValidator("custom", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}
		if str != "valid" {
			return fmt.Errorf("value must be 'valid'")
		}
		return nil
	})
	defer validator.UnregisterValidator("custom")

	type Config struct {
		Field string `validate:"required,custom"`
	}

	config := Config{Field: "valid"}

	err := validator.ValidateStruct(config)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Validation passed")
	// Output: Validation passed
}

func ExampleValidationErrors() {
	type User struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"required,range=18,120"`
	}

	user := User{
		Email: "invalid-email",
		Age:   15,
	}

	err := validator.ValidateStruct(user)
	if err != nil {
		if validationErrs, ok := validator.AsValidationErrors(err); ok {
			fmt.Printf("Found %d validation errors:\n", len(validationErrs.Errors()))
			for i, e := range validationErrs.Errors() {
				fmt.Printf("%d. %s: %s (value: %v)\n", i+1, e.Field, e.Message, e.Value)
			}
		}
	}

	// Output:
	// Found 2 validation errors:
	// 1. Email: invalid email format (value: invalid-email)
	// 2. Age: value 15 is less than minimum 18 (value: 15)
}
