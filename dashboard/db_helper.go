package dashboard

import (
	"database/sql"
	"errors"
	"log"
)

// Network represents a downstream network
type Network struct {
	ID   sql.NullInt64
	Name sql.NullString
}

func getNetwork(id int) (*Network, error) {
	query := `
		SELECT id, name
		FROM network
		WHERE id = ?
	`
	row := db.QueryRow(query, id)
	if row == nil {
		return nil, errors.New("Network not found")
	}

	var net Network
	if err := row.Scan(&net.ID, &net.Name); err != nil {
		return nil, errors.New("Failed to create Network from query")
	}

	return &net, nil
}

func getNetworks() ([]Network, error) {
	rows, err := db.Query("select id, name from network;")

	if err != nil {
		return nil, err
	}

	var networks = make([]Network, 0)

	for rows.Next() {
		var net Network

		if err = rows.Scan(&net.ID, &net.Name); err != nil {
			log.Fatal(err)
			continue
		}

		networks = append(networks, net)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	return networks, nil
}
