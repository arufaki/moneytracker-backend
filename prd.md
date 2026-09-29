# 📄 Product Requirement Document (PRD)

## Project Name: Money Tracker AI API



---



## 1. Project Overview

**Money Tracker AI API** adalah backend service berbasis **Go (Golang)** yang mengintegrasikan **Google Gemini API** untuk mengotomatisasi pencatatan keuangan pribadi melalui perintah teks alami (Natural Language Processing). 



Sistem dapat memproses pesan seperti *"Habis beli dimsum 35k pake GoPay"*, mengekstrak datanya secara presisi menjadi JSON terstruktur, dan memperbarui saldo dompet serta riwayat transaksi secara *real-time* di **PostgreSQL** menggunakan **GORM** dengan proteksi *concurrency*.



---



## 2. Tech Stack & Architecture

* **Language:** Go (Golang) 1.20+

* **Framework:** Gin-Gonic (HTTP Web Framework)

* **ORM:** GORM (`gorm.io/gorm`)

* **Database & Driver:** PostgreSQL (`gorm.io/driver/postgres`)

* **AI Provider:** Google Gemini API SDK (`github.com/google/generative-ai-go`)

* **Concurrency Safety:** `sync.Mutex`

* **Architecture Pattern:** Clean Architecture (Config, Controllers, Services, Repositories)



---



## 3. Core Features & User Stories



### 🤖 Feature 1: Smart AI Chat & Parsing (`POST /api/chat`)

* **User Story:** Sebagai pengguna, saya ingin mencatat transaksi dengan mengetik pesan teks santai agar tidak perlu mengisi form manual yang rumit.

* **Flow Sistem:**

  1. Client mengirim pesan ke endpoint: `{"message": "Beli dimsum mentai 35k dari GoPay"}`.

  2. Service mengoper pesan ke Gemini AI dengan *System Prompt* khusus.

  3. AI merespons dengan JSON terstruktur:

     ```json

     {

       "amount": 35000,

       "type": "expense",

       "category": "Makanan",

       "wallet": "GoPay",

       "description": "dimsum mentai"

     }

     ```

  4. Service mengeksekusi transaksi database via GORM: mengurangi saldo dompet `GoPay` dan memasukkan baris baru ke tabel `transactions`.

  5. AI mengembalikan pesan konfirmasi ramah ke client beserta informasi saldo terkini.



### 💳 Feature 2: Wallet & Account Management (`/api/wallets`)

* **User Story:** Sebagai pengguna, saya ingin mengelola beberapa sumber dana (Cash, Bank BCA, GoPay) dan memantau saldonya.

* **Capabilities:**

  * Auto-adjust balance (Saldo otomatis bertambah jika *income* dan berkurang jika *expense*).

  * Proteksi *Mutex Locking* (`sync.Mutex`) saat kalkulasi saldo agar aman dari *race condition*.



### 🏷️ Feature 3: Category Management (`/api/categories`)

* **User Story:** Sebagai pengguna, saya ingin setiap transaksi dikelompokkan ke dalam kategori tertentu.

* **Capabilities:**

  * Kategori bawaan: *Makanan, Transportasi, Gaji, Hiburan, Belanja, Tagihan*.

  * Dukungan penambahan kategori kustom oleh user.



### 📊 Feature 4: Financial Analytics & Summary (`GET /api/summary`)

* **User Story:** Sebagai pengguna, saya ingin melihat laporan ringkasan pengeluaran dan pemasukan saya.

* **Output Data:**

  * Total Pemasukan & Pengeluaran bulanan.

  * Total Net Balance (Gabungan dari seluruh saldo dompet yang aktif).

  * Persentase & breakdown pengeluaran per kategori.



### ⚠️ Feature 5: Budget Warning System

* **User Story:** Sebagai pengguna, saya ingin diberi peringatan jika pengeluaran kategori tertentu sudah melebihi batas anggaran.

* **Capabilities:**

  * Batas budget per kategori (misal: Batas Makanan Rp 1.500.000/bulan).

  * Peringatan disisipkan langsung pada pesan balik AI jika transaksi membuat kategori tersebut *overbudget*.



---



## 4. GORM Domain Models & Database Schema



### GORM Struct Definitions (Go)



```go

package models



import (

"time"

)



// 1. Model Wallet (Sumber Dana)

type Wallet struct {

ID        uint      `gorm:"primaryKey" json:"id"`

Name      string    `gorm:"type:varchar(100);unique;not null" json:"name"`

Balance   float64   `gorm:"type:numeric(15,2);default:0.0" json:"balance"`

CreatedAt time.Time `json:"created_at"`

UpdatedAt time.Time `json:"updated_at"`

}



// 2. Model Category (Kategori Transaksi)

type Category struct {

ID        uint      `gorm:"primaryKey" json:"id"`

Name      string    `gorm:"type:varchar(100);unique;not null" json:"name"`

Type      string    `gorm:"type:varchar(20);not null" json:"type"` // 'income' / 'expense'

CreatedAt time.Time `json:"created_at"`

UpdatedAt time.Time `json:"updated_at"`

}



// 3. Model Transaction (Riwayat Transaksi)

type Transaction struct {

ID          uint      `gorm:"primaryKey" json:"id"`

WalletID    uint      `gorm:"not null" json:"wallet_id"`

Wallet      Wallet    `gorm:"foreignKey:WalletID;constraint:OnDelete:CASCADE;" json:"wallet"`

CategoryID  uint      `gorm:"not null" json:"category_id"`

Category    Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;" json:"category"`

Amount      float64   `gorm:"type:numeric(15,2);not null" json:"amount"`

Type        string    `gorm:"type:varchar(20);not null" json:"type"` // 'income' / 'expense'

Description string    `gorm:"type:text" json:"description"`

CreatedAt   time.Time `json:"created_at"`

}



// 4. Model AILog (Audit Trail AI)

type AILog struct {

ID            uint      `gorm:"primaryKey" json:"id"`

RawMessage    string    `gorm:"type:text;not null" json:"raw_message"`

ExtractedJSON string    `gorm:"type:jsonb;not null" json:"extracted_json"`

CreatedAt     time.Time `json:"created_at"`

}

