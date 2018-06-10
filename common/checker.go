package common // import "breve.us/authsvc/common"

import (
	"context"
	"net/http"
)

//
// Request Checkers
//

// RequestChecker describes functionality to verify requests
type RequestChecker interface {
	IsAuthenticated(r *http.Request) string
}

// RequestCheckers combines multiple RequestCheckers
func RequestCheckers(checkers ...RequestChecker) RequestChecker {
	var valid []RequestChecker
	for _, cc := range checkers {
		if cc != nil {
			valid = append(valid, cc)
		}
	}
	return &requestChecker{checkers: valid}
}

type requestChecker struct {
	checkers []RequestChecker
}

func (c *requestChecker) IsAuthenticated(r *http.Request) string {
	for _, cc := range c.checkers {
		if username := cc.IsAuthenticated(r); username != "" {
			return username
		}
	}
	return ""
}

//
// Password Checkers
//

// PasswordChecker describes functionality to verify passwords
type PasswordChecker interface {
	IsAuthenticated(username string, password string) bool
}

// PasswordCheckers combines multiple PasswordCheckers
func PasswordCheckers(checkers ...PasswordChecker) PasswordChecker {
	var valid []PasswordChecker
	for _, cc := range checkers {
		if cc != nil {
			valid = append(valid, cc)
		}
	}
	return &passwordChecker{checkers: valid}
}

type passwordChecker struct {
	checkers []PasswordChecker
}

func (c *passwordChecker) IsAuthenticated(username, password string) bool {
	for _, cc := range c.checkers {
		if cc.IsAuthenticated(username, password) {
			return true
		}
	}
	return false
}

//
// context.Context helpers
//

// SetUsername returns a context with the given username set
func SetUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// GetUsername retrieves authenticated username from Context
func GetUsername(ctx context.Context) string {
	if username, ok := ctx.Value(usernameKey).(string); ok {
		return username
	}
	return ""
}

type contextKey string

const (
	usernameKey contextKey = "username"
)
