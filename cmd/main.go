package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	db.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Connected to PostgreSQL")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "QRIS Optimizer Running")
	})

	http.HandleFunc("/qris/inquiry", inquiryHandler)
	http.HandleFunc("/qris/payment", paymentHandler)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func inquiryHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	time.Sleep(1 * time.Second) // simulasi legacy latency

	merchantID := r.URL.Query().Get("merchant_id")
	if merchantID == "" {
		http.Error(w, "merchant_id required", http.StatusBadRequest)
		return
	}

	row := db.QueryRow(
		"SELECT id, name, city, status FROM merchants WHERE id=$1",
		merchantID,
	)

	var id int
	var name, city, status string

	err := row.Scan(&id, &name, &city, &status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ID:%d Name:%s City:%s Status:%s\n", id, name, city, status)
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	time.Sleep(1 * time.Second) // simulasi legacy latency

	merchantIDStr := r.URL.Query().Get("merchant_id")
	amountStr := r.URL.Query().Get("amount")

	if merchantIDStr == "" || amountStr == "" {
		http.Error(w, "merchant_id and amount required", http.StatusBadRequest)
		return
	}

	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		http.Error(w, "invalid merchant_id", http.StatusBadRequest)
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	status := "success"

	_, err = db.Exec(
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