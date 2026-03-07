# QRIS Transaction Performance Optimizer

## Project Overview

Project ini bertujuan untuk menganalisis dan meningkatkan performa sistem transaksi QRIS real-time dengan pendekatan optimasi pada beberapa layer sistem, yaitu:

* Application layer
* Database layer
* Asynchronous processing
* Caching layer

Fokus utama dari proyek ini adalah mengurangi **latency transaksi**, meningkatkan **throughput sistem**, serta menghilangkan bottleneck yang muncul pada arsitektur synchronous.

---

# System Architecture

Setelah dilakukan optimasi, arsitektur sistem menjadi:

Client
↓
API Server (Go)

Inquiry Flow
API → Redis Cache → PostgreSQL

Payment Flow
API → RabbitMQ Queue → Worker → PostgreSQL

Pendekatan ini memisahkan proses **read-heavy** dan **write-heavy** sehingga sistem dapat menangani beban transaksi lebih efisien.

---

# Optimization Stages

## Baseline System (Day 1–6)

Sistem awal menggunakan arsitektur synchronous dimana setiap request langsung menunggu proses database.

Flow:

Client → API → PostgreSQL → Response

Hasil Load Test:

50 Virtual Users
Average Latency ~3.4 ms
Throughput ~14,481 req/sec

100 Virtual Users
Average Latency ~6.2 ms
Throughput ~15,864 req/sec

150 Virtual Users
Average Latency ~9.2 ms
Throughput ~16,163 req/sec

Bottleneck utama ditemukan pada operasi **database write dan query database langsung dari API**.

---

# Async Processing Optimization 

Untuk mengatasi bottleneck write-heavy, sistem diubah menggunakan **RabbitMQ asynchronous processing**.

Flow baru:

Client → API → Queue → Response
Worker → Queue → PostgreSQL

Keuntungan:

* API tidak menunggu database
* Write diproses di background worker
* Latency API menurun drastis

Hasil Load Test:

Average Latency ~1 ms
Throughput ~84,000 req/sec

---

# Redis Caching Optimization =

Endpoint inquiry yang bersifat **read-heavy** dioptimalkan menggunakan Redis caching.

Flow:

Client → API → Redis Cache
↓
PostgreSQL (cache miss)

Monitoring Redis menunjukkan:

keyspace_hits: 519,872
keyspace_misses: 18

Artinya lebih dari **99.9% request berhasil dilayani dari cache**, sehingga query database berkurang secara signifikan.

---

# System Monitoring and Bottleneck Analysis 
Monitoring dilakukan pada tiga komponen utama:

### RabbitMQ

Queue monitoring menunjukkan:

Ready messages ≈ 0
Publish rate ≈ Deliver rate

Hal ini menunjukkan worker mampu memproses message tanpa backlog.

---

### Redis

Redis menunjukkan rasio:

keyspace_hits >> keyspace_misses

Ini menunjukkan caching layer bekerja secara efektif.

---

### PostgreSQL

Monitoring menggunakan:

SELECT * FROM pg_stat_activity;

Hasil menunjukkan mayoritas koneksi berada pada status **idle**, yang menandakan query diproses dengan cepat tanpa penumpukan.

---

# Performance Comparison

| Stage            | Endpoint | Avg Latency | Throughput |
| ---------------- | -------- | ----------- | ---------- |
| Baseline         | Payment  | ~6 ms       | ~15k RPS   |
| Async Processing | Payment  | ~1 ms       | ~84k RPS   |
| Redis Cache      | Inquiry  | ~1.8 ms     | ~52k RPS   |

Optimasi yang dilakukan berhasil meningkatkan performa sistem secara signifikan dengan menurunkan latency dan meningkatkan throughput.

---

# Key Technologies

* Golang (API Server)
* PostgreSQL (Database)
* RabbitMQ (Async Message Queue)
* Redis (Caching Layer)
* Docker (Containerization)
* k6 (Load Testing)

---

# Conclusion

Implementasi asynchronous processing menggunakan RabbitMQ dan caching menggunakan Redis berhasil mengurangi bottleneck pada operasi write-heavy dan read-heavy.

Arsitektur sistem menjadi lebih scalable, mampu menangani throughput tinggi dengan latency rendah, serta mendistribusikan beban kerja secara lebih efisien antara API server, queue system, cache layer, dan database.
