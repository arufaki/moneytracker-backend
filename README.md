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

- Autentikasi pengguna dengan JWT (access token + refresh token)
- Registrasi dengan verifikasi email
- Login via Google OAuth 2.0
- Pencatatan transaksi berbasis percakapan AI (natural language input)
- Manajemen wallet (dompet/rekening)
- Manajemen kategori transaksi dengan batas anggaran (budget limit)
- Peringatan otomatis jika pengeluaran kategori melampaui budget
- Ringkasan dan analitik keuangan bulanan dengan breakdown per kategori
- Rate limiter pada endpoint chat (10 request per menit per IP)
- Pembatasan ukuran body request pada endpoint chat (maks 4KB)
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
| `github.com/golang-jwt/jwt/v5` | Pembuatan dan validasi JWT |
| `golang.org/x/crypto` | Hashing password dengan bcrypt |
| `golang.org/x/oauth2` | Implementasi Google OAuth 2.0 |
| `github.com/wneessen/go-mail` | Pengiriman email verifikasi via SMTP |
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
|   |-- user.go                      # Struct model database untuk tabel users
|   |-- email_verification.go        # Struct model database untuk tabel email_verifications
|   |-- refresh_token.go             # Struct model database untuk tabel refresh_tokens
|   |-- wallet.go                    # Struct model database untuk tabel wallets
|   |-- category.go                  # Struct model database untuk tabel categories
|   |-- transaction.go               # Struct model database untuk tabel transactions
|   |-- ai_log.go                    # Struct model database untuk tabel ai_logs
|   |-- chat.go                      # DTO untuk request dan response endpoint /api/chat
|   |-- ai_dto.go                    # DTO ParsedTransaction hasil ekstraksi AI
|   `-- analytics_dto.go             # DTO MonthlySummary dan CategoryBreakdown untuk analitik
|
|-- repositories/
|   |-- user_repository.go           # Interface dan implementasi akses data user
|   |-- email_verification_repository.go  # Interface dan implementasi akses data token verifikasi email
|   |-- refresh_token_repository.go  # Interface dan implementasi akses data refresh token
|   |-- wallet_repository.go         # Interface dan implementasi akses data wallet
|   |-- category_repository.go       # Interface dan implementasi akses data category
|   |-- transaction_repository.go    # Interface dan implementasi akses data transaction
|   `-- ai_log_repository.go         # Interface dan implementasi penyimpanan log AI
|
|-- services/
|   |-- auth_service.go              # Logika autentikasi: register, login, JWT, refresh token, logout
|   |-- email_service.go             # Pengiriman email verifikasi via SMTP
|   |-- oauth_service.go             # Integrasi Google OAuth 2.0
|   |-- wallet_service.go            # Logika bisnis wallet (buat, ambil)
|   |-- category_service.go          # Logika bisnis category (buat, ambil, hapus)
|   |-- transaction_service.go       # Logika bisnis utama: parsing pesan AI, update saldo, cek budget
|   |-- analytics_service.go         # Logika agregasi data untuk ringkasan bulanan
|   `-- ai_service.go                # Integrasi Gemini AI: kirim prompt, parsing JSON response
|
|-- controllers/
|   |-- root_controller.go           # HTTP handler untuk endpoint / (dokumentasi API)
|   |-- auth_controller.go           # HTTP handler untuk endpoint /auth/*
|   |-- wallet_controller.go         # HTTP handler untuk endpoint /api/wallets
|   |-- category_controller.go       # HTTP handler untuk endpoint /api/categories
|   |-- chat_controller.go           # HTTP handler untuk endpoint /api/chat
|   `-- analytics_controller.go      # HTTP handler untuk endpoint /api/summary
|
|-- middleware/
|   |-- jwt_auth.go                  # Middleware validasi JWT untuk route yang dilindungi
|   `-- security.go                  # Middleware rate limiter per IP dan pembatasan ukuran body
|
|-- routes/
|   `-- routes.go                    # Pendaftaran semua route API ke Gin router
|
`-- tests/
    |-- setup_test.go                # Setup test: inisialisasi DB test, router, helper request
    |-- ping_test.go                 # Test endpoint /ping (health check)
    |-- root_test.go                 # Test endpoint / (dokumentasi API)
    |-- auth_test.go                 # Integration test untuk endpoint autentikasi
    |-- security_test.go             # Test rate limiter dan pembatasan body request
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
Middleware        <- Validasi JWT, rate limit, pembatasan body size
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

**Alur Autentikasi:**

1. Client mengirim `POST /auth/register` dengan email, nama, dan password
2. Sistem membuat user, membuat token verifikasi, dan mengirim email
3. Client mengklik link verifikasi di email (`GET /auth/verify?token=...`)
4. Client mengirim `POST /auth/login` dan mendapatkan access token (JWT) dan refresh token
5. Client menyertakan access token di header `Authorization: Bearer <token>` pada setiap request ke `/api/*`
6. Saat access token kedaluwarsa, client mengirim `POST /auth/refresh` dengan refresh token untuk mendapatkan token baru (token rotation diterapkan)
7. Client mengirim `POST /auth/logout` dengan refresh token untuk menghapus sesi

**Alur Pencatatan Transaksi via Chat:**

1. Client mengirim `POST /api/chat` dengan pesan teks bebas (disertai JWT)
2. `ChatController` meneruskan pesan ke `TransactionService`
3. `TransactionService` memanggil `AIService` untuk mem-parsing pesan menggunakan Gemini AI
4. AI mengembalikan data terstruktur: nominal, tipe, kategori, wallet, dan deskripsi
5. Service mencari atau membuat wallet dan kategori yang relevan di database
6. Transaksi disimpan dan saldo wallet diperbarui dalam satu database transaction (atomic)
7. Jika ada budget limit yang terlampaui, pesan peringatan ditambahkan ke response

---

## Schema Database

### Tabel `users`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik user |
| `email` | VARCHAR(191) UNIQUE NOT NULL | Alamat email pengguna |
| `name` | VARCHAR(100) NOT NULL | Nama pengguna |
| `password_hash` | VARCHAR(255) | Hash bcrypt password (NULL untuk akun OAuth) |
| `google_id` | VARCHAR(191) UNIQUE | ID Google dari OAuth (NULL untuk akun email) |
| `is_verified` | BOOLEAN DEFAULT false | Status verifikasi email |
| `created_at` | TIMESTAMP | Waktu data dibuat |
| `updated_at` | TIMESTAMP | Waktu data terakhir diperbarui |

### Tabel `email_verifications`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik record |
| `user_id` | INTEGER NOT NULL | Foreign key ke tabel `users` |
| `token` | VARCHAR(255) NOT NULL | Token verifikasi (hex string 32 bytes) |
| `expires_at` | TIMESTAMP | Waktu token kedaluwarsa (24 jam sejak dibuat) |

### Tabel `refresh_tokens`

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | SERIAL PRIMARY KEY | ID unik record |
| `user_id` | INTEGER NOT NULL | Foreign key ke tabel `users` |
| `token` | VARCHAR(255) UNIQUE NOT NULL | Hash SHA-256 dari refresh token |
| `expires_at` | TIMESTAMP | Waktu token kedaluwarsa (default 7 hari) |

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

Semua endpoint di bawah `/api/*` dan `POST /auth/logout` memerlukan header autentikasi:

```
Authorization: Bearer <access_token>
```

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

### Dokumentasi API

**GET /**

Mengembalikan daftar semua endpoint beserta contoh request dan response dalam format JSON.

---

### Auth

**POST /auth/register**

Mendaftarkan akun baru. Sistem mengirim email verifikasi ke alamat yang didaftarkan.

Request body:
```json
{
  "email": "user@example.com",
  "name": "Budi Santoso",
  "password": "rahasia123"
}
```

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `email` | string | Ya | Alamat email yang valid |
| `name` | string | Ya | Nama pengguna |
| `password` | string | Ya | Password minimal 6 karakter |

Response `201 Created`:
```json
{
  "message": "Registration successful. Please check your email for verification link."
}
```

---

**GET /auth/verify?token=...**

Memverifikasi email pengguna menggunakan token yang dikirim via email.

| Query Parameter | Tipe | Keterangan |
|---|---|---|
| `token` | string | Token verifikasi dari email |

Response `200 OK`:
```json
{
  "message": "Email verified successfully. You can now login."
}
```

---

**POST /auth/login**

Login dengan email dan password. Mengembalikan access token dan refresh token.

Request body:
```json
{
  "email": "user@example.com",
  "password": "rahasia123"
}
```

Response `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a3f8c2d1e4b7..."
}
```

Response `401 Unauthorized`:
```json
{ "error": "invalid email or password" }
```

---

**GET /auth/google**

Redirect pengguna ke halaman login Google OAuth. Tidak ada request body.

---

**GET /auth/google/callback**

Callback dari Google setelah autentikasi OAuth. Sistem membuat atau menghubungkan akun, kemudian redirect ke frontend dengan access token dan refresh token sebagai query parameter.

Redirect target: `<FRONTEND_URL>/auth/callback?access_token=...&refresh_token=...`

---

**POST /auth/refresh**

Menukar refresh token yang valid dengan pasangan token baru (token rotation: refresh token lama dihapus).

Request body:
```json
{
  "refresh_token": "a3f8c2d1e4b7..."
}
```

Response `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "b9e2f4a6c1d3..."
}
```

---

**POST /auth/logout** `[Memerlukan JWT]`

Menghapus refresh token dari database sehingga tidak dapat digunakan lagi.

Request body:
```json
{
  "refresh_token": "a3f8c2d1e4b7..."
}
```

Response `200 OK`:
```json
{
  "message": "Logged out successfully"
}
```

---

### Wallets

**GET /api/wallets** `[Memerlukan JWT]`

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

**GET /api/wallets/:id** `[Memerlukan JWT]`

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

**POST /api/wallets** `[Memerlukan JWT]`

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

**GET /api/categories** `[Memerlukan JWT]`

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

**GET /api/categories/:id** `[Memerlukan JWT]`

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

**POST /api/categories** `[Memerlukan JWT]`

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

**DELETE /api/categories/:id** `[Memerlukan JWT]`

Menghapus kategori berdasarkan ID.

| Parameter | Tipe | Keterangan |
|---|---|---|
| `id` | integer (path) | ID kategori |

Response `200 OK`:
```json
{
  "message": "Category deleted successfully"
}
```

Response `404 Not Found`:
```json
{ "error": "category not found" }
```

---

### Chat (Pencatatan Transaksi via AI)

**POST /api/chat** `[Memerlukan JWT]`

Memproses pesan teks bebas dari pengguna dan mencatat transaksi secara otomatis menggunakan AI.

Endpoint ini menerapkan rate limiter (10 request per menit per IP) dan membatasi ukuran body request maksimal 4KB.

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

Response `429 Too Many Requests` (rate limit terlampaui):
```json
{
  "success": false,
  "error": "Too many requests, please try again later"
}
```

Catatan: Jika wallet atau kategori yang disebutkan AI belum ada di database, sistem akan membuatnya secara otomatis.

---

### Analytics

**GET /api/summary** `[Memerlukan JWT]`

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
- (Opsional) Google OAuth credentials untuk login via Google
- (Opsional) Konfigurasi SMTP untuk pengiriman email verifikasi

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

# JWT
JWT_SECRET=supersecretkey
JWT_EXPIRY_MINUTES=15
REFRESH_TOKEN_EXPIRY_DAYS=7

# Google OAuth (opsional)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# SMTP untuk email verifikasi (opsional, jika kosong sistem log ke console)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=

# URL frontend untuk redirect setelah OAuth
FRONTEND_URL=http://localhost:3000
```

Migrasi tabel dan seeding data default akan dilakukan secara otomatis saat server pertama kali dijalankan.

**Catatan SMTP:** Jika `SMTP_HOST` tidak diisi, sistem tidak mengirim email sungguhan. Token verifikasi dicetak ke log console sehingga tetap bisa digunakan untuk development.

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
go test ./tests/... -v -run TestAuth
go test ./tests/... -v -run TestWallet
go test ./tests/... -v -run TestCategory
go test ./tests/... -v -run TestChat
go test ./tests/... -v -run TestAnalytics
go test ./tests/... -v -run TestPing
go test ./tests/... -v -run TestSecurity
```

### Catatan tentang Test

- Test menggunakan database yang sama dengan konfigurasi di `.env`, bukan database terpisah. Setiap test case akan membersihkan (TRUNCATE) tabel yang relevan sebelum dijalankan.
- AI service digantikan dengan `MockAIService` agar test tidak bergantung pada koneksi Gemini API dan memberikan hasil yang deterministik.
- `setup_test.go` berisi helper `SetupTestRouter()`, `CleanDatabase()`, dan `DoRequest()` yang digunakan bersama oleh semua file test.