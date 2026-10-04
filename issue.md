# Issue: Add API Documentation on Root Endpoint `/`

## Background

Backend ini sudah di-deploy di Vercel (serverless). Swagger UI tidak bisa dipakai di environment serverless karena butuh static file serving yang tidak tersedia. Solusinya: buat endpoint `/` yang langsung me-return dokumentasi API dalam bentuk JSON — berisi daftar endpoint beserta contoh response-nya.

---

## Goal

Ketika user/developer membuka `GET /`, response-nya berupa JSON yang menjelaskan cara penggunaan API: ada daftar endpoint lengkap dan contoh response tiap endpoint.

---

## Scope

Tidak ada perubahan di endpoint lain. Hanya menambah satu handler baru di root `/`.

---

## Steps

### 1. Buat handler baru di `main.go` atau `routes/routes.go`

Daftarkan route `GET /` di Gin engine, **sebelum** `SetupRoutes` dipanggil.

```go
r.GET("/", func(c *gin.Context) {
    c.JSON(200, gin.H{ /* isi dokumentasi */ })
})
```

### 2. Isi response dengan daftar endpoint & contoh response

Struktur JSON yang dikembalikan cukup berisi:
- `title` dan `version`
- `base_url`
- `endpoints` → array of object, tiap object berisi:
  - `method`
  - `path`
  - `description`
  - `example_response`

Endpoint yang perlu didokumentasikan (lihat `routes/routes.go`):

| Method | Path | Keterangan |
|---|---|---|
| GET | /api/wallets | List semua wallet |
| GET | /api/wallets/:id | Detail wallet |
| POST | /api/wallets | Buat wallet baru |
| GET | /api/categories | List semua kategori |
| GET | /api/categories/:id | Detail kategori |
| POST | /api/categories | Buat kategori baru |
| POST | /api/chat | Chat AI untuk input transaksi |
| GET | /api/summary | Ringkasan / analytics |
| GET | /ping | Healthcheck |

Untuk `example_response`, cukup gunakan data dummy yang representatif (tidak perlu hit DB).

### 3. Pastikan tidak ada konflik route

Verifikasi tidak ada route `/` yang sudah terdaftar sebelumnya di `main.go` maupun `routes/routes.go`.

### 4. Test manual

Hit `GET /` di local dan di Vercel, pastikan response JSON muncul dengan benar.

---

## Out of Scope

- Tidak perlu interactive UI (HTML)
- Tidak perlu Swagger/OpenAPI spec
- Tidak perlu autentikasi di endpoint ini
- Tidak perlu validasi request body di endpoint ini

---

## Acceptance Criteria

- [ ] `GET /` mengembalikan HTTP 200 dengan `Content-Type: application/json`
- [ ] Response berisi semua endpoint yang ada beserta contoh response-nya
- [ ] Tidak ada breaking change di endpoint lain
- [ ] Berjalan di Vercel (serverless)
