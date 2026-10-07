package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChat(t *testing.T) {
	t.Run("Skenario 8.1: Transaksi pengeluaran (expense) normal dengan saldo mencukupi", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "GoPay", Balance: 100000})
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})

		msg := "Beli bakso 30k pake GoPay"
		mockAIService.ExpectedCalls = nil // reset mock
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:      models.ActionAddTransaction,
					Amount:      30000,
					Type:        "expense",
					Category:    "Makanan",
					Wallet:      "GoPay",
					Description: "Beli bakso",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		var wallet models.Wallet
		config.DB.Where("name = ? AND user_id = ?", "GoPay", DefaultTestUser.ID).First(&wallet)
		assert.Equal(t, float64(70000), wallet.Balance)
	})

	t.Run("Skenario 8.2: Transaksi pengeluaran dengan saldo tidak mencukupi (insufficient balance)", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "GoPay", Balance: 10000})
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})

		msg := "Beli makanan 50k pake GoPay"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:   models.ActionAddTransaction,
					Amount:   50000,
					Type:     "expense",
					Category: "Makanan",
					Wallet:   "GoPay",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
		assert.Equal(t, "insufficient balance", response["error"])
	})

	t.Run("Skenario 8.3: Transaksi pemasukan (income) normal", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "BCA", Balance: 500000})
		config.DB.Create(&models.Category{Name: "Gaji", Type: "income"})

		msg := "Gaji masuk 200k ke BCA"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:   models.ActionAddTransaction,
					Amount:   200000,
					Type:     "income",
					Category: "Gaji",
					Wallet:   "BCA",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		var wallet models.Wallet
		config.DB.Where("name = ? AND user_id = ?", "BCA", DefaultTestUser.ID).First(&wallet)
		assert.Equal(t, float64(700000), wallet.Balance)
	})

	t.Run("Skenario 8.4: Transaksi dengan wallet baru yang belum terdaftar di database", func(t *testing.T) {
		CleanDatabase()

		msg := "Dapat hadiah 50k ke ShopeePay"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:   models.ActionAddTransaction,
					Amount:   50000,
					Type:     "income",
					Category: "Hadiah",
					Wallet:   "ShopeePay",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody, "10.0.1.1")

		assert.Equal(t, http.StatusOK, w.Code)

		var wallet models.Wallet
		config.DB.Where("LOWER(name) = LOWER(?) AND user_id = ?", "ShopeePay", DefaultTestUser.ID).First(&wallet)
		assert.Equal(t, float64(50000), wallet.Balance)
	})

	t.Run("Skenario 8.5: Transaksi dengan kategori baru yang belum terdaftar di database", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 50000})

		msg := "Donasi 10k dari Cash"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:   models.ActionAddTransaction,
					Amount:   10000,
					Type:     "expense",
					Category: "Donasi",
					Wallet:   "Cash",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody, "10.0.1.2")

		assert.Equal(t, http.StatusOK, w.Code)

		var cat models.Category
		config.DB.Where("name = ? AND user_id = ?", "Donasi", DefaultTestUser.ID).First(&cat)
		assert.Equal(t, "expense", cat.Type)
	})

	t.Run("Skenario 8.6: Transaksi pengeluaran yang melebihi batas budget kategori (Budget Warning)", func(t *testing.T) {
		CleanDatabase()
		wallet := models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 100000}
		config.DB.Create(&wallet)

		limit := float64(100000)
		uid := DefaultTestUser.ID
		cat := models.Category{UserID: &uid, Name: "Hiburan", Type: "expense", BudgetLimit: limit}
		config.DB.Create(&cat)
		config.DB.Create(&models.Transaction{
			WalletID: wallet.ID, CategoryID: cat.ID, Amount: 80000, Type: "expense",
		}) // sisa budget 20000

		msg := "Nonton 30k dari Cash"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{
					Action:   models.ActionAddTransaction,
					Amount:   30000,
					Type:     "expense",
					Category: "Hiburan",
					Wallet:   "Cash",
				},
			},
		}, nil).Once()

		reqBody := map[string]interface{}{"message": msg}
		w := DoRequest("POST", "/api/chat", reqBody, "10.0.1.3")

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		msgResponse, ok := response["message"].(string)
		assert.True(t, ok)
		assert.Contains(t, msgResponse, "⚠️ PERINGATAN")
	})

	t.Run("Skenario 8.7: Request body kosong atau tidak memuat field message", func(t *testing.T) {
		w := DoRequest("POST", "/api/chat", map[string]interface{}{}, "10.0.1.4")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Field 'message' wajib diisi", response["error"])
	})

	t.Run("Skenario 8.8: Error saat proses parsing AI", func(t *testing.T) {
		CleanDatabase()

		msg := "test error"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return((*models.ParsedIntent)(nil), errors.New("timeout")).Once()

		w := DoRequest("POST", "/api/chat", map[string]interface{}{"message": msg}, "10.0.1.5")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response["success"].(bool))
	})

	t.Run("Skenario 8.9: Multiple transaction input dalam satu pesan", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 500000})
		config.DB.Create(&models.Category{Name: "Belanja", Type: "expense"})
		config.DB.Create(&models.Category{Name: "Transportasi", Type: "expense"})

		msg := "Beli susu 25k, bensin 50k pake Cash"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentRecordTransaction,
			Actions: []models.ParsedAction{
				{Action: models.ActionAddTransaction, Amount: 25000, Type: "expense", Category: "Belanja", Wallet: "Cash", Description: "susu"},
				{Action: models.ActionAddTransaction, Amount: 50000, Type: "expense", Category: "Transportasi", Wallet: "Cash", Description: "bensin"},
			},
		}, nil).Once()

		w := DoRequest("POST", "/api/chat", map[string]interface{}{"message": msg}, "10.0.1.6")
		assert.Equal(t, http.StatusOK, w.Code)

		var wallet models.Wallet
		config.DB.Where("name = ? AND user_id = ?", "Cash", DefaultTestUser.ID).First(&wallet)
		assert.Equal(t, float64(425000), wallet.Balance)
	})

	t.Run("Skenario 8.10: Update saldo & create wallet via chat", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "Cash", Balance: 100000})

		msg := "update saldo cash 0, saldo BRI 500k"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentManageWallet,
			Actions: []models.ParsedAction{
				{Action: models.ActionSetBalance, Wallet: "cash", Balance: 0},
				{Action: models.ActionCreateWallet, Wallet: "BRI", Balance: 500000},
			},
		}, nil).Once()

		w := DoRequest("POST", "/api/chat", map[string]interface{}{"message": msg}, "10.0.1.7")
		assert.Equal(t, http.StatusOK, w.Code)

		var cash models.Wallet
		config.DB.Where("name = ? AND user_id = ?", "Cash", DefaultTestUser.ID).First(&cash)
		assert.Equal(t, float64(0), cash.Balance)

		var bri models.Wallet
		config.DB.Where("name = ? AND user_id = ?", "Bri", DefaultTestUser.ID).First(&bri)
		assert.Equal(t, float64(500000), bri.Balance)
	})

	t.Run("Skenario 8.11: Hapus wallet via chat", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Wallet{UserID: DefaultTestUser.ID, Name: "OldWallet", Balance: 100000})

		msg := "hapus wallet OldWallet"
		mockAIService.ExpectedCalls = nil
		mockAIService.On("ParseIntentPrompt", msg, mock.Anything).Return(&models.ParsedIntent{
			Intent: models.IntentManageWallet,
			Actions: []models.ParsedAction{
				{Action: models.ActionDeleteWallet, Wallet: "OldWallet"},
			},
		}, nil).Once()

		w := DoRequest("POST", "/api/chat", map[string]interface{}{"message": msg}, "10.0.1.8")
		assert.Equal(t, http.StatusOK, w.Code)

		var count int64
		config.DB.Model(&models.Wallet{}).Where("name = ? AND user_id = ?", "OldWallet", DefaultTestUser.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
