package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {

	connStr := "host=localhost port=5432 user=qris password=qris dbname=qris_db sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database not reachable:", err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)

	fmt.Println("Connected to PostgreSQL")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "QRIS Optimizer Running")
	})

	http.HandleFunc("/qris/inquiry", inquiryHandler)
	http.HandleFunc("/qris/payment", paymentHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

}

func inquiryHandler(w http.ResponseWriter, r *http.Request) {

	// time.Sleep(1 * time.Second)

	rows, err := db.Query("SELECT id, name, city, status FROM merchants")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, city, status string

		rows.Scan(&id, &name, &city, &status)

		fmt.Fprintf(w, "ID:%d Name:%s City:%s Status:%s\n", id, name, city, status)
	}
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	// time.Sleep(1 * time.Second)

	merchantID := 1
	amount := 10000
	status := "success"

	_, err := db.Exec(
		"INSERT INTO transactions (merchant_id, amount, status) VALUES ($1,$2,$3)",
		merchantID,
		amount,
		status,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Payment Success Merchant:%d Amount:%d", merchantID, amount)
}