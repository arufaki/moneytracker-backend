package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
)

func TestWallet_GetAll(t *testing.T) {
	t.Run("Skenario 2.1: Mengambil daftar wallet saat data kosong", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("GET", "/api/wallets", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Empty(t, response["data"])
	})

	t.Run("Skenario 2.2: Mengambil daftar wallet saat data tersedia", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "Cash", Balance: 100000})
		config.DB.Create(&models.Wallet{Name: "BCA", Balance: 500000})

		w := DoRequest("GET", "/api/wallets", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

func TestWallet_GetByID(t *testing.T) {
	t.Run("Skenario 3.1: Mengambil wallet dengan ID yang valid dan terdaftar", func(t *testing.T) {
		CleanDatabase()
		wallet := models.Wallet{Name: "OVO", Balance: 50000}
		config.DB.Create(&wallet)

		url := fmt.Sprintf("/api/wallets/%d", wallet.ID)
		w := DoRequest("GET", url, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "OVO", data["name"])
	})

	t.Run("Skenario 3.2: Mengambil wallet dengan ID yang tidak ditemukan di database", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("GET", "/api/wallets/9999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "wallet not found", response["error"])
	})

	t.Run("Skenario 3.3: Mengambil wallet dengan format ID tidak valid", func(t *testing.T) {
		w := DoRequest("GET", "/api/wallets/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid wallet ID", response["error"])
	})
}

func TestWallet_Create(t *testing.T) {
	t.Run("Skenario 4.1: Pembuatan wallet dengan data valid dan saldo awal positif", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Bank Jago", "balance": 500000}
		w := DoRequest("POST", "/api/wallets", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Bank Jago", data["name"])
		assert.Equal(t, float64(500000), data["balance"])
	})

	t.Run("Skenario 4.2: Pembuatan wallet tanpa saldo / saldo 0 (default balance)", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "OVO", "balance": 0}
		w := DoRequest("POST", "/api/wallets", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(0), data["balance"])
	})

	t.Run("Skenario 4.3: Validasi field name kosong atau tidak disertakan", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "", "balance": 100000}
		w := DoRequest("POST", "/api/wallets", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 4.4: Validasi saldo awal negatif", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Dompet Minus", "balance": -50000}
		w := DoRequest("POST", "/api/wallets", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 4.5: Pembuatan wallet dengan nama yang sudah ada (duplikasi / unique constraint)", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "BCA", Balance: 100000})
		
		reqBody := map[string]interface{}{"name": "BCA", "balance": 100000}
		w := DoRequest("POST", "/api/wallets", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 4.6: Format body JSON tidak valid (malformed JSON)", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("POST", "/api/wallets", "invalid json")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
