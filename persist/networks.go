package persist

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// AddEncoder adds a new encoder for the given network
func AddEncoder(encoder Encoder, network Network) (*Encoder, error) {
	query := `
    INSERT INTO encoder (
      name, ip_address, port, status, network_id
    ) VALUES (
      ?, ?, ?, ?, ?
    );
  `
	result, err := db.Exec(query, encoder.Name, encoder.IPAddress, encoder.Port, encoder.Status, network.ID)
	if err != nil {
		return nil, err
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	newEncoder := Encoder{
		ID:        sql.NullInt64{Int64: rowID, Valid: true},
		Name:      encoder.Name,
		IPAddress: encoder.IPAddress,
		Port:      encoder.Port,
		Status:    encoder.Status,
		NetworkID: network.ID,
	}

	return &newEncoder, nil
}

// EncoderToJSON Removes SQL fields and transforms to a []byte of JSON data.
func EncoderToJSON(encoder Encoder) ([]byte, error) {
	if !encoder.ID.Valid || !encoder.Name.Valid {
		return nil, errors.New("Error validating Encoder")
	}

	newEncoder := struct {
		ID        int64
		Name      string
		IPAddress string
		Port      int
		Status    int
		NetworkID int
	}{
		encoder.ID.Int64,
		encoder.Name.String,
		encoder.IPAddress,
		encoder.Port,
		encoder.Status,
		encoder.NetworkID,
	}

	return json.Marshal(newEncoder)
}
