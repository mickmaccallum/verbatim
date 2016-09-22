package dashboard

import (
	"database/sql"
	"log"
)

type network struct {
	id   sql.NullInt64
	name sql.NullString
}

func getNetworks() {
	rows, err := db.Query("select id, name from network;")

	if err != nil {
		log.Println(err.Error())
		return
	}

	cols, _ := rows.Columns()

	log.Println(cols)
	log.Println("++++++++++++++++++++++++++++++++++")
	for rows.Next() {
		var net network

		if err = rows.Scan(&net.id, &net.name); err != nil {
			log.Fatal(err)
		}

		log.Println(net)
	}
}
