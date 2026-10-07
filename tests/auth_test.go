package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"money-tracker-ai/config"
	"money-tracker-ai/models"

	"github.com/stretchr/testify/assert"
)

func TestAuth_Flow(t *testing.T) {
	t.Run("Registrasi manual & verifikasi email & login flow", func(t *testing.T) {
		CleanDatabase()

		// 1. Register
		regBody := map[string]interface{}{
			"email":    "user1@example.com",
			"name":     "User One",
			"password": "password123",
		}
		w := DoRequestWithToken("POST", "/auth/register", regBody, "")
		assert.Equal(t, http.StatusCreated, w.Code)

		// Try registering duplicate email -> error
		wDup := DoRequestWithToken("POST", "/auth/register", regBody, "")
		assert.Equal(t, http.StatusBadRequest, wDup.Code)

		// Try login before verified -> error
		loginBody := map[string]interface{}{
			"email":    "user1@example.com",
			"password": "password123",
		}
		wLoginUnverif := DoRequestWithToken("POST", "/auth/login", loginBody, "")
		assert.Equal(t, http.StatusUnauthorized, wLoginUnverif.Code)

		// 2. Verify email
		var ev models.EmailVerification
		err := config.DB.First(&ev).Error
		assert.NoError(t, err)

		wVerif := DoRequestWithToken("GET", fmt.Sprintf("/auth/verify?token=%s", ev.Token), nil, "")
		assert.Equal(t, http.StatusOK, wVerif.Code)

		// 3. Login after verified
		wLogin := DoRequestWithToken("POST", "/auth/login", loginBody, "")
		assert.Equal(t, http.StatusOK, wLogin.Code)

		var loginResp map[string]interface{}
		json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
		accessToken := loginResp["access_token"].(string)
		refreshToken := loginResp["refresh_token"].(string)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// 4. Refresh token
		refBody := map[string]interface{}{
			"refresh_token": refreshToken,
		}
		wRefresh := DoRequestWithToken("POST", "/auth/refresh", refBody, "")
		assert.Equal(t, http.StatusOK, wRefresh.Code)
		var refResp map[string]interface{}
		json.Unmarshal(wRefresh.Body.Bytes(), &refResp)
		newAccessToken := refResp["access_token"].(string)
		newRefreshToken := refResp["refresh_token"].(string)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)

		// Old refresh token rotated out -> should fail if reused
		wOldRefresh := DoRequestWithToken("POST", "/auth/refresh", refBody, "")
		assert.Equal(t, http.StatusUnauthorized, wOldRefresh.Code)

		// 5. Logout
		wLogout := DoRequestWithToken("POST", "/auth/logout", map[string]interface{}{
			"refresh_token": newRefreshToken,
		}, newAccessToken)
		assert.Equal(t, http.StatusOK, wLogout.Code)
	})

	t.Run("JWT Middleware protection", func(t *testing.T) {
		CleanDatabase()
		// Request protected endpoint without token -> 401
		wNoToken := DoRequestWithToken("GET", "/api/wallets", nil, "")
		assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)

		// Request protected endpoint with invalid token -> 401
		wInvalid := DoRequestWithToken("GET", "/api/wallets", nil, "invalid.jwt.token")
		assert.Equal(t, http.StatusUnauthorized, wInvalid.Code)
	})

	t.Run("Isolasi Data antar user", func(t *testing.T) {
		CleanDatabase()

		// User A creates wallet
		userA := &models.User{Email: "usera@example.com", Name: "User A", IsVerified: true}
		config.DB.Create(userA)
		walletA := &models.Wallet{UserID: userA.ID, Name: "Wallet User A", Balance: 100000}
		config.DB.Create(walletA)

		// User B creates wallet
		userB := &models.User{Email: "userb@example.com", Name: "User B", IsVerified: true}
		config.DB.Create(userB)
		walletB := &models.Wallet{UserID: userB.ID, Name: "Wallet User B", Balance: 200000}
		config.DB.Create(walletB)

		tokenA := GenerateTestToken(userA.ID)
		tokenB := GenerateTestToken(userB.ID)

		// User A lists wallets -> sees only Wallet A
		wA := DoRequestWithToken("GET", "/api/wallets", nil, tokenA)
		assert.Equal(t, http.StatusOK, wA.Code)
		var respA map[string]interface{}
		json.Unmarshal(wA.Body.Bytes(), &respA)
		walletsA := respA["data"].([]interface{})
		assert.Len(t, walletsA, 1)
		assert.Equal(t, "Wallet User A", walletsA[0].(map[string]interface{})["name"])

		// User A attempts to access Wallet B by ID -> 404
		wAccessB := DoRequestWithToken("GET", fmt.Sprintf("/api/wallets/%d", walletB.ID), nil, tokenA)
		assert.Equal(t, http.StatusNotFound, wAccessB.Code)

		// User B lists wallets -> sees only Wallet B
		wB := DoRequestWithToken("GET", "/api/wallets", nil, tokenB)
		assert.Equal(t, http.StatusOK, wB.Code)
		var respB map[string]interface{}
		json.Unmarshal(wB.Body.Bytes(), &respB)
		walletsB := respB["data"].([]interface{})
		assert.Len(t, walletsB, 1)
		assert.Equal(t, "Wallet User B", walletsB[0].(map[string]interface{})["name"])
	})

	t.Run("Category CRUD & isolation & global category protection", func(t *testing.T) {
		CleanDatabase()

		// Global Category
		globalCat := &models.Category{Name: "Global Cat", Type: "expense"}
		config.DB.Create(globalCat)

		// User category
		uid := DefaultTestUser.ID
		userCat := &models.Category{UserID: &uid, Name: "User Cat", Type: "expense"}
		config.DB.Create(userCat)

		// User lists categories -> gets global + user categories
		token := GenerateTestToken(DefaultTestUser.ID)
		wList := DoRequestWithToken("GET", "/api/categories", nil, token)
		assert.Equal(t, http.StatusOK, wList.Code)
		var resp map[string]interface{}
		json.Unmarshal(wList.Body.Bytes(), &resp)
		cats := resp["data"].([]interface{})
		assert.Len(t, cats, 2)

		// User cannot delete global category -> error
		wDelGlobal := DoRequestWithToken("DELETE", fmt.Sprintf("/api/categories/%d", globalCat.ID), nil, token)
		assert.Equal(t, http.StatusBadRequest, wDelGlobal.Code)

		// User can delete their own category
		wDelUser := DoRequestWithToken("DELETE", fmt.Sprintf("/api/categories/%d", userCat.ID), nil, token)
		assert.Equal(t, http.StatusOK, wDelUser.Code)
	})
}
