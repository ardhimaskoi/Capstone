package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
	_ "github.com/lib/pq"
)

var db *sql.DB
var mqChannel *amqp.Channel
var redisClient *redis.Client
var ctx = context.Background()

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

		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		_, err = redisClient.Ping(ctx).Result()
		if err != nil {
			log.Fatal("Redis connection error:", err)
		}

		fmt.Println("Connected to Redis")

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

	http.HandleFunc("/qris/inquiry", inquiryHandler)
	http.HandleFunc("/qris/payment", paymentHandler)
	startWorker()

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}

func inquiryHandler(w http.ResponseWriter, r *http.Request) {

    cacheKey := "merchant_list"

    // cek redis dulu
    cached, err := redisClient.Get(ctx, cacheKey).Result()
    if err == nil {
        fmt.Fprintln(w, cached)
        return
    }

    // cache miss → query DB
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

        line := fmt.Sprintf("ID: %d | Name: %s | City: %s | Status: %s\n",
            id, name, city, status)

        result += line
    }

    // simpan ke redis selama 60 detik
    redisClient.Set(ctx, cacheKey, result, time.Minute)

    fmt.Fprintln(w, result)
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {

    payment := map[string]interface{}{
        "merchant_id": 1,
        "amount":      10000,
        "status":      "success",
    }

    body, _ := json.Marshal(payment)

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
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Payment Queued Successfully")
}

func startWorker() {
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

    go func() {
        for d := range msgs {
            var payment map[string]interface{}
            json.Unmarshal(d.Body, &payment)

			merchantID := int(payment["merchant_id"].(float64))
			amount := int(payment["amount"].(float64))
			status := payment["status"].(string)

			_, err := db.Exec(
			"INSERT INTO transactions (merchant_id, amount, status) VALUES ($1, $2, $3)",
			merchantID,
			amount,
			status,
		)

			if err != nil {
				log.Println("DB insert error:", err)
			}
        }
    }()
}