package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"money-tracker-ai/config"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	t.Run("Skenario 1.1: Database terhubung normal (Healthy)", func(t *testing.T) {
		w := DoRequest("GET", "/ping", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "connected", response["db_status"])
	})

	t.Run("Skenario 1.2: Database terputus / error (Unhealthy)", func(t *testing.T) {
		// Simulasikan database terputus
		sqlDB, _ := config.DB.DB()
		sqlDB.Close()

		w := DoRequest("GET", "/ping", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "disconnected", response["db_status"])

		// Reconnect untuk test lain
		SetupTestRouter()
	})
}
