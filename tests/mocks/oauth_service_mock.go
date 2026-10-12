package mocks

import (
	"money-tracker-ai/services"

	"github.com/stretchr/testify/mock"
)

type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetAuthURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func (m *MockOAuthService) ExchangeCode(code string) (*services.GoogleUserInfo, error) {
	args := m.Called(code)
	if args.Get(0) != nil {
		return args.Get(0).(*services.GoogleUserInfo), args.Error(1)
	}
	return nil, args.Error(1)
}
