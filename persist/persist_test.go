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

	t.Log(admins)
}

func TestGetAdminByID(t *testing.T) {
	admin, err := GetAdminForID(1)

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(admin)
}

func TestGetAdminByCredentials(t *testing.T) {
	admin, err := GetAdminForCredentials("0x7fs", "1234567890")

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(admin)
}

// AddNetwork, UpdateNetwork, DeleteNetwork
func TestAddNetwork(t *testing.T) {
	newNetwork := model.Network{
		Name:          "MSNBC",
		ListeningPort: 4040,
	}

	network, err := AddNetwork(newNetwork)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(network)
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

	t.Log(networks)
}

func TestGetNetworkByID(t *testing.T) {
	network, err := GetNetwork(1)

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(network)
}

func TestUpdateNetwork(t *testing.T) {
	network, _ := GetNetwork(1)
	network.ListeningPort = 6000

	err := UpdateNetwork(*network)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(network)
}

// AddEncoder, UpdateEncoder, DeleteEncoder
func TestAddEncoder(t *testing.T) {
	network, _ := GetNetwork(1)
	t.Log("Adding encoder for network: ")
	t.Log(network.ID)
	encoder := model.Encoder{
		IPAddress: "19.34.76.34",
		Port:      3456,
		Name:      sql.NullString{String: "my encoder", Valid: true},
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

		if enc.Name.Valid && enc.Name.String != "my encoder" {
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

	t.Log(newEncoder)
}

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

	t.Log(encoders)
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

	t.Log(encoder)
}

func TestUpdateEncoder(t *testing.T) {
	encoder, _ := GetEncoder(1)
	encoder.IPAddress = "127.21.33.134"

	err := UpdateEncoder(*encoder)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(err)
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

func TestDeleteNetwork(t *testing.T) {
	network, _ := GetNetwork(1)
	err := DeleteNetwork(*network)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	t.Log(network)
}

// TODO: Test EncoderToJSON & NetworkToJSON
