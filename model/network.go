package model

import (
	"errors"
	"net/url"
	"strconv"
)

// Network represents a downstream network
type Network struct {
	ID            int
	ListeningPort int
	Name          string
}

// FormValuesToNetwork validates that a Network can be created
// from the given form values and creates it.
func FormValuesToNetwork(values url.Values) (*Network, error) {
	portString, name := values.Get("listening_port"), values.Get("name")

	// Ports [1, 65535]
	if len(portString) < 1 || len(portString) > 5 {
		return nil, errors.New("Invalid Network Port")
	}

	if len(name) < 1 || len(name) > 255 {
		return nil, errors.New("Invalid Network Name")
	}

	port, err := strconv.Atoi(portString)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("Invalid port")
	}

	network := Network{
		Name:          name,
		ListeningPort: port,
	}

	return &network, nil
}
