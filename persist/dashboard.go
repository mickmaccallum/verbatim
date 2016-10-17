package persist

import (
	"database/sql"
	"errors"
	"log"

	"github.com/0x7fffffff/verbatim/model"
)

// GetEncoder Gets the Encoder with a given identifier.
func GetEncoder(id int) (*model.Encoder, error) {
	query := `
		SELECT id, ip_address, port, name, handle, password, network_id
		FROM encoder
		WHERE id = ?
	`

	row := db.QueryRow(query, id)
	if row == nil {
		return nil, errors.New("Encoder not found")
	}

	var encoder model.Encoder
	if err := row.Scan(
		&encoder.ID,
		&encoder.IPAddress,
		&encoder.Port,
		&encoder.Name,
		&encoder.Handle,
		&encoder.Password,
		&encoder.NetworkID,
	); err != nil {
		return nil, errors.New("Failed to find specified Encoder")
	}

	return &encoder, nil
}

// GetEncoders Get all of the encoders
func GetEncoders() ([]model.Encoder, error) {
	query := `
		SELECT id, ip_address, port, name, handle, password, network_id
		FROM encoder
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	return queryEncoders(rows)
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

	return queryEncoders(rows)
}

func queryEncoders(rows *sql.Rows) ([]model.Encoder, error) {
	var encoders = make([]model.Encoder, 0)

	for rows.Next() {
		var encoder model.Encoder

		if err := rows.Scan(
			&encoder.ID,
			&encoder.IPAddress,
			&encoder.Port,
			&encoder.Name,
			&encoder.Handle,
			&encoder.Password,
			&encoder.NetworkID,
		); err != nil {
			log.Fatal(err)
			continue
		}

		encoders = append(encoders, encoder)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return encoders, nil
}

// UpdateEncoder updates all fields for the given Encoder.
func UpdateEncoder(encoder model.Encoder) error {
	query := `
		UPDATE encoder
			SET
				ip_address = ?,
				port = ?,
				name = ?,
				handle = ?,
				password = ?,
				network_id = ?
			WHERE
				id = ?
	`

	_, err := db.Exec(query, encoder.IPAddress, encoder.Port, encoder.Name, encoder.Handle, encoder.Password, encoder.NetworkID, encoder.ID)
	return err
}

// DeleteEncoder deletes the specified Encoder.
func DeleteEncoder(encoder model.Encoder) error {
	query := `
		DELETE from encoder
		WHERE id = ?
	`

	_, err := db.Exec(query, encoder.ID)
	return err
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

// AddNetwork Adds the given Network.
func AddNetwork(network model.Network) (*model.Network, error) {
	query := `
		INSERT INTO network (
			listening_port, name
		) VALUES (
			?, ?
		);
	`

	result, err := db.Exec(query, network.ListeningPort, network.Name)
	if err != nil {
		return nil, err
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	newNetwork := model.Network{
		ID:            int(rowID),
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

	_, err := db.Exec(query, network.Name, network.ListeningPort, network.ID)
	return err
}

// DeleteNetwork deletes the specified Network.
func DeleteNetwork(network model.Network) error {
	transaction, err := db.Begin()
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
