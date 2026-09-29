package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
)

func TestChat(t *testing.T) {
	t.Run("Skenario 8.1: Transaksi pengeluaran (expense) normal dengan saldo mencukupi", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "GoPay", Balance: 100000})
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})

		msg := "Beli bakso 30k pake GoPay"
		mockAIService.ExpectedCalls = nil // reset mock
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:      30000,
			Type:        "expense",
			Category:    "Makanan",
			Wallet:      "GoPay",
			Description: "Beli bakso",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))

		var wallet models.Wallet
		config.DB.Where("name = ?", "GoPay").First(&wallet)
		assert.Equal(t, float64(70000), wallet.Balance)
	})

	t.Run("Skenario 8.2: Transaksi pengeluaran dengan saldo tidak mencukupi (insufficient balance)", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "GoPay", Balance: 10000})
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})

		msg := "Beli makanan 50k pake GoPay"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:   50000,
			Type:     "expense",
			Category: "Makanan",
			Wallet:   "GoPay",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.False(t, response["success"].(bool))
		assert.Equal(t, "insufficient balance", response["error"])
	})

	t.Run("Skenario 8.3: Transaksi pemasukan (income) normal", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "BCA", Balance: 500000})
		config.DB.Create(&models.Category{Name: "Gaji", Type: "income"})

		msg := "Gaji masuk 200k ke BCA"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:   200000,
			Type:     "income",
			Category: "Gaji",
			Wallet:   "BCA",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["success"].(bool))

		var wallet models.Wallet
		config.DB.Where("name = ?", "BCA").First(&wallet)
		assert.Equal(t, float64(700000), wallet.Balance)
	})

	t.Run("Skenario 8.4: Transaksi dengan wallet baru yang belum terdaftar di database", func(t *testing.T) {
		CleanDatabase()

		msg := "Dapat hadiah 50k ke ShopeePay"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:   50000,
			Type:     "income",
			Category: "Hadiah",
			Wallet:   "ShopeePay",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)

		var wallet models.Wallet
		config.DB.Where("name = ?", "ShopeePay").First(&wallet)
		assert.Equal(t, float64(50000), wallet.Balance)
	})

	t.Run("Skenario 8.5: Transaksi dengan kategori baru yang belum terdaftar di database", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "Cash", Balance: 50000})

		msg := "Donasi 10k dari Cash"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:   10000,
			Type:     "expense",
			Category: "Donasi",
			Wallet:   "Cash",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)

		var cat models.Category
		config.DB.Where("name = ?", "Donasi").First(&cat)
		assert.Equal(t, "expense", cat.Type)
	})

	t.Run("Skenario 8.6: Transaksi pengeluaran yang melebihi batas budget kategori (Budget Warning)", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{Name: "Cash", Balance: 100000})
		
		limit := float64(100000)
		config.DB.Create(&models.Category{Name: "Hiburan", Type: "expense", BudgetLimit: limit})
		config.DB.Create(&models.Transaction{
			WalletID: 1, CategoryID: 1, Amount: 80000, Type: "expense",
		}) // sisa budget 20000

		msg := "Nonton 30k dari Cash"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return(&models.ParsedTransaction{
			Amount:   30000,
			Type:     "expense",
			Category: "Hiburan",
			Wallet:   "Cash",
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		msgResponse, ok := response["message"].(string)
		assert.True(t, ok)
		assert.Contains(t, msgResponse, "⚠️ PERINGATAN")
	})

	t.Run("Skenario 8.7: Request body kosong atau tidak memuat field message", func(t *testing.T) {
		w := DoRequest("POST", "/api/chat", map[string]interface{}{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Field 'message' wajib diisi", response["error"])
	})

	t.Run("Skenario 8.8: Error saat proses parsing AI", func(t *testing.T) {
		CleanDatabase()
		
		msg := "test error"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseTransactionPrompt", msg).Return((*models.ParsedTransaction)(nil), errors.New("timeout")).Once()

		w := DoRequest("POST", "/api/chat", map[string]interface{}{"message": msg})
		
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.False(t, response["success"].(bool))
	})
}
