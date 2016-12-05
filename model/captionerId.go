package model

import (
	// "database/sql"
	// "encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// CaptionerID represents a unique captioner connected to microphone.
type CaptionerID struct {
	IPAddr    string
	NumConn   int
	NetworkID NetworkID
}

// TableID string representation of CaptionerID, colon delimited fields.
func (c CaptionerID) TableID() string {
	return fmt.Sprint(c.IPAddr, ":", c.NumConn, ":", c.NetworkID)
}

func (c CaptionerID) String() string {
	return fmt.Sprint(c.IPAddr, ":", c.NumConn, ":", c.NetworkID)
}

// FormValuesToCaptionerID validates that an CaptionerID can be created
// from the given form values and creates it.
func FormValuesToCaptionerID(values url.Values) (*CaptionerID, error) {
	ipAddress, numConnString, networkIDString :=
		values.Get("ipAddress"),
		values.Get("numConn"),
		values.Get("networkId")

	// validate IPv4 & IPv6 addresses
	/*
		valid, err := isValidIp(ipAddress)
		if err != nil {
			return nil, err
		}

		if !valid {
			return nil, errors.New("Invalid IP address")
		}
	*/

	// sizeof(int); derp
	if len(numConnString) == 0 || len(numConnString) > 10 {
		return nil, errors.New("Invalid Number of Connections")
	}

	// sizeof(int); derp
	if len(networkIDString) == 0 || len(networkIDString) > 10 {
		return nil, errors.New("Invalid Network ID")
	}

	numConn, err := strconv.Atoi(numConnString)
	if err != nil || numConn < 0 {
		return nil, errors.New("Invalid Number of Connections")
	}

	networkID, err := strconv.Atoi(networkIDString)
	if err != nil {
		return nil, errors.New("Invalid Network ID")
	}

	captionerID := CaptionerID{
		IPAddr:    ipAddress,
		NumConn:   numConn,
		NetworkID: NetworkID(networkID),
	}

	return &captionerID, nil
}
