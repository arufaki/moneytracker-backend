# Money Tracker AI - Backend

Backend REST API untuk aplikasi pencatatan keuangan pribadi berbasis AI. Pengguna dapat mencatat transaksi hanya dengan mengirim pesan teks bebas dalam bahasa Indonesia, kemudian sistem akan mengekstrak informasi transaksi secara otomatis menggunakan Google Gemini AI.

---

## Daftar Isi

- [Tentang Aplikasi](#tentang-aplikasi)
- [Tech Stack](#tech-stack)
- [Library yang Digunakan](#library-yang-digunakan)
- [Struktur Folder dan File](#struktur-folder-dan-file)
- [Arsitektur Aplikasi](#arsitektur-aplikasi)
- [Schema Database](#schema-database)
- [API yang Tersedia](#api-yang-tersedia)
- [Setup Project](#setup-project)
- [Cara Menjalankan Aplikasi](#cara-menjalankan-aplikasi)
- [Cara Menjalankan Test](#cara-menjalankan-test)

---

## Tentang Aplikasi

Money Tracker AI adalah REST API backend yang memungkinkan pengguna mencatat pemasukan dan pengeluaran keuangan melalui pesan teks bahasa alami (natural language). Pesan seperti `"tadi beli mie ayam 15rb dari gopay"` akan diproses oleh Google Gemini AI untuk mengekstrak detail transaksi secara otomatis, kemudian disimpan ke database.

Fitur utama:

- Pencatatan transaksi berbasis percakapan AI (natural language input)
- Manajemen wallet (dompet/rekening)
- Manajemen kategori transaksi dengan batas anggaran (budget limit)
- Peringatan otomatis jika pengeluaran kategori melampaui budget
- Ringkasan dan analitik keuangan bulanan dengan breakdown per kategori
- Auto-migration dan seeding data default saat pertama kali dijalankan

---

## Tech Stack

| Komponen | Teknologi |
|---|---|
| Language | Go 1.26 |
| Web Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| AI / LLM | Google Gemini (via google.golang.org/genai) |
| Environment | godotenv |

---

## Library yang Digunakan

| Library | Fungsi |
|---|---|
| `github.com/gin-gonic/gin` | HTTP web framework untuk routing dan handler |
| `gorm.io/gorm` | ORM untuk interaksi dengan database |
| `gorm.io/driver/postgres` | Driver PostgreSQL untuk GORM |
| `github.com/joho/godotenv` | Memuat konfigurasi dari file .env |
| `google.golang.org/genai` | Client resmi Google Gemini AI |
| `github.com/stretchr/testify` | Assertion library untuk unit testing |
| `github.com/jackc/pgx/v5` | Driver PostgreSQL tingkat rendah (digunakan GORM secara internal) |

---

## Struktur Folder dan File

```
moneytracker/
|
|-- main.go                          # Entry point aplikasi, dependency injection, dan inisialisasi server
|
|-- .env                             # File konfigurasi environment (tidak di-commit ke git)
|-- .env.example                     # Contoh konfigurasi environment
|-- .gitignore                       # Daftar file yang diabaikan git
|-- go.mod                           # Deklarasi module dan dependensi Go
|-- go.sum                           # Checksum dependensi
|
|-- config/
|   `-- database.go                  # Koneksi database, auto-migration, dan seeding data default
|
|-- models/
|   |-- wallet.go                    # Struct model database untuk tabel wallets
|   |-- category.go                  # Struct model database untuk tabel categories
|   |-- transaction.go               # Struct model database untuk tabel transactions
|   |-- ai_log.go                    # Struct model database untuk tabel ai_logs
|   |-- chat.go                      # DTO untuk request dan response endpoint /api/chat
|   |-- ai_dto.go                    # DTO ParsedTransaction hasil ekstraksi AI
|   `-- analytics_dto.go             # DTO MonthlySummary dan CategoryBreakdown untuk analitik
|
|-- repositories/
|   |-- wallet_repository.go         # Interface dan implementasi akses data wallet
|   |-- category_repository.go       # Interface dan implementasi akses data category
|   |-- transaction_repository.go    # Interface dan implementasi akses data transaction
|   `-- ai_log_repository.go         # Interface dan implementasi penyimpanan log AI
|
|-- services/
|   |-- wallet_service.go            # Logika bisnis wallet (buat, ambil)
|   |-- category_service.go          # Logika bisnis category (buat, ambil)
|   |-- transaction_service.go       # Logika bisnis utama: parsing pesan AI, update saldo, cek budget
|   |-- analytics_service.go         # Logika agregasi data untuk ringkasan bulanan
|   `-- ai_service.go                # Integrasi Gemini AI: kirim prompt, parsing JSON response
|
|-- controllers/
|   |-- wallet_controller.go         # HTTP handler untuk endpoint /api/wallets
|   |-- category_controller.go       # HTTP handler untuk endpoint /api/categories
|   |-- chat_controller.go           # HTTP handler untuk endpoint /api/chat
|   `-- analytics_controller.go      # HTTP handler untuk endpoint /api/summary
|
|-- routes/
|   `-- routes.go                    # Pendaftaran semua route API ke Gin router
|
`-- tests/
    |-- setup_test.go                # Setup test: inisialisasi DB test, router, helper request
    |-- ping_test.go                 # Test endpoint /ping (health check)
    |-- wallet_test.go               # Integration test untuk endpoint wallet
    |-- category_test.go             # Integration test untuk endpoint category
    |-- chat_test.go                 # Integration test untuk endpoint chat (dengan mock AI)
    |-- analytics_test.go            # Integration test untuk endpoint summary
    `-- mocks/
        `-- ai_service_mock.go       # Mock implementation dari AIService untuk testing
```

### Konvensi Penamaan File

- Nama file menggunakan format `<domain>_<layer>.go`, contoh: `wallet_service.go`, `category_repository.go`
- File model database menggunakan nama entitas tunggal, contoh: `wallet.go`, `category.go`
- File DTO menggunakan suffix `_dto.go`, contoh: `ai_dto.go`, `analytics_dto.go`
- File test menggunakan suffix `_test.go` sesuai standar Go

---

## Arsitektur Aplikasi

Aplikasi mengikuti pola arsitektur berlapis (Layered Architecture) dengan pemisahan tanggung jawab yang jelas:

```
HTTP Request
    |
    v
Controller        <- Menerima request, validasi input, mengembalikan response HTTP
    |
    v
Service           <- Logika bisnis, orkestrasi antar repository, integrasi AI
    |
    v
Repository        <- Akses data ke database (abstraksi query GORM)
    |
    v
Database (PostgreSQL)
```

Dependency Injection dilakukan secara manual di `main.go`. Setiap layer hanya bergantung pada interface dari layer di bawahnya, bukan implementasi konkret. Hal ini memudahkan penggantian implementasi dan penulisan unit test menggunakan mock.

**Alur Pencatatan Transaksi via Chat:**

1. Client mengirim `POST /api/chat` dengan pesan teks bebas
2. `ChatController` meneruskan pesan ke `TransactionService`
3. `TransactionService` memanggil `AIService` untuk mem-parsing pesan menggunakan Gemini AI
4. AI mengembalikan data terstruktur: nominal, tipe, kategori, wallet, dan deskripsi
5. Service mencari atau membuat wallet dan kategori yang relevan di database
6. Transaksi disimpan dan saldo wallet diperbarui dalam satu database transaction (atomic)
7. Jika ada budget limit yang terlampaui, pesan peringatan ditambahkan ke response

---

## Schema Database

### Tabel `wallets`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik wallet |
| `name` | VARCHAR(100) UNIQUE NOT NULL | Nama wallet (contoh: GoPay, BCA) |
| `balance` | NUMERIC(15,2) DEFAULT 0 | Saldo wallet saat ini |
| `created_at` | TIMESTAMP | Waktu data dibuat |
| `updated_at` | TIMESTAMP | Waktu data terakhir diperbarui |

### Tabel `categories`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik kategori |
| `name` | VARCHAR(100) UNIQUE NOT NULL | Nama kategori (contoh: Makanan, Gaji) |
| `type` | VARCHAR(20) NOT NULL | Tipe kategori: `income` atau `expense` |
| `budget_limit` | NUMERIC(15,2) DEFAULT 0 | Batas anggaran bulanan (0 berarti tidak ada batas) |
| `created_at` | TIMESTAMP | Waktu data dibuat |
| `updated_at` | TIMESTAMP | Waktu data terakhir diperbarui |

### Tabel `transactions`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik transaksi |
| `wallet_id` | INTEGER NOT NULL | Foreign key ke tabel `wallets` |
| `category_id` | INTEGER NOT NULL | Foreign key ke tabel `categories` |
| `amount` | NUMERIC(15,2) NOT NULL | Nominal transaksi |
| `type` | VARCHAR(20) NOT NULL | Tipe transaksi: `income` atau `expense` |
| `description` | TEXT | Deskripsi singkat transaksi |
| `created_at` | TIMESTAMP | Waktu transaksi dibuat |

Relasi: `transactions.wallet_id` merujuk ke `wallets.id` dengan constraint `ON DELETE CASCADE`. `transactions.category_id` merujuk ke `categories.id` dengan constraint `ON DELETE CASCADE`.

### Tabel `ai_logs`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik log |
| `raw_message` | TEXT NOT NULL | Pesan asli yang dikirim pengguna |
| `extracted_json` | JSONB NOT NULL | JSON hasil ekstraksi dari Gemini AI |
| `created_at` | TIMESTAMP | Waktu log dibuat |

### Data Seed Default

Saat pertama kali dijalankan dan tabel masih kosong, sistem akan otomatis mengisi data berikut:

**Kategori default:**
- Makanan (expense)
- Transportasi (expense)
- Gaji (income)
- Hiburan (expense)
- Belanja (expense)
- Tagihan (expense)

**Wallet default:**
- Cash (saldo 0)

---

## API yang Tersedia

Base URL: `http://localhost:8080`

### Health Check

**GET /ping**

Memeriksa status server dan koneksi database.

Response:
```json
{
  "status": "ok",
  "db_status": "connected"
}
```

---

### Wallets

**GET /api/wallets**

Mengambil semua wallet yang tersedia.

Response `200 OK`:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Cash",
      "balance": 500000,
      "created_at": "2026-09-01T10:00:00Z",
      "updated_at": "2026-09-30T15:00:00Z"
    }
  ]
}
```

---

**GET /api/wallets/:id**

Mengambil detail satu wallet berdasarkan ID.

| Parameter | Tipe | Keterangan |
|---|---|---|
| `id` | integer (path) | ID wallet |

Response `200 OK`:
```json
{
  "data": {
    "id": 1,
    "name": "Cash",
    "balance": 500000,
    "created_at": "2026-09-01T10:00:00Z",
    "updated_at": "2026-09-30T15:00:00Z"
  }
}
```

Response `404 Not Found`:
```json
{ "error": "wallet not found" }
```

---

**POST /api/wallets**

Membuat wallet baru.

Request body:
```json
{
  "name": "GoPay",
  "balance": 100000
}
```

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `name` | string | Ya | Nama wallet, harus unik |
| `balance` | number | Tidak | Saldo awal, default 0 |

Response `201 Created`:
```json
{
  "data": {
    "id": 2,
    "name": "GoPay",
    "balance": 100000,
    "created_at": "2026-09-30T19:00:00Z",
    "updated_at": "2026-09-30T19:00:00Z"
  },
  "message": "Wallet created successfully"
}
```

---

### Categories

**GET /api/categories**

Mengambil semua kategori yang tersedia.

Response `200 OK`:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Makanan",
      "type": "expense",
      "budget_limit": 500000,
      "created_at": "2026-09-01T10:00:00Z",
      "updated_at": "2026-09-01T10:00:00Z"
    }
  ]
}
```

---

**GET /api/categories/:id**

Mengambil detail satu kategori berdasarkan ID.

| Parameter | Tipe | Keterangan |
|---|---|---|
| `id` | integer (path) | ID kategori |

Response `200 OK`:
```json
{
  "data": {
    "id": 1,
    "name": "Makanan",
    "type": "expense",
    "budget_limit": 500000,
    "created_at": "2026-09-01T10:00:00Z",
    "updated_at": "2026-09-01T10:00:00Z"
  }
}
```

Response `404 Not Found`:
```json
{ "error": "category not found" }
```

---

**POST /api/categories**

Membuat kategori baru.

Request body:
```json
{
  "name": "Investasi",
  "type": "income"
}
```

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `name` | string | Ya | Nama kategori, harus unik |
| `type` | string | Ya | Nilai: `income` atau `expense` |

Response `201 Created`:
```json
{
  "data": {
    "id": 7,
    "name": "Investasi",
    "type": "income",
    "budget_limit": 0,
    "created_at": "2026-09-30T19:00:00Z",
    "updated_at": "2026-09-30T19:00:00Z"
  },
  "message": "Category created successfully"
}
```

---

### Chat (Pencatatan Transaksi via AI)

**POST /api/chat**

Memproses pesan teks bebas dari pengguna dan mencatat transaksi secara otomatis menggunakan AI.

Request body:
```json
{
  "message": "tadi beli mie ayam 15rb dari gopay"
}
```

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `message` | string | Ya | Pesan teks bebas yang mendeskripsikan transaksi |

Response `200 OK`:
```json
{
  "success": true,
  "message": "Berhasil mencatat pengeluaran Makanan sebesar Rp 15000 dari GoPay. Sisa saldo GoPay kamu sekarang Rp 85000.",
  "transaction": {
    "id": 10,
    "wallet_id": 2,
    "wallet": { "id": 2, "name": "GoPay", "balance": 85000 },
    "category_id": 1,
    "category": { "id": 1, "name": "Makanan", "type": "expense" },
    "amount": 15000,
    "type": "expense",
    "description": "mie ayam",
    "created_at": "2026-09-30T19:00:00Z"
  },
  "updated_wallet": {
    "id": 2,
    "name": "GoPay",
    "balance": 85000
  },
  "parsed_data": {
    "amount": 15000,
    "type": "expense",
    "category": "Makanan",
    "wallet": "GoPay",
    "description": "mie ayam"
  }
}
```

Response `400 Bad Request` (saldo tidak cukup):
```json
{
  "success": false,
  "error": "insufficient balance"
}
```

Catatan: Jika wallet atau kategori yang disebutkan AI belum ada di database, sistem akan membuatnya secara otomatis.

---

### Analytics

**GET /api/summary**

Mengambil ringkasan keuangan bulanan, termasuk total pemasukan, pengeluaran, saldo bersih seluruh wallet, dan breakdown pengeluaran per kategori.

| Query Parameter | Tipe | Keterangan |
|---|---|---|
| `month` | integer | Bulan (1-12). Default: bulan berjalan |
| `year` | integer | Tahun. Default: tahun berjalan |

Contoh: `GET /api/summary?month=9&year=2026`

Response `200 OK`:
```json
{
  "month": 9,
  "year": 2026,
  "income": 5000000,
  "expense": 1500000,
  "net_balance": 3500000,
  "breakdown": [
    {
      "category_name": "Makanan",
      "total": 800000,
      "percentage": 53.33,
      "budget_limit": 500000,
      "is_over_budget": true
    },
    {
      "category_name": "Transportasi",
      "total": 700000,
      "percentage": 46.67,
      "budget_limit": 0,
      "is_over_budget": false
    }
  ]
}
```

---

## Setup Project

### Prasyarat

- Go 1.21 atau lebih baru
- PostgreSQL 13 atau lebih baru
- Google Gemini API Key (dapatkan di https://aistudio.google.com/)

### Langkah-langkah

1. Clone repository

```bash
git clone https://github.com/arufaki/moneytracker-backend.git
cd moneytracker-backend
```

2. Install dependensi

```bash
go mod download
```

3. Buat database PostgreSQL

```sql
CREATE DATABASE money_tracker;
```

4. Salin file konfigurasi

```bash
cp .env.example .env
```

5. Isi konfigurasi di file `.env`

```env
PORT=8080
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=money_tracker
DB_PORT=5432
GEMINI_API_KEY=your_gemini_api_key_here
```

Migrasi tabel dan seeding data default akan dilakukan secara otomatis saat server pertama kali dijalankan.

---

## Cara Menjalankan Aplikasi

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080` (atau sesuai nilai `PORT` di `.env`).

Output yang diharapkan saat server berhasil dijalankan:

```
Database connection successfully opened
Database migrated successfully
AI Service initialized
[GIN-debug] Listening and serving HTTP on :8080
```

---

## Cara Menjalankan Test

Test berada di folder `tests/` dan merupakan integration test yang berjalan terhadap database nyata. Pastikan konfigurasi database di `.env` sudah benar sebelum menjalankan test.

Menjalankan semua test:

```bash
go test ./tests/...
```

Menjalankan test dengan output detail:

```bash
go test ./tests/... -v
```

Menjalankan test untuk domain tertentu:

```bash
go test ./tests/... -v -run TestWallet
go test ./tests/... -v -run TestCategory
go test ./tests/... -v -run TestChat
go test ./tests/... -v -run TestAnalytics
go test ./tests/... -v -run TestPing
```

### Catatan tentang Test

- Test menggunakan database yang sama dengan konfigurasi di `.env`, bukan database terpisah. Setiap test case akan membersihkan (TRUNCATE) tabel yang relevan sebelum dijalankan.
- AI service digantikan dengan `MockAIService` agar test tidak bergantung pada koneksi Gemini API dan memberikan hasil yang deterministik.
- `setup_test.go` berisi helper `SetupTestRouter()`, `CleanDatabase()`, dan `DoRequest()` yang digunakan bersama oleh semua file test.
