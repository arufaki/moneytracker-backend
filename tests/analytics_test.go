package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
)

func TestSummary(t *testing.T) {
	t.Run("Skenario 9.1: Ringkasan saat belum ada data transaksi sama sekali", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("GET", "/api/summary", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(0), response["income"])
		assert.Equal(t, float64(0), response["expense"])
		assert.Equal(t, float64(0), response["net_balance"])
		assert.Empty(t, response["breakdown"])
	})

	t.Run("Skenario 9.2: Ringkasan tanpa query parameter (Default bulan & tahun berjalan)", func(t *testing.T) {
		CleanDatabase()
		wWallet := models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 500000}
		config.DB.Create(&wWallet)
		cat := models.Category{Name: "Makanan", Type: "expense"}
		config.DB.Create(&cat)

		now := time.Now().UTC()
		config.DB.Create(&models.Transaction{
			WalletID: wWallet.ID, CategoryID: cat.ID, Amount: 10000, Type: "expense",
			CreatedAt: now,
		})
		
		lastMonth := now.AddDate(0, -1, 0)
		config.DB.Create(&models.Transaction{
			WalletID: wWallet.ID, CategoryID: cat.ID, Amount: 20000, Type: "expense",
			CreatedAt: lastMonth,
		})

		w := DoRequest("GET", "/api/summary", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, float64(10000), response["expense"])
	})

	t.Run("Skenario 9.3: Ringkasan dengan filter bulan dan tahun spesifik", func(t *testing.T) {
		CleanDatabase()
		wWallet := models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 500000}
		config.DB.Create(&wWallet)
		cat := models.Category{Name: "Makanan", Type: "expense"}
		config.DB.Create(&cat)

		dateMay2026 := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
		config.DB.Create(&models.Transaction{
			WalletID: wWallet.ID, CategoryID: cat.ID, Amount: 15000, Type: "expense",
			CreatedAt: dateMay2026,
		})

		dateJune2026 := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
		config.DB.Create(&models.Transaction{
			WalletID: wWallet.ID, CategoryID: cat.ID, Amount: 25000, Type: "expense",
			CreatedAt: dateJune2026,
		})

		w := DoRequest("GET", "/api/summary?month=5&year=2026", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(15000), response["expense"])
	})

	t.Run("Skenario 9.4: Kalkulasi Net Balance dari seluruh wallet aktif", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "Wallet A", Balance: 300000})
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "Wallet B", Balance: 200000})

		w := DoRequest("GET", "/api/summary", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(500000), response["net_balance"])
	})

	t.Run("Skenario 9.5: Kalkulasi persentase dan perincian (breakdown) per kategori pengeluaran", func(t *testing.T) {
		CleanDatabase()
		wWallet := models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 500000}
		config.DB.Create(&wWallet)
		
		catMakanan := models.Category{Name: "Makanan", Type: "expense"}
		config.DB.Create(&catMakanan)
		catTransport := models.Category{Name: "Transportasi", Type: "expense"}
		config.DB.Create(&catTransport)

		now := time.Now().UTC()
		config.DB.Create(&models.Transaction{WalletID: wWallet.ID, CategoryID: catMakanan.ID, Amount: 75000, Type: "expense", CreatedAt: now})
		config.DB.Create(&models.Transaction{WalletID: wWallet.ID, CategoryID: catTransport.ID, Amount: 25000, Type: "expense", CreatedAt: now})

		w := DoRequest("GET", "/api/summary", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		breakdowns := response["breakdown"].([]interface{})
		assert.Len(t, breakdowns, 2)

		var mknPct, trsPct float64
		for _, b := range breakdowns {
			bk := b.(map[string]interface{})
			if bk["category_name"] == "Makanan" {
				mknPct = bk["percentage"].(float64)
			}
			if bk["category_name"] == "Transportasi" {
				trsPct = bk["percentage"].(float64)
			}
		}

		assert.Equal(t, float64(75), mknPct)
		assert.Equal(t, float64(25), trsPct)
	})

	t.Run("Skenario 9.6: Verifikasi status is_over_budget pada breakdown kategori", func(t *testing.T) {
		CleanDatabase()
		wWallet := models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 500000}
		config.DB.Create(&wWallet)
		
		limit := float64(100000)
		catBelanja := models.Category{Name: "Belanja", Type: "expense", BudgetLimit: limit}
		config.DB.Create(&catBelanja)

		now := time.Now().UTC()
		config.DB.Create(&models.Transaction{WalletID: wWallet.ID, CategoryID: catBelanja.ID, Amount: 150000, Type: "expense", CreatedAt: now})

		w := DoRequest("GET", "/api/summary", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		breakdowns := response["breakdown"].([]interface{})
		
		assert.True(t, breakdowns[0].(map[string]interface{})["is_over_budget"].(bool))
	})

	t.Run("Skenario 9.7: Validasi nilai query parameter month di luar rentang 1 - 12", func(t *testing.T) {
		w := DoRequest("GET", "/api/summary?month=13&year=2026", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "month must be between 1 and 12", response["error"])
	})

	t.Run("Skenario 9.8: Validasi format query parameter month bukan angka", func(t *testing.T) {
		w := DoRequest("GET", "/api/summary?month=july", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "month")
	})

	t.Run("Skenario 9.9: Validasi format query parameter year bukan angka", func(t *testing.T) {
		w := DoRequest("GET", "/api/summary?month=6&year=duaribu", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"].(string), "year")
	})
}
