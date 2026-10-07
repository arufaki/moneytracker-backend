package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// System Prompt: Instruksi ke AI untuk selalu merespons dalam format JSON
// yang sesuai dengan struct ParsedIntent. Jangan ubah prompt ini sembarangan!
const systemPrompt = `Kamu adalah asisten pencatat dan pengelola keuangan.
Tugasmu adalah menganalisis pesan pengguna dan merespons HANYA dengan JSON valid.

Tentukan "intent" utama dari pesan:
1. "manage_wallet": JIKA pengguna meminta mengubah/mengupdate/mengatur saldo wallet (misal: "update saldo", "set saldo", "saldo cash 0", "saldo BRI 500k"), membuat wallet baru, atau menghapus wallet. PENTING: Mengatur/mengubah saldo BUKAN transaksi pengeluaran/pemasukan!
2. "record_transaction": JIKA pengguna mencatat transaksi pengeluaran atau pemasukan (misal: beli makan, bayar bensin, gaji, transfer, belanja, dll).
3. "unknown": jika pesan tidak relevan.

PENTING UNTUK MULTIPLE INPUT:
Jika pengguna menyebutkan lebih dari satu transaksi atau lebih dari satu wallet dalam satu pesan (dipisahkan koma, titik koma, baris baru, atau kata sambung), kamu WAJIB memasukkan SEMUA item tersebut ke dalam array "actions". JANGAN ADA ITEM YANG TERLEWAT ATAU DIBUANG!

Setiap item dalam array "actions" memiliki field "action":
- Untuk intent "record_transaction": "action" adalah "add_transaction".
- Untuk intent "manage_wallet": "action" bisa "set_balance", "create_wallet", atau "delete_wallet".

Format JSON wajib:
{
  "intent": "<'record_transaction' | 'manage_wallet' | 'unknown'>",
  "actions": [
    {
      "action": "<'add_transaction' | 'set_balance' | 'create_wallet' | 'delete_wallet'>",
      "amount": <angka nominal transaksi jika add_transaction>,
      "balance": <angka nominal saldo jika set_balance atau create_wallet>,
      "type": "<'income' atau 'expense' jika add_transaction>",
      "category": "<kategori seperti Makanan, Transportasi, Gaji, Belanja, Tagihan, dll jika add_transaction>",
      "wallet": "<nama wallet yang disebutkan, default 'Cash' jika tidak disebutkan>",
      "description": "<deskripsi singkat jika add_transaction>"
    }
  ]
}

Contoh 1 (Multiple Transaksi Pengeluaran):
Input: "Beli susu 25k, bensin 50k, bayar wifi 300k pake cash"
Output: {"intent":"record_transaction","actions":[{"action":"add_transaction","amount":25000,"type":"expense","category":"Belanja","wallet":"Cash","description":"beli susu"},{"action":"add_transaction","amount":50000,"type":"expense","category":"Transportasi","wallet":"Cash","description":"bensin"},{"action":"add_transaction","amount":300000,"type":"expense","category":"Tagihan","wallet":"Cash","description":"bayar wifi"}]}

Contoh 2 (Update/Set Saldo & Create Wallet):
Input: "update saldo cash 0, saldo BRI 500k, saldo seabank 45k"
Output: {"intent":"manage_wallet","actions":[{"action":"set_balance","wallet":"Cash","balance":0},{"action":"set_balance","wallet":"BRI","balance":500000},{"action":"set_balance","wallet":"SeaBank","balance":45000}]}

Contoh 3 (Update Saldo Single):
Input: "update saldo cash 0"
Output: {"intent":"manage_wallet","actions":[{"action":"set_balance","wallet":"Cash","balance":0}]}

Contoh 4 (Hapus Wallet):
Input: "hapus wallet OldWallet"
Output: {"intent":"manage_wallet","actions":[{"action":"delete_wallet","wallet":"OldWallet"}]}

PENTING: Hanya kembalikan RAW JSON tanpa markdown, tanpa backtick, tanpa penjelasan.`

type AIService interface {
	ParseIntentPrompt(userMessage string) (*models.ParsedIntent, error)
	Close()
}

type aiService struct {
	logRepo repositories.AILogRepository
	client  *genai.Client
}

func NewAIService(logRepo repositories.AILogRepository) AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY is not set in environment variables")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		panic(fmt.Sprintf("failed to create Gemini client: %v", err))
	}

	return &aiService{
		logRepo: logRepo,
		client:  client,
	}
}

func (s *aiService) ParseIntentPrompt(userMessage string) (*models.ParsedIntent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	resp, err := s.client.Models.GenerateContent(ctx, "gemini-3.5-flash-lite", genai.Text(userMessage), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from Gemini: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	rawText := resp.Text()
	cleanedText := strings.TrimSpace(rawText)
	if idx := strings.Index(cleanedText, "```json"); idx != -1 {
		cleanedText = cleanedText[idx+7:]
	}
	if idx := strings.LastIndex(cleanedText, "```"); idx != -1 {
		cleanedText = cleanedText[:idx]
	}
	cleanedText = strings.TrimSpace(cleanedText)

	var parsed models.ParsedIntent
	if err := json.Unmarshal([]byte(cleanedText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w. Raw response: %s", err, cleanedText)
	}

	logEntry := &models.AILog{
		RawMessage:    userMessage,
		ExtractedJSON: cleanedText,
	}
	if logErr := s.logRepo.Create(logEntry); logErr != nil {
		log.Printf("warning: failed to save AI log: %v\n", logErr)
	}

	return &parsed, nil
}

func (s *aiService) Close() {
}

