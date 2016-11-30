package persist

import (
	"errors"

	"github.com/0x7fffffff/verbatim/model"
)

// CreateBackup Creates a backup of the given data. Data should
// represent an undelivered caption burst.
func CreateBackup(data []byte, network model.Network) error {
	query := `
	    INSERT INTO backup (
	      payload, network_id
	    ) VALUES (
	      ?, ?
	    );
	`

	backupError := errors.New("Failed to create backup")
	result, err := DB.Exec(query, string(data), int(network.ID))
	if err != nil {
		return backupError
	}

	_, err = result.LastInsertId()
	if err != nil {
		return backupError
	}

	return nil
}
