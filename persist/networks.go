package persist

// AddEncoder adds a new encoder for the given network
func AddEncoder(encoder Encoder, network Network) error {
	query := `
    INSERT INTO encoder (
      ip_address, port, status, network_id
    ) VALUES (
      ?, ?, ?, ?
    );
  `

	_, err := db.Exec(query, encoder.IPAddress, encoder.Port, encoder.Status, network.ID)
	return err
}
