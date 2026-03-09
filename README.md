# QRIS Performance Optimizer (Baseline)

Prototype backend system untuk mensimulasikan transaksi QRIS real-time dan menganalisis performa sistem sebelum dilakukan optimasi.

Project ini digunakan untuk mengidentifikasi bottleneck pada sistem transaksi real-time yang terintegrasi dengan sistem legacy dan database, serta melakukan pengujian performa menggunakan load testing.

Versi repository ini merupakan **baseline implementation (synchronous processing)** sebelum dilakukan optimasi seperti caching dan asynchronous processing.

---

# Background

Transaksi QRIS dan sistem pembayaran real-time membutuhkan waktu respon yang cepat dan stabil.

Namun pada implementasi nyata, latency sering meningkat karena beberapa faktor seperti:

* integrasi dengan **sistem legacy**
* query database yang tidak optimal
* **beban transaksi tinggi (peak load)**
* proses yang berjalan secara **synchronous**

Ketika API harus menunggu proses pada sistem legacy dan database selesai, waktu respon akan meningkat saat jumlah transaksi bertambah.

Prototype ini dibuat untuk **mengukur performa sistem sebelum dilakukan optimasi** seperti caching atau asynchronous processing.

---

# System Overview

Baseline system terdiri dari beberapa komponen utama:

### API Service

Backend service yang mensimulasikan endpoint transaksi QRIS.

### PostgreSQL Database

Digunakan untuk menyimpan data merchant dan transaksi.

### Legacy System Simulation

Untuk mensimulasikan integrasi dengan sistem lama, API menambahkan delay menggunakan:

```
time.Sleep(1 * time.Second)
```

Delay ini merepresentasikan waktu respon dari sistem legacy seperti core banking system.

### Load Testing Tool

k6 digunakan untuk melakukan pengujian performa sistem dengan berbagai skenario virtual users.

---

# Technology Stack

### Backend

Go (Golang)

### Database

PostgreSQL

### Load Testing

k6

### Containerization

Docker

---

# API Endpoints

## Inquiry QRIS

Digunakan untuk mengecek informasi merchant.

```
GET /qris/inquiry?merchant_id=1
```

Flow:

```
Client → API → Legacy Delay → Database Query → Response
```

---

## Payment QRIS

Mensimulasikan proses pembayaran QRIS.

```
POST /qris/payment?merchant_id=1&amount=10000
```

Pada baseline system, proses dilakukan secara **synchronous**:

```
Client → API → Legacy Delay → Database Insert → Response
```

API harus menunggu database selesai melakukan insert sebelum mengirim response ke client.

---

# Load Testing

Pengujian performa dilakukan menggunakan **k6** untuk mensimulasikan beban transaksi pada endpoint inquiry dan payment.

Setiap pengujian dijalankan selama **10 detik** dengan jumlah virtual users berbeda untuk melihat dampak concurrency terhadap latency dan throughput.

---

# Load Testing Results

## 100 Virtual Users

Average Latency
~1.0 s

p95 Latency
~1.02 s

Throughput
~98 requests/sec

---

## 300 Virtual Users

Average Latency
~1.0 s

p95 Latency
~1.02 s

Throughput
~295 requests/sec

---

# Catatan

Hasil load testing dapat sedikit berbeda tergantung spesifikasi perangkat dan kondisi sistem saat pengujian.

Namun karena sistem menggunakan **legacy delay ~1 detik**, latency rata-rata akan berada di sekitar **1 detik per request**.

---

# Baseline Characteristics

Pada baseline implementation:

* API melakukan **database query dan insert secara langsung**
* proses bersifat **synchronous**
* terdapat **simulasi latency dari legacy system (~1s)**
* throughput meningkat seiring bertambahnya virtual users
* latency relatif konstan karena dipengaruhi oleh delay legacy

Baseline ini digunakan sebagai **acuan sebelum implementasi optimasi sistem**.

---

# How to Run

## Start database container

```
docker compose up -d
```

## Run API server

```
go run cmd/main.go
```

Server berjalan di:

```
http://localhost:8080
```

---

# Run Load Testing

Jalankan load test menggunakan k6.

Payment test:

```
k6 run k6/payment_test.js
```

Inquiry test:

```
k6 run k6/inquiry_test.js
```

---
