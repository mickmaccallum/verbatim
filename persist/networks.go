package persist

import (
	"encoding/json"
	"errors"

	"github.com/0x7fffffff/verbatim/model"
)

// AddEncoder adds a new encoder for the given network
func AddEncoder(encoder model.Encoder, network model.Network) (*model.Encoder, error) {
	query := `
    INSERT INTO encoder (
      ip_address, port, name, handle, password, network_id
    ) VALUES (
      ?, ?, ?, ?, ?, ?
    );
  `

	result, err := db.Exec(query, encoder.IPAddress, encoder.Port, encoder.Name, encoder.Handle, encoder.Password, encoder.NetworkID)

	if err != nil {
		return nil, err
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	newEncoder := model.Encoder{
		ID:        int(rowID),
		IPAddress: encoder.IPAddress,
		Port:      encoder.Port,
		Name:      encoder.Name,
		Handle:    encoder.Handle,
		Password:  encoder.Password,
		NetworkID: network.ID,
	}

	return &newEncoder, nil
}

// EncoderToJSON Removes SQL fields and transforms to a []byte of JSON data.
func EncoderToJSON(encoder model.Encoder) ([]byte, error) {
	if !encoder.Name.Valid {
		return nil, errors.New("Error validating Encoder")
	}

	newEncoder := struct {
		ID        int
		IPAddress string
		Port      int
		Name      string
		Handle    string
		Password  string
		NetworkID int
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
