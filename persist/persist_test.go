package persist

import (
	"database/sql"
	"github.com/0x7fffffff/verbatim/model"
	"testing"
)

// 2.2.2.6
// The RSS shall store information about admins, TV networks, and TV encoders in a local SQLite database.

// 2.2.1.5
// The RSS shall retrieve information about admins, TV networks, and TV encoders from a local SQLite database.
// func TestStorage(*testing.T) {

func TestGetAdmins(t *testing.T) {
	admins, err := GetAdmins()

	if err == nil {
		if len(admins) == 0 {
			t.Log("Got 0 admins, expected 2")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetAdminByID(t *testing.T) {
	_, err := GetAdminForID(1)

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetAdminByCredentials(t *testing.T) {
	_, err := GetAdminForCredentials("mick2", "abc123")

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetNetworks(t *testing.T) {
	networks, err := GetNetworks()

	if err == nil {
		if len(networks) == 0 {
			t.Log("Got 0 networks, expected 2")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetNetworkByID(t *testing.T) {
	_, err := GetNetwork(1)

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}

// AddNetwork, UpdateNetwork, DeleteNetwork

func TestGetEncoders(t *testing.T) {
	encoders, err := GetEncoders()

	if err == nil {
		if len(encoders) == 0 {
			t.Log("Got 0 encoders, expected 3")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetEncoder(t *testing.T) {
	encoder, err := GetEncoder(1)

	if err == nil {
		if encoder == nil {
			t.Log("Encoder was unexpectedly nil")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestGetEncodersForNetwork(t *testing.T) {
	network, _ := GetNetwork(1)
	encoders, err := GetEncodersForNetwork(*network)

	if err == nil {
		if len(encoders) == 0 {
			t.Log("Got 0 encoders, expected 2")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

// AddEncoder, UpdateEncoder, DeleteEncoder
func TestAddEncoder(t *testing.T) {
	network, _ := GetNetwork(1)
	encoder := model.Encoder{
		IPAddress: "19.34.76.34",
		Port:      3456,
		Name:      sql.NullString{String: "cool encoder", Valid: true},
		Handle:    "username",
		Password:  "password1",
	}

	newEncoder, err := AddEncoder(encoder, *network)

	if err == nil {
		enc := *newEncoder

		if enc.Port != 3456 {
			t.Log("failed to properly save encoder port")
			t.Fail()
		}

		if enc.IPAddress != "19.34.76.34" {
			t.Log("failed to properly save encoder ip address")
			t.Fail()
		}

		if enc.Name.Valid && enc.Name.String != "cool encoder" {
			t.Log("failed to properly save encoder ")
			t.Fail()
		}

		if enc.Handle != "username" {
			t.Log("failed to properly save encoder handle")
			t.Fail()
		}

		if enc.Password != "password1" {
			t.Log("failed to properly save encoder password")
			t.Fail()
		}
	} else {
		t.Error(err.Error())
		t.Fail()
	}
}

func TestUpdateEncoder(t *testing.T) {

}

func TestDeleteEncoder(t *testing.T) {
	network, _ := GetNetwork(1)
	encoder := model.Encoder{
		IPAddress: "19.34.76.34",
		Port:      3456,
		Name:      sql.NullString{String: "cool encoder", Valid: true},
		Handle:    "username",
		Password:  "password1",
	}

	newEncoder, _ := AddEncoder(encoder, *network)
	err := DeleteEncoder(*newEncoder)

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}

// TODO: Test EncoderToJSON & NetworkToJSON
