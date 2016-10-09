package model

// Admin represents a downstream network
type Admin struct {
	ID             int
	Handle         string
	HashedPassword string
}
