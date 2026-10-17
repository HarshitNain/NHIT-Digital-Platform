package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	var id int64
	var gross, tds, net, roundOff float64
	var noteNo string
	err = db.QueryRow("SELECT id, note_no, gross_amount, tds_amount, net_payable_amount, net_payable_round_off FROM payment_notes ORDER BY created_at DESC LIMIT 1").Scan(&id, &noteNo, &gross, &tds, &net, &roundOff)
	if err != nil {
		log.Fatalf("failed to query payment note: %v", err)
	}

	fmt.Printf("Payment Note ID: %d\n", id)
	fmt.Printf("Note No: %s\n", noteNo)
	fmt.Printf("Gross Amount: %.2f\n", gross)
	fmt.Printf("TDS Amount: %.2f\n", tds)
	fmt.Printf("Net Payable Amount: %.2f\n", net)
	fmt.Printf("Round Off: %.2f\n", roundOff)
}
