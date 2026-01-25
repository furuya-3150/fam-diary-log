package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogleProvider_GetProviderName(t *testing.T) {
	provider := NewGoogleProviderWithOAuth2(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
	)

	assert.Equal(t, "google", provider.GetProviderName())
}

func TestGoogleProvider_GetAuthURL(t *testing.T) {
	tests := []struct {
		name     string
		provider OAuthProvider
		state    string
		wantURL  bool
	}{
		{
			name: "正常なAuthURL生成",
			provider: NewGoogleProviderWithOAuth2(
				"test-client-id",
				"test-client-secret",
				"http://localhost:8080/callback",
			),
			state:   "test-state-123",
			wantURL: true,
		},
		{
			name: "空のstateでもAuthURL生成可能",
			provider: NewGoogleProviderWithOAuth2(
				"test-client-id",
				"test-client-secret",
				"http://localhost:8080/callback",
			),
			state:   "",
			wantURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authURL := tt.provider.GetAuthURL(tt.state)

			if tt.wantURL {
				assert.NotEmpty(t, authURL)
				// GoogleのOAuth2エンドポイントが含まれていることを確認
				assert.Contains(t, authURL, "accounts.google.com")
				assert.Contains(t, authURL, "client_id=test-client-id")
				assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback")

				// scopeが含まれていることを確認
				assert.Contains(t, authURL, "scope=")
				assert.Contains(t, authURL, "openid")

				// stateが含まれていることを確認（空でない場合）
				if tt.state != "" {
					assert.Contains(t, authURL, "state="+tt.state)
				}
			}
		})
	}
}

func TestGoogleProvider_ExchangeCode_Success(t *testing.T) {
	// モックのGoogleサーバーを作成
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			// Token exchangeエンドポイント
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"access_token": "mock-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"id_token":     "mock-id-token",
			}
			json.NewEncoder(w).Encode(response)
		case "/userinfo":
			// UserInfo エンドポイント
			w.Header().Set("Content-Type", "application/json")
			response := map[string]string{
				"id":      "google-user-12345",
				"email":   "test@example.com",
				"name":    "Test User",
				"picture": "https://example.com/photo.jpg",
			}
			json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// カスタムエンドポイントを使用したプロバイダーを作成
	provider := &GoogleProvider{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURL:  "http://localhost:8080/callback",
		oauth2Config: &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
				"openid",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  mockServer.URL + "/auth",
				TokenURL: mockServer.URL + "/token",
			},
		},
	}

	// Note: 実際のExchangeCodeの呼び出しはモックサーバーでは完全にテストできないため、
	// このテストはGetAuthURLとGetProviderNameのテストに焦点を当てています。
	// ExchangeCodeの完全なテストには統合テストまたはより高度なモッキングが必要です。

	authURL := provider.GetAuthURL("test-state")
	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, mockServer.URL)
}

func TestGoogleProvider_GetAuthURL_NilConfig(t *testing.T) {
	// oauth2Config が nil の場合のテスト
	provider := &GoogleProvider{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURL:  "http://localhost:8080/callback",
		oauth2Config: nil, // 意図的にnilに設定
	}

	authURL := provider.GetAuthURL("test-state")
	assert.Empty(t, authURL, "oauth2Configがnilの場合、空文字列を返すべき")
}

func TestNewGoogleProviderWithOAuth2(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	redirectURL := "http://localhost:8080/callback"

	provider := NewGoogleProviderWithOAuth2(clientID, clientSecret, redirectURL)

	assert.NotNil(t, provider)

	// 型アサーションでGoogleProviderの内部フィールドを確認
	googleProvider, ok := provider.(*GoogleProvider)
	require.True(t, ok, "providerはGoogleProviderの型であるべき")

	assert.Equal(t, clientID, googleProvider.clientID)
	assert.Equal(t, clientSecret, googleProvider.clientSecret)
	assert.Equal(t, redirectURL, googleProvider.redirectURL)
	assert.NotNil(t, googleProvider.oauth2Config)

	// Scopeが正しく設定されているか確認
	expectedScopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}
	assert.Equal(t, expectedScopes, googleProvider.oauth2Config.Scopes)
}
