# QRIS Performance Optimizer (Baseline)

Prototype backend system untuk mensimulasikan transaksi QRIS real-time dan menganalisis performa sistem sebelum dilakukan optimasi.

Project ini digunakan untuk mengidentifikasi bottleneck pada sistem transaksi real-time yang terintegrasi dengan database dan melakukan pengujian performa menggunakan load testing.

Versi repository ini merupakan baseline implementation (synchronous processing) sebelum dilakukan optimasi lebih lanjut.

## Background

Transaksi QRIS dan sistem pembayaran real-time membutuhkan waktu respon yang cepat dan stabil.

Namun pada implementasi nyata, latency sering meningkat karena beberapa faktor seperti:

- integrasi dengan sistem legacy
- query database yang tidak optimal
- beban transaksi tinggi (peak load)
- proses yang berjalan secara synchronous

Ketika API harus menunggu database menyelesaikan prosesnya, waktu respon akan meningkat saat jumlah transaksi bertambah.

Prototype ini dibuat untuk mengukur performa sistem sebelum dilakukan optimasi seperti caching atau asynchronous processing.

## System Overview

Baseline system terdiri dari beberapa komponen utama:

API Service  
Backend service yang mensimulasikan endpoint transaksi QRIS.

PostgreSQL Database  
Digunakan untuk menyimpan data merchant dan transaksi.

Load Testing Tool  
k6 digunakan untuk melakukan pengujian performa sistem dengan berbagai skenario virtual users.

## Technology Stack

Backend  
Go (Golang)

Database  
PostgreSQL

Load Testing  
k6

Containerization  
Docker


## API Endpoints

### Inquiry QRIS

Digunakan untuk mengecek merchant yang tersedia.

GET /qris/inquiry

Endpoint ini mengambil data merchant dari database.

### Payment QRIS

Mensimulasikan proses pembayaran QRIS.

POST /qris/payment

Pada baseline system, proses ini dilakukan secara synchronous:

Client → API → Database Insert → Response

API harus menunggu database selesai melakukan insert sebelum mengirim response ke client.


## Load Testing

Pengujian performa dilakukan menggunakan k6 untuk mensimulasikan beban transaksi pada endpoint payment.  
Setiap pengujian dijalankan selama 10 detik dengan jumlah virtual users yang berbeda untuk melihat dampak concurrency terhadap latency dan throughput sistem.

1. 50 Virtual Users

Average Latency  
~3.4 ms

Throughput  
~14,481 requests/sec

2. 100 Virtual Users

Average Latency  
~6.2 ms

Throughput  
~15,864 requests/sec

3. 150 Virtual Users

Average Latency  
~9.2 ms

Throughput  
~16,163 requests/sec

### Catatan
Hasil load testing dapat sedikit berbeda tergantung spesifikasi perangkat dan kondisi sistem saat pengujian. Namun pola peningkatan latency seiring bertambahnya jumlah virtual users tetap konsisten.

## Baseline Characteristics

Pada baseline implementation:

- API melakukan database insert secara langsung
- proses bersifat synchronous
- latency meningkat saat concurrency meningkat
- database menjadi bottleneck utama saat load tinggi

Baseline ini digunakan sebagai **acuan sebelum implementasi optimasi sistem**.

## How to Run

1. Start database container

docker compose up -d

2. Run API server

go run cmd/main.go


3. Server berjalan di

http://localhost:8080

## Run Load Testing

Jalankan load test menggunakan k6

k6 run k6/payment_test.js

