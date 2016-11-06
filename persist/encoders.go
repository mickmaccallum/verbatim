package persist

import (
	"database/sql"
	"encoding/json"
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

	row := DB.QueryRow(query, id)
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

	rows, err := DB.Query(query)
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

	rows, err := DB.Query(query, network.ID)
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

// AddEncoder adds a new encoder for the given network
func AddEncoder(encoder model.Encoder, network model.Network) (*model.Encoder, error) {
	query := `
    INSERT INTO encoder (
      ip_address, port, name, handle, password, network_id
    ) VALUES (
      ?, ?, ?, ?, ?, ?
    );
  `

	result, err := DB.Exec(query, encoder.IPAddress, encoder.Port, encoder.Name, encoder.Handle, encoder.Password, encoder.NetworkID)

	if err != nil {
		return nil, err
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	newEncoder := model.Encoder{
		ID:        model.EncoderID(rowID),
		IPAddress: encoder.IPAddress,
		Port:      encoder.Port,
		Name:      encoder.Name,
		Handle:    encoder.Handle,
		Password:  encoder.Password,
		NetworkID: network.ID,
	}

	return &newEncoder, nil
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

	_, err := DB.Exec(query, encoder.IPAddress, encoder.Port, encoder.Name, encoder.Handle, encoder.Password, encoder.NetworkID, encoder.ID)
	return err
}

// DeleteEncoder deletes the specified Encoder.
func DeleteEncoder(encoder model.Encoder) error {
	query := `
		DELETE from encoder
		WHERE id = ?
	`

	_, err := DB.Exec(query, encoder.ID)
	return err
}

// EncoderToJSON Removes SQL fields and transforms to a []byte of JSON data.
func EncoderToJSON(encoder model.Encoder) ([]byte, error) {
	if !encoder.Name.Valid {
		return nil, errors.New("Error validating Encoder")
	}

	newEncoder := struct {
		ID        model.EncoderID
		IPAddress string
		Port      int
		Name      string
		Handle    string
		Password  string
		NetworkID model.NetworkID
	}{
		encoder.ID,
		encoder.IPAddress,
		encoder.Port,
		encoder.Name.String,
		encoder.Handle,
		encoder.Password,
		encoder.NetworkID,
	}

	return json.Marshal(newEncoder)
}

// NetworkToJSON Removes SQL fields and transforms to a []byte of JSON data. This is not exactly needed yet, but we're using this here in case Network ever gets fields that are SQL types.
func NetworkToJSON(network model.Network) ([]byte, error) {
	return json.Marshal(network)
}
