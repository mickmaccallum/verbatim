package persist

import (
	"errors"
	"github.com/0x7fffffff/verbatim/model"
	"golang.org/x/crypto/bcrypt"
)

func GetAdminForCredentials(handle string, password string) (*model.Admin, error) {
	query := `
		SELECT id, handle, hashed_password
		FROM admin
		WHERE handle = ?
	`

	row := DB.QueryRow(query, handle)
	if row == nil {
		return nil, errors.New("Invalid Credentials")
	}

	var admin model.Admin
	if err := row.Scan(
		&admin.ID,
		&admin.Handle,
		&admin.HashedPassword,
	); err != nil {
		return nil, errors.New("Invalid Credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.HashedPassword), []byte(password)); err != nil {
		return nil, errors.New("Incorrect Password")
	}

	return &admin, nil
}
