package model
// Encoder represents a single downstream encoder for a given network
type Encoder struct {
	ID        int
	IPAddress string
	Port      int
	Name      sql.NullString
	Handle    string
	Password  string
	NetworkID int
}
