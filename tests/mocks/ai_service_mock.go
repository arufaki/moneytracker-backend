package mocks

import (
	"money-tracker-ai/models"
	"github.com/stretchr/testify/mock"
)

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) ParseIntentPrompt(userMessage string, userID uint) (*models.ParsedIntent, error) {
	args := m.Called(userMessage, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ParsedIntent), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAIService) Close() {
	m.Called()
}
