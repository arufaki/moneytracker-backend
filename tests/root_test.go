package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// totalExpectedEndpoints harus diupdate setiap ada endpoint baru didaftarkan.
// Ini sengaja eksplisit agar developer tidak lupa update dokumentasi.
const totalExpectedEndpoints = 10

func TestGetAPIDocumentation(t *testing.T) {
	t.Run("GET / returns 200 OK with API documentation JSON", func(t *testing.T) {
		w := DoRequest("GET", "/", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Validasi field utama
		assert.Equal(t, "MoneyTracker AI API", response["name"])
		assert.Equal(t, "1.0", response["version"])
		assert.NotEmpty(t, response["description"])

		// Validasi jumlah endpoint
		endpoints, ok := response["endpoints"].([]interface{})
		assert.True(t, ok, "field 'endpoints' harus berupa array")
		assert.Len(t, endpoints, totalExpectedEndpoints,
			"jumlah endpoint terdokumentasi harus %d, update jika ada endpoint baru", totalExpectedEndpoints)

		// Validasi struktur tiap endpoint: harus punya method, path, description, example_response
		requiredFields := []string{"method", "path", "description", "example_response"}
		for i, ep := range endpoints {
			epMap, ok := ep.(map[string]interface{})
			assert.True(t, ok, "endpoint index %d harus berupa object", i)

			for _, field := range requiredFields {
				assert.Contains(t, epMap, field,
					"endpoint index %d (%v) harus punya field '%s'", i, epMap["path"], field)
			}

			// method harus salah satu dari HTTP method yang valid
			method, _ := epMap["method"].(string)
			validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
			assert.Contains(t, validMethods, method,
				"endpoint index %d punya method tidak valid: %s", i, method)

			// path tidak boleh kosong
			path, _ := epMap["path"].(string)
			assert.NotEmpty(t, path, "endpoint index %d punya path kosong", i)
		}
	})
}
