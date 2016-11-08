package persist

import (
	"errors"

	"github.com/0x7fffffff/verbatim/model"
	"golang.org/x/crypto/bcrypt"
)

// GetAdmins gets all of the administrators.
func GetAdmins() ([]model.Admin, error) {
	query := `
		SELECT id, handle, hashed_password
		FROM admin
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	var admins = make([]model.Admin, 0)

	for rows.Next() {
		var admin model.Admin

		if err := rows.Scan(
			&admin.ID,
			&admin.Handle,
			&admin.HashedPassword,
		); err != nil {
			return nil, err
		}

		admins = append(admins, admin)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return admins, nil
}

// GetAdminForID Looks up an admin by their ID.
func GetAdminForID(id int) (*model.Admin, error) {
	query := `
		SELECT id, handle, hashed_password
		FROM admin
		WHERE id = ?
	`

	row := DB.QueryRow(query, id)
	if row == nil {
		return nil, errors.New("Invalid Admin")
	}

	var admin model.Admin
	if err := row.Scan(
		&admin.ID,
		&admin.Handle,
		&admin.HashedPassword,
	); err != nil {
		return nil, errors.New("Invalid Admin")
	}

	return &admin, nil
}

// GetAdminForCredentials Looks up an admin by their login credentials.
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

// DeleteAdmin deletes the specified administrator.
func DeleteAdmin(admin model.Admin) error {
	query := `
		DELETE from admin
		WHERE id = ?
	`

	_, err := DB.Exec(query, admin.ID)
	return err
}
