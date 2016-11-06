package persist

import (
	"errors"
	"log"

	"github.com/0x7fffffff/verbatim/model"
)

// GetNetwork gets the Network for a given id.
func GetNetwork(id int) (*model.Network, error) {
	query := `
		SELECT id, listening_port, name
		FROM network
		WHERE id = ?
	`

	row := DB.QueryRow(query, id)
	if row == nil {
		return nil, errors.New("Network not found")
	}

	var net model.Network
	if err := row.Scan(&net.ID, &net.ListeningPort, &net.Name); err != nil {
		return nil, errors.New("Failed to find specified Network")
	}

	return &net, nil
}

// GetNetworks Gets all Networks in the database.
func GetNetworks() ([]model.Network, error) {
	query := `
		SELECT id, listening_port, name
		FROM network
	`
	rows, err := DB.Query(query)

	if err != nil {
		return nil, err
	}

	var networks = make([]model.Network, 0)

	for rows.Next() {
		var net model.Network

		if err = rows.Scan(&net.ID, &net.ListeningPort, &net.Name); err != nil {
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

// AddNetwork Adds the given Network.
func AddNetwork(network model.Network) (*model.Network, error) {
	query := `
		INSERT INTO network (
			listening_port, name
		) VALUES (
			?, ?
		);
	`

	result, err := DB.Exec(query, network.ListeningPort, network.Name)
	if err != nil {
		return nil, err
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	newNetwork := model.Network{
		ID:            model.NetworkID(rowID),
		Name:          network.Name,
		ListeningPort: network.ListeningPort,
	}

	return &newNetwork, nil
}

// UpdateNetwork update the info for a given Network
func UpdateNetwork(network model.Network) error {
	query := `
		UPDATE network
			SET
				name = ?,
				listening_port = ?
			WHERE
				id = ?
	`

	_, err := DB.Exec(query, network.Name, network.ListeningPort, network.ID)
	return err
}

// DeleteNetwork deletes the specified Network.
func DeleteNetwork(network model.Network) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	deleteNetwork := `
		DELETE from network
		WHERE id = ?
	`

	deleteEncoders := `
		DELETE from encoder
		WHERE network_id = ?
	`

	_, err = transaction.Exec(deleteEncoders, network.ID)
	_, err = transaction.Exec(deleteNetwork, network.ID)
	if err != nil {
		return transaction.Rollback()
	}

	return transaction.Commit()
}
