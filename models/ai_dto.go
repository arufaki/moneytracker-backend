package models

// ParsedTransaction adalah struct yang mewakili hasil ekstraksi data transaksi
// dari pesan teks bebas user yang diproses oleh Gemini AI.
type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type ParsedTransaction struct {
	Amount      float64         `json:"amount"`      // Nominal uang, contoh: 35000
	Type        TransactionType `json:"type"`        // "income" atau "expense"
	Category    string          `json:"category"`    // Nama kategori, contoh: "Makanan"
	Wallet      string          `json:"wallet"`      // Nama wallet, contoh: "GoPay"
	Description string          `json:"description"` // Deskripsi singkat, contoh: "dimsum mentai"
}
