package dashboard

import (
	"errors"
	"log"
)

// Network represents a downstream network
type Network struct {
	ID   int
	Name string
}

// Encoder represents a single downstream encoder for a given network
type Encoder struct {
	ID        int
	IPAddress string
	Port      int
	Status    int
	networkID int
}

func getEncodersForNetwork(network Network) ([]Encoder, error) {
	query := `
		SELECT id, ip_address, port, status, network_id
		FROM encoder
		WHERE network_id = ?
	`

	rows, err := db.Query(query, network.ID)

	if err != nil {
		return nil, err
	}

	var encoders = make([]Encoder, 0)

	for rows.Next() {
		var encoder Encoder

		if err = rows.Scan(&encoder.ID, &encoder.IPAddress, &encoder.Port, &encoder.Status, &encoder.networkID); err != nil {
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
	query := `
		SELECT id, name
		FROM network
	`
	rows, err := db.Query(query)

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
