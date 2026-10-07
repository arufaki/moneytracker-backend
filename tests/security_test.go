package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func DoRequestWithIP(method, url string, body interface{}, ip string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = []byte(v)
		default:
			reqBody, _ = json.Marshal(body)
		}
	}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if ip != "" {
		req.Header.Set("X-Forwarded-For", ip)
		req.RemoteAddr = ip + ":12345"
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestSecurityHardening(t *testing.T) {
	t.Run("Input validation: message kosong ke /api/chat mengembalikan 400", func(t *testing.T) {
		w := DoRequestWithIP("POST", "/api/chat", map[string]interface{}{"message": ""}, "10.0.0.1")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Input validation: message > 500 karakter mengembalikan 400", func(t *testing.T) {
		longMsg := strings.Repeat("a", 501)
		w := DoRequestWithIP("POST", "/api/chat", map[string]interface{}{"message": longMsg}, "10.0.0.2")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Body size limit: request body > 4KB mengembalikan 413 atau 400", func(t *testing.T) {
		hugeMsg := strings.Repeat("x", 5000)
		w := DoRequestWithIP("POST", "/api/chat", map[string]interface{}{"message": hugeMsg}, "10.0.0.3")
		assert.True(t, w.Code == http.StatusRequestEntityTooLarge || w.Code == http.StatusBadRequest)
	})

	t.Run("Rate limiting: request berulang melampaui limit mengembalikan 429", func(t *testing.T) {
		reqBody := map[string]interface{}{"message": "test rate limit"}
		testIP := "10.0.0.4"
		var lastCode int
		for i := 0; i < 12; i++ {
			w := DoRequestWithIP("POST", "/api/chat", reqBody, testIP)
			lastCode = w.Code
			if w.Code == http.StatusTooManyRequests {
				break
			}
		}
		assert.Equal(t, http.StatusTooManyRequests, lastCode)
	})

	t.Run("Error tidak membocorkan detail database/internal", func(t *testing.T) {
		w := DoRequestWithIP("GET", "/api/wallets/99999", nil, "10.0.0.5")
		bodyStr := w.Body.String()
		assert.NotContains(t, bodyStr, "pq:")
		assert.NotContains(t, bodyStr, "gorm")
		assert.NotContains(t, bodyStr, "sql:")
		assert.NotContains(t, bodyStr, "GEMINI")

		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "wallet not found", resp["error"])
	})
}
