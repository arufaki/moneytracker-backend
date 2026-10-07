package models

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type ActionType string

const (
	ActionAddTransaction ActionType = "add_transaction"
	ActionSetBalance     ActionType = "set_balance"
	ActionCreateWallet   ActionType = "create_wallet"
	ActionDeleteWallet   ActionType = "delete_wallet"
)

type IntentType string

const (
	IntentRecordTransaction IntentType = "record_transaction"
	IntentManageWallet      IntentType = "manage_wallet"
	IntentUnknown           IntentType = "unknown"
)

type ParsedAction struct {
	Action      ActionType      `json:"action"`
	Amount      float64         `json:"amount,omitempty"`
	Balance     float64         `json:"balance,omitempty"`
	Type        TransactionType `json:"type,omitempty"`
	Category    string          `json:"category,omitempty"`
	Wallet      string          `json:"wallet,omitempty"`
	Description string          `json:"description,omitempty"`
}

type ParsedIntent struct {
	Intent  IntentType     `json:"intent"`
	Actions []ParsedAction `json:"actions"`
}



