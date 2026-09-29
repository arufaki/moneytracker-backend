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

	"google.golang.org/genai"
)

// System Prompt: Instruksi ke AI untuk selalu merespons dalam format JSON
// yang sesuai dengan struct ParsedTransaction. Jangan ubah prompt ini sembarangan!
const systemPrompt = `Kamu adalah asisten pencatat keuangan. 
Tugasmu adalah mengekstrak informasi transaksi keuangan dari pesan pengguna dan merespons HANYA dengan JSON valid.

Format JSON yang harus kamu kembalikan (TANPA markdown, TANPA tanda backtick, HANYA raw JSON):
{
  "amount": <angka dalam satuan rupiah, tanpa titik/koma>,
  "type": "<'income' atau 'expense'>",
  "category": "<pilih dari: Makanan, Transportasi, Gaji, Hiburan, Belanja, Tagihan, atau tebak yang paling sesuai>",
  "wallet": "<nama dompet yang disebutkan user, jika tidak ada tulis 'Cash'>",
  "description": "<deskripsi singkat dalam bahasa Indonesia>"
}

Contoh input: "Habis beli dimsum mentai 35k pake GoPay"
Contoh output: {"amount":35000,"type":"expense","category":"Makanan","wallet":"GoPay","description":"dimsum mentai"}

PENTING: Jangan tambahkan teks apapun selain JSON. Tidak ada penjelasan, tidak ada markdown.`

type AIService interface {
	ParseTransactionPrompt(userMessage string) (*models.ParsedTransaction, error)
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

// ParseTransactionPrompt mengirim pesan user ke Gemini AI dan meng-parse
// respons JSON-nya menjadi struct ParsedTransaction.
func (s *aiService) ParseTransactionPrompt(userMessage string) (*models.ParsedTransaction, error) {
	ctx := context.Background()

	// Setup model config dengan system instruction
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	// Kirim pesan ke AI
	resp, err := s.client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Text(userMessage), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from Gemini: %w", err)
	}

	// Ekstrak teks dari respons
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	rawText := resp.Text()

	// Bersihkan dari kemungkinan markdown yang lolos dari instruksi
	cleanedText := strings.TrimSpace(rawText)
	if idx := strings.Index(cleanedText, "```json"); idx != -1 {
		cleanedText = cleanedText[idx+7:]
	}
	if idx := strings.LastIndex(cleanedText, "```"); idx != -1 {
		cleanedText = cleanedText[:idx]
	}
	cleanedText = strings.TrimSpace(cleanedText)

	// Parse JSON ke struct
	var parsed models.ParsedTransaction
	if err := json.Unmarshal([]byte(cleanedText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w. Raw response: %s", err, cleanedText)
	}

	// Simpan ke ai_logs sebagai audit trail (async tidak perlu, cukup sync)
	logEntry := &models.AILog{
		RawMessage:    userMessage,
		ExtractedJSON: cleanedText,
	}
	// Log error tapi jangan blokir response ke user
	if logErr := s.logRepo.Create(logEntry); logErr != nil {
		log.Printf("warning: failed to save AI log: %v\n", logErr)
	}

	return &parsed, nil
}
