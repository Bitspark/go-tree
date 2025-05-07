// Package samplepackage demonstrates various Go constructs for testing the parser.
package samplepackage

// User represents an application user
type User struct {
	// ID is the unique identifier for the user
	ID int `json:"id"`

	// User's name
	Name string `json:"name"`

	// Optional contact information
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty" validate:"optional"`

	// Embedded fields
	Authentication
	timestamps
}

// Authentication contains user credentials
type Authentication struct {
	Username string
	Password string // This should be hashed
}

// Private embedded type
type timestamps struct {
	CreatedAt int64
	UpdatedAt int64
}

// Role defines user permissions.
type Role string

// List of predefined roles
const (
	// RoleAdmin has full permissions.
	RoleAdmin Role = "admin"

	// RoleUser has limited permissions.
	RoleUser Role = "user"

	// RoleGuest has read-only access.
	RoleGuest Role = "guest"
)

// Interface types

// Authenticator is an interface.
type Authenticator interface {
	// Login attempts to authenticate the user
	Login(username, password string) (bool, error)

	Logout() error

	// Embedded interface

	// Validator is embedded.
	Validator
}

// Validator is an embedded interface
type Validator interface {
	Validate() error
}

// Type alias example

// UserMap is a user map.
type UserMap = map[string]*User

// Function type

// AuthHandler is an auth handler.
type AuthHandler func(username, password string) bool
