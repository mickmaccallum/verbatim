package persist

import (
	"errors"
	"log"

	"github.com/0x7fffffff/verbatim/model"
)

// GetEncoders Get all of the encoders
func GetEncoders() ([]model.Encoder, error) {
	return nil, nil
}

// GetEncodersForNetwork Gets a slice of Encoders for a given Network.
func GetEncodersForNetwork(network model.Network) ([]model.Encoder, error) {
	query := `
		SELECT id, ip_address, port, name, handle, password, network_id
		FROM encoder
		WHERE network_id = ?
	`

	rows, err := db.Query(query, network.ID)

	if err != nil {
		return nil, err
	}

	var encoders = make([]model.Encoder, 0)

	for rows.Next() {
		var encoder model.Encoder

		if err = rows.Scan(
			&encoder.ID,
			&encoder.IPAddress,
			&encoder.Port,
			&encoder.Handle,
			&encoder.Password,
			&encoder.NetworkID,
		); err != nil {
			log.Fatal(err)
			continue
		}

		encoders = append(encoders, encoder)
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	return encoders, nil
}

// GetNetwork gets the Network for a given id.
func GetNetwork(id int) (*model.Network, error) {
	query := `
		SELECT id, listening_port, name
		FROM network
		WHERE id = ?
	`

	row := db.QueryRow(query, id)
	if row == nil {
		return nil, errors.New("Network not found")
	}

	var net model.Network
	if err := row.Scan(&net.ID, &net.ListeningPort, &net.Name); err != nil {
		return nil, errors.New("Failed to find specified Network")
	}

	return &net, nil
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

	_, err := db.Exec(query, network.Name, network.ListeningPort, network.ID)
	return err
}

// DeleteNetwork deletes the specified Network.
func DeleteNetwork(network model.Network) error {
	query := `
		DELETE from network
		WHERE id = ?
	`

	_, err := db.Exec(query, network.ID)
	return err
}

// GetNetworks Gets all Networks in the database.
func GetNetworks() ([]model.Network, error) {
	query := `
		SELECT id, listening_port, name
		FROM network
	`
	rows, err := db.Query(query)

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
