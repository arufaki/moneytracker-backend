package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// apiEndpoints adalah data dokumentasi semua endpoint API.
// Dipisah ke package-level var agar mudah di-maintain tanpa
// menyentuh logika handler sama sekali.
var apiEndpoints = []gin.H{
	{
		"method":      "GET",
		"path":        "/",
		"description": "Dokumentasi petunjuk penggunaan API ini",
		"example_response": gin.H{
			"name":        "MoneyTracker AI API",
			"version":     "1.0",
			"description": "REST API untuk aplikasi pencatat keuangan berbasis AI.",
			"endpoints":   "[ ... ] (daftar endpoint seperti dokumen ini)",
		},
	},
	{
		"method":      "GET",
		"path":        "/ping",
		"description": "Healthcheck endpoint untuk memeriksa status server dan database",
		"example_response": gin.H{
			"status":    "ok",
			"db_status": "connected",
		},
	},
	{
		"method":      "GET",
		"path":        "/api/wallets",
		"description": "Mendapatkan daftar semua dompet (wallet)",
		"example_response": gin.H{
			"data": []gin.H{
				{
					"id":         1,
					"name":       "Dompet Utama",
					"balance":    1500000,
					"created_at": "2026-10-01T10:00:00Z",
				},
			},
		},
	},
	{
		"method":      "GET",
		"path":        "/api/wallets/:id",
		"description": "Mendapatkan detail dompet berdasarkan ID",
		"example_response": gin.H{
			"data": gin.H{
				"id":         1,
				"name":       "Dompet Utama",
				"balance":    1500000,
				"created_at": "2026-10-01T10:00:00Z",
			},
		},
	},
	{
		"method":      "POST",
		"path":        "/api/wallets",
		"description": "Membuat dompet baru",
		"example_request": gin.H{
			"name":    "Tabungan",
			"balance": 500000,
		},
		"example_response": gin.H{
			"message": "Wallet created successfully",
			"data": gin.H{
				"id":         2,
				"name":       "Tabungan",
				"balance":    500000,
				"created_at": "2026-10-04T12:00:00Z",
			},
		},
	},
	{
		"method":      "GET",
		"path":        "/api/categories",
		"description": "Mendapatkan daftar semua kategori transaksi",
		"example_response": gin.H{
			"data": []gin.H{
				{
					"id":   1,
					"name": "Makanan & Minuman",
					"type": "expense",
				},
			},
		},
	},
	{
		"method":      "GET",
		"path":        "/api/categories/:id",
		"description": "Mendapatkan detail kategori berdasarkan ID",
		"example_response": gin.H{
			"data": gin.H{
				"id":   1,
				"name": "Makanan & Minuman",
				"type": "expense",
			},
		},
	},
	{
		"method":      "POST",
		"path":        "/api/categories",
		"description": "Membuat kategori transaksi baru",
		"example_request": gin.H{
			"name": "Investasi",
			"type": "expense",
		},
		"example_response": gin.H{
			"message": "Category created successfully",
			"data": gin.H{
				"id":   3,
				"name": "Investasi",
				"type": "expense",
			},
		},
	},
	{
		"method":      "POST",
		"path":        "/api/chat",
		"description": "Mencatat transaksi menggunakan masukan teks berbasis AI",
		"example_request": gin.H{
			"message": "Beli makan siang 25rb pakai Dompet Utama",
		},
		"example_response": gin.H{
			"status":  "success",
			"message": "Transaction parsed and created successfully",
			"data": gin.H{
				"id":            10,
				"wallet_id":     1,
				"category_id":   1,
				"amount":        25000,
				"type":          "expense",
				"description":   "Beli makan siang",
				"wallet_name":   "Dompet Utama",
				"category_name": "Makanan & Minuman",
			},
		},
	},
	{
		"method":      "GET",
		"path":        "/api/summary",
		"description": "Mendapatkan ringkasan / analytics transaksi keuangan",
		"example_response": gin.H{
			"data": gin.H{
				"total_balance": 1475000,
				"total_income":  2000000,
				"total_expense": 525000,
			},
		},
	},
}

type RootController struct{}

func NewRootController() *RootController {
	return &RootController{}
}

// GetAPIDocumentation godoc
// @Summary     Get API Documentation
// @Description Mengembalikan dokumentasi ringkas penggunaan API dalam format JSON
// @Tags        Root
// @Produce     json
// @Success     200  {object}  map[string]interface{}
// @Router      / [get]
func (ctrl *RootController) GetAPIDocumentation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "MoneyTracker AI API",
		"version":     "1.0",
		"description": "REST API untuk aplikasi pencatat keuangan berbasis AI.",
		"endpoints":   apiEndpoints,
	})
}
