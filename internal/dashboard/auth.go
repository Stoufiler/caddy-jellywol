package dashboard

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/sirupsen/logrus"
)

// OIDCProvider handles OIDC authentication
type OIDCProvider struct {
	config   config.OIDCConfig
	logger   *logrus.Logger
	sessions sync.Map // session ID -> session data

	// OIDC endpoints (discovered from issuer)
	authEndpoint  string
	tokenEndpoint string
	userEndpoint  string
}

// Session holds user session data
type Session struct {
	UserID      string
	Email       string
	AccessToken string
	ExpiresAt   time.Time
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider(cfg config.OIDCConfig, logger *logrus.Logger) (*OIDCProvider, error) {
	if !cfg.Enabled {
		return &OIDCProvider{config: cfg, logger: logger}, nil
	}

	provider := &OIDCProvider{
		config: cfg,
		logger: logger,
	}

	// Discover OIDC endpoints
	if err := provider.discover(); err != nil {
		return nil, fmt.Errorf("OIDC discovery failed: %w", err)
	}

	return provider, nil
}

// discover fetches OIDC configuration from the issuer
func (p *OIDCProvider) discover() error {
	discoveryURL := strings.TrimSuffix(p.config.IssuerURL, "/") + "/.well-known/openid-configuration"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery endpoint returned %d", resp.StatusCode)
	}

	var config struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		UserinfoEndpoint      string `json:"userinfo_endpoint"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return err
	}

	p.authEndpoint = config.AuthorizationEndpoint
	p.tokenEndpoint = config.TokenEndpoint
	p.userEndpoint = config.UserinfoEndpoint

	p.logger.Infof("OIDC discovery complete: auth=%s, token=%s", p.authEndpoint, p.tokenEndpoint)
	return nil
}

// IsEnabled returns whether OIDC is enabled
func (p *OIDCProvider) IsEnabled() bool {
	return p.config.Enabled
}

// AuthMiddleware protects routes with OIDC authentication
func (p *OIDCProvider) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("jwp_session")
		if err != nil || cookie.Value == "" {
			p.redirectToLogin(w, r)
			return
		}

		// Validate session
		sessionData, ok := p.sessions.Load(cookie.Value)
		if !ok {
			p.redirectToLogin(w, r)
			return
		}

		session := sessionData.(*Session)
		if time.Now().After(session.ExpiresAt) {
			p.sessions.Delete(cookie.Value)
			p.redirectToLogin(w, r)
			return
		}

		// Session valid, continue
		next.ServeHTTP(w, r)
	})
}

// redirectToLogin redirects to the OIDC authorization endpoint
func (p *OIDCProvider) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state parameter
	state := generateRandomString(32)

	// Store state in cookie for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "jwp_oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Build authorization URL
	scopes := p.config.Scopes
	if scopes == "" {
		scopes = "openid email profile"
	}

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		p.authEndpoint,
		url.QueryEscape(p.config.ClientID),
		url.QueryEscape(p.config.RedirectURL),
		url.QueryEscape(scopes),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// CallbackHandler handles the OIDC callback
func (p *OIDCProvider) CallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !p.config.Enabled {
			http.Error(w, "OIDC not enabled", http.StatusNotFound)
			return
		}

		// Verify state
		stateCookie, err := r.Cookie("jwp_oauth_state")
		if err != nil {
			http.Error(w, "Missing state cookie", http.StatusBadRequest)
			return
		}

		state := r.URL.Query().Get("state")
		if state != stateCookie.Value {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "jwp_oauth_state",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		// Check for errors
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errParam, errDesc), http.StatusBadRequest)
			return
		}

		// Exchange code for token
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}

		token, err := p.exchangeCode(r.Context(), code)
		if err != nil {
			p.logger.Errorf("Token exchange failed: %v", err)
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		// Get user info
		userInfo, err := p.getUserInfo(r.Context(), token.AccessToken)
		if err != nil {
			p.logger.Errorf("Failed to get user info: %v", err)
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		// Create session
		sessionID := generateRandomString(32)
		session := &Session{
			UserID:      userInfo.Sub,
			Email:       userInfo.Email,
			AccessToken: token.AccessToken,
			ExpiresAt:   time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		}
		p.sessions.Store(sessionID, session)

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "jwp_session",
			Value:    sessionID,
			Path:     "/",
			MaxAge:   token.ExpiresIn,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})

		p.logger.Infof("User %s logged in via OIDC", userInfo.Email)

		// Redirect to dashboard
		http.Redirect(w, r, "/status", http.StatusTemporaryRedirect)
	}
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents OIDC user info
type UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// exchangeCode exchanges an authorization code for tokens
func (p *OIDCProvider) exchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// getUserInfo fetches user info from the userinfo endpoint
func (p *OIDCProvider) getUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if p.userEndpoint == "" {
		return nil, errors.New("userinfo endpoint not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d", resp.StatusCode)
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// LogoutHandler handles user logout
func (p *OIDCProvider) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear session
		cookie, err := r.Cookie("jwp_session")
		if err == nil && cookie.Value != "" {
			p.sessions.Delete(cookie.Value)
		}

		// Clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "jwp_session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		// Redirect to status page
		http.Redirect(w, r, "/status", http.StatusTemporaryRedirect)
	}
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
