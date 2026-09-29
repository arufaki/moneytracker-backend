package services

import (
	"context"
	"encoding/json"
	"fmt"
	"money-tracker-ai/models"
	"money-tracker-ai/repositories"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
	apiKey  string
}

func NewAIService(logRepo repositories.AILogRepository) AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY is not set in environment variables")
	}
	return &aiService{
		logRepo: logRepo,
		apiKey:  apiKey,
	}
}

// ParseTransactionPrompt mengirim pesan user ke Gemini AI dan meng-parse
// respons JSON-nya menjadi struct ParsedTransaction.
func (s *aiService) ParseTransactionPrompt(userMessage string) (*models.ParsedTransaction, error) {
	ctx := context.Background()

	// Inisialisasi Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Setup model dengan system instruction
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	// Kirim pesan ke AI
	resp, err := model.GenerateContent(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from Gemini: %w", err)
	}

	// Ekstrak teks dari respons
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}
	rawText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Bersihkan dari kemungkinan markdown yang lolos dari instruksi
	cleanedText := strings.TrimSpace(rawText)
	cleanedText = strings.TrimPrefix(cleanedText, "```json")
	cleanedText = strings.TrimSuffix(cleanedText, "```")
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
		fmt.Printf("warning: failed to save AI log: %v\n", logErr)
	}

	return &parsed, nil
}
