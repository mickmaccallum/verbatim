package model

import (
	"errors"
	"net/url"

	"golang.org/x/crypto/bcrypt"
)

// Admin represents a downstream network
type Admin struct {
	ID             int
	Handle         string
	HashedPassword string `json:"-"`
}

// FormValuesToAdmin validates that an Admin can be created
// from the given form values and creates it.
func FormValuesToAdmin(values url.Values) (*Admin, error) {
	handle, password, confirmPassword :=
		values.Get("handle"), values.Get("password"), values.Get("confirm_password")

	if len(handle) == 0 {
		return nil, errors.New("Missing Handle")
	}

	if len(handle) > 255 {
		return nil, errors.New("Handle too long")
	}

	if password != confirmPassword {
		return nil, errors.New("Passwords do not match.")
	}

	if len(password) < 8 {
		return nil, errors.New("Password too short. Must be at least 8 characters.")
	}

	if len(password) > 255 {
		return nil, errors.New("Password too long. Must be under 255 characters.")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Password too long. Must be under 255 characters.")
	}

	admin := Admin{
		Handle:         handle,
		HashedPassword: string(hashed),
	}

	return &admin, nil
}

// HasPassword returns whether or not the given password is the same
// as the receiver's password
func (admin Admin) HasPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), []byte(password))
	return err == nil
}
