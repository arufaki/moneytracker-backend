package models

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	Transaction   *Transaction      `json:"transaction,omitempty"`
	UpdatedWallet *Wallet           `json:"updated_wallet,omitempty"`
	ParsedData    *ParsedTransaction `json:"parsed_data,omitempty"`
}
