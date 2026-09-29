package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
)

func TestCategory_GetAll(t *testing.T) {
	t.Run("Skenario 5.1: Mengambil daftar kategori saat data kosong", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("GET", "/api/categories", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Empty(t, response["data"])
	})

	t.Run("Skenario 5.2: Mengambil daftar kategori saat data tersedia", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})
		config.DB.Create(&models.Category{Name: "Gaji", Type: "income"})

		w := DoRequest("GET", "/api/categories", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

func TestCategory_GetByID(t *testing.T) {
	t.Run("Skenario 6.1: Mengambil kategori dengan ID valid dan terdaftar", func(t *testing.T) {
		CleanDatabase()
		cat := models.Category{Name: "Belanja", Type: "expense"}
		config.DB.Create(&cat)

		w := DoRequest("GET", "/api/categories/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Belanja", data["name"])
	})

	t.Run("Skenario 6.2: Mengambil kategori dengan ID yang tidak ada di database", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("GET", "/api/categories/9999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "category not found", response["error"])
	})

	t.Run("Skenario 6.3: Mengambil kategori dengan format ID tidak valid (bukan angka)", func(t *testing.T) {
		w := DoRequest("GET", "/api/categories/xyz", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid category ID", response["error"])
	})
}

func TestCategory_Create(t *testing.T) {
	t.Run("Skenario 7.1: Pembuatan kategori pengeluaran (expense) yang valid", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Kesehatan", "type": "expense"}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Kesehatan", data["name"])
	})

	t.Run("Skenario 7.2: Pembuatan kategori pemasukan (income) yang valid", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Freelance", "type": "income"}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Skenario 7.3: Validasi field name kosong", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "", "type": "expense"}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 7.4: Validasi field type kosong", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Investasi", "type": ""}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 7.5: Validasi field type tidak valid (selain 'income' / 'expense')", func(t *testing.T) {
		CleanDatabase()
		reqBody := map[string]interface{}{"name": "Tabungan", "type": "transfer"}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 7.6: Pembuatan kategori dengan nama duplikat (unique constraint)", func(t *testing.T) {
		CleanDatabase()
		config.DB.Create(&models.Category{Name: "Makanan", Type: "expense"})

		reqBody := map[string]interface{}{"name": "Makanan", "type": "expense"}
		w := DoRequest("POST", "/api/categories", reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Skenario 7.7: Format body JSON tidak valid (malformed JSON)", func(t *testing.T) {
		CleanDatabase()
		w := DoRequest("POST", "/api/categories", "invalid json")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
