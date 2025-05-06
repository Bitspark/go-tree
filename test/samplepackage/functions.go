package samplepackage

import (
	"errors"
	"fmt"
	"time"
)

// Global variables
var (
	// DefaultTimeout is used when no timeout is specified
	DefaultTimeout = 30 * time.Second

	// Error messages
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrPermissionDenied   = errors.New("permission denied")
)

// NewUser creates a new user with default values
func NewUser(name string) *User {
	return &User{
		Name: name,
		Authentication: Authentication{
			Username: name,
		},
		timestamps: timestamps{
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
	}
}

// UpdatePassword updates a user's password
func (u *User) UpdatePassword(password string) error {
	// Check password strength
	if len(password) < 8 {
		return fmt.Errorf("password too short: must be at least 8 characters")
	}

	// Update password and timestamp
	u.Password = password
	u.UpdatedAt = time.Now().Unix()

	return nil
}

// Login implements the Authenticator interface
func (u *User) Login(username, password string) (bool, error) {
	if username != u.Username || password != u.Password {
		return false, ErrInvalidCredentials
	}
	return true, nil
}

// Logout implements the Authenticator interface
func (u *User) Logout() error {
	// Implement logout logic
	return nil
}

// Validate implements the Validator interface
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// FormatUser returns a formatted string representation of the user
func FormatUser(u *User) string {
	return fmt.Sprintf("User %s (ID: %d)", u.Name, u.ID)
}
