package model

import (
	"errors"
	"github.com/0x7fffffff/verbatim/states"
	"net/url"
	"strconv"
)

// NetworkID lint
type NetworkID int

// Network represents a downstream network
type Network struct {
	ID            NetworkID
	ListeningPort int
	Name          string
	Timeout       int
	State         states.Network
}

// FormValuesToNetwork validates that a Network can be created
// from the given form values and creates it.
func FormValuesToNetwork(values url.Values) (*Network, error) {
	portString, name, timeoutString :=
		values.Get("listening_port"),
		values.Get("name"),
		values.Get("timeout")

	// Ports [1, 65535]
	if len(portString) < 1 || len(portString) > 5 {
		return nil, errors.New("Invalid Network Port")
	}

	if len(name) < 1 || len(name) > 255 {
		return nil, errors.New("Invalid Network Name")
	}

	if len(timeoutString) > 4 {
		return nil, errors.New("Invalid Timeout")
	}

	port, err := strconv.Atoi(portString)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("Invalid port")
	}

	timeout, err := strconv.Atoi(timeoutString)
	if err != nil {
		return nil, errors.New("Invalid Timeout")
	}

	if timeout != 0 && (timeout < 10 || timeout > 3600) {
		return nil, errors.New("Invalid Timeout")
	}

	network := Network{
		Name:          name,
		ListeningPort: port,
		Timeout:       timeout,
	}

	return &network, nil
}
