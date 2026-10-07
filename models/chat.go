package models

type ChatRequest struct {
	Message string `json:"message" binding:"required,max=500"`
}

type ChatResponse struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	Transaction   *Transaction      `json:"transaction,omitempty"`
	Transactions  []*Transaction    `json:"transactions,omitempty"`
	UpdatedWallet *Wallet           `json:"updated_wallet,omitempty"`
	ParsedData    *ParsedTransaction `json:"parsed_data,omitempty"`
	ParsedIntent  *ParsedIntent     `json:"parsed_intent,omitempty"`
}

