package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIDocumentation(t *testing.T) {
	t.Run("GET / returns 200 OK with API documentation JSON", func(t *testing.T) {
		w := DoRequest("GET", "/", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "MoneyTracker AI API", response["name"])
		assert.Equal(t, "1.0", response["version"])

		endpoints, ok := response["endpoints"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, endpoints)
	})
}
