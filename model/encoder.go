package model

import (
	"database/sql"
	"errors"
	"net/url"
	"strconv"

	"github.com/0x7fffffff/verbatim/states"
)

// EncoderID represents the id of an encoder.
type EncoderID int

// Encoder represents a single downstream encoder for a given network
type Encoder struct {
	ID        EncoderID
	IPAddress string
	Port      int
	Name      sql.NullString
	Handle    string
	Password  string
	NetworkID NetworkID
	Status    states.Encoder
}

// FormValuesToEncoder validates that an Encoder can be created
// from the given form values and creates it.
func FormValuesToEncoder(values url.Values) (*Encoder, error) {
	ipAddress, portString, name, handle, password, networkIDString :=
		values.Get("ip_address"),
		values.Get("port"),
		values.Get("name"),
		values.Get("handle"),
		values.Get("password"),
		values.Get("network_id")

	// validate IPv4 & IPv6 addresses
	valid, err := isValidIp(ipAddress)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("Invalid IP address")
	}

	// Ports [1, 65535]
	if len(portString) < 1 || len(portString) > 5 {
		return nil, errors.New("Invalid port")
	}

	if len(name) > 255 {
		return nil, errors.New("Name is too long")
	}

	if len(handle) == 0 || len(handle) > 255 {
		return nil, errors.New("Invalid handle")
	}

	if len(password) == 0 || len(password) > 255 {
		return nil, errors.New("Invalid password")
	}

	if len(networkIDString) == 0 || len(networkIDString) > 10 {
		return nil, errors.New("Invalid Network ID")
	}

	port, err := strconv.Atoi(portString)
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("Invalid port")
	}

	networkID, err := strconv.Atoi(networkIDString)
	if err != nil {
		return nil, errors.New("Invalid Network ID")
	}

	encoder := Encoder{
		IPAddress: ipAddress,
		Port:      port,
		Name:      sql.NullString{String: name, Valid: true},
		Handle:    handle,
		Password:  password,
		NetworkID: NetworkID(networkID),
	}

	return &encoder, nil
}
