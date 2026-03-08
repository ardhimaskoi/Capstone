package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
	_ "github.com/lib/pq"
)

var db *sql.DB
var mqChannel *amqp.Channel
var redisClient *redis.Client
var ctx = context.Background()

// async publish queue
var publishChan = make(chan []byte, 10000)

type Payment struct {
	MerchantID int    `json:"merchant_id"`
	Amount     int    `json:"amount"`
	Status     string `json:"status"`
}

func main() {

	connStr := "host=localhost port=5432 user=qris password=qris dbname=qris_db sslmode=disable"

	var err error

	// PostgreSQL
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database not reachable:", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)

	fmt.Println("Connected to PostgreSQL")

	// Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis connection error:", err)
	}

	fmt.Println("Connected to Redis")

	// RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("RabbitMQ connection error:", err)
	}

	mqChannel, err = conn.Channel()
	if err != nil {
		log.Fatal("Channel error:", err)
	}

	_, err = mqChannel.QueueDeclare(
		"payment_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("Queue declare error:", err)
	}

	startPublisher()
	startWorker()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "QRIS Optimizer Running")
	})

	http.HandleFunc("/qris/inquiry", inquiryHandler)
	http.HandleFunc("/qris/payment", paymentHandler)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)

}

func inquiryHandler(w http.ResponseWriter, r *http.Request) {

	cacheKey := "merchant_list"

	cached, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		fmt.Fprintln(w, cached)
		return
	}

	rows, err := db.Query("SELECT id, name, city, status FROM merchants")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := ""

	for rows.Next() {

		var id int
		var name, city, status string

		err := rows.Scan(&id, &name, &city, &status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		line := fmt.Sprintf(
			"ID:%d Name:%s City:%s Status:%s\n",
			id, name, city, status,
		)

		result += line
	}

	// cache result selama 60 detik
	redisClient.Set(ctx, cacheKey, result, 60*time.Second)

	fmt.Fprintln(w, result)

}

func paymentHandler(w http.ResponseWriter, r *http.Request) {

	payment := Payment{
		MerchantID: 1,
		Amount:     10000,
		Status:     "success",
	}

	body, _ := json.Marshal(payment)

	// async publish ke queue
	publishChan <- body

	fmt.Fprintf(w, "Payment queued")

}

func startPublisher() {

	go func() {

		for body := range publishChan {

			err := mqChannel.Publish(
				"",
				"payment_queue",
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			)

			if err != nil {
				log.Println("Publish error:", err)
			}

		}

	}()

}

// Worker pool untuk memproses payment queue secara asynchronous
func startWorker() {

	workerCount := 8
	batchSize := 20

	msgs, err := mqChannel.Consume(
		"payment_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal("Consumer error:", err)
	}

	for i := 0; i < workerCount; i++ {

		go func() {

			payments := make([]Payment, 0, batchSize)

			for d := range msgs {

				var payment Payment

				err := json.Unmarshal(d.Body, &payment)
				if err != nil {
					continue
				}

				payments = append(payments, payment)

				if len(payments) >= batchSize {

					var values []string
					var args []interface{}

					for i, p := range payments {

						values = append(values,
							fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3),
						)

						args = append(args,
							p.MerchantID,
							p.Amount,
							p.Status,
						)

					}

					query := fmt.Sprintf(
						"INSERT INTO transactions (merchant_id, amount, status) VALUES %s",
						strings.Join(values, ","),
					)

					_, err := db.Exec(query, args...)
					if err != nil {
						log.Println("DB insert error:", err)
					}

					payments = payments[:0]

				}

			}

		}()

	}

}