package model

import (
	"database/sql"
	"errors"
	"net/url"
	"regexp"
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

	// Min length of IPv6, max length of IPv6.
	if len(ipAddress) < 3 || len(ipAddress) > 45 {
		return nil, errors.New("Invalid IP Address length")
	}

	// Source: https://www.safaribooksonline.com/library/view/regular-expressions-cookbook/9780596802837/ch07s16.html
	ipv4Pattern := "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"

	// Source: http://stackoverflow.com/a/17871737/716216
	ipv6Pattern := "(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))"

	// validate IPv4 address
	match, err := regexp.MatchString(ipv4Pattern, ipAddress)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, errors.New("Invalid IP address")
	}

	match, err = regexp.MatchString(ipv6Pattern, ipAddress)
	if err != nil {
		return nil, err
	}

	if !match {
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
