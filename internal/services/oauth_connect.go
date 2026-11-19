package services

// "Sign in with Trunc": a small OAuth 2.0 authorization-code server for the
// Trunc plugins (Chrome extension, WordPress, Shopify, Canva).
//
// A plugin sends the user to /connect on trunc.site. After the user allows it,
// the browser goes back to the plugin with a one-time code, and the plugin
// exchanges that code at /api/v1/oauth/token for an API key of its own. The
// key appears on the user's API keys page, where it can be revoked.
//
// Public clients (the extension, WordPress, Shopify) must use PKCE (S256).
// Canva is a confidential client: Canva's servers hold a client secret.

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/URLshorter/url-shortener/internal/models"
)

const (
	connectCodeTTL    = 10 * time.Minute
	connectCodePrefix = "oauth:code:"
	// ConnectScope is the only scope: create, read, update and delete the
	// user's links and read their analytics (the default API key permissions).
	ConnectScope = "links"
	// connectKeyMarker starts the description of every key issued here, so
	// connected apps are recognisable on the API keys page.
	connectKeyMarker = "Connected with Sign in with Trunc"
)

// ConnectError is an OAuth 2.0 error (RFC 6749 section 5.2).
type ConnectError struct {
	Code        string
	Description string
	Status      int
}

func (e *ConnectError) Error() string { return e.Code + ": " + e.Description }

func connectErr(status int, code, desc string) *ConnectError {
	return &ConnectError{Code: code, Description: desc, Status: status}
}

// ConnectClient is one registered plugin.
type ConnectClient struct {
	ID   string
	Name string
	// Confidential clients authenticate with a secret; public ones must use PKCE.
	Confidential bool
	// AnyHost means the redirect host is chosen by the plugin (a WordPress
	// site or a Shopify app URL), so the consent page shows it prominently.
	AnyHost bool
	secret  string
	allow   func(u *url.URL) bool
}

// ConnectStore keeps one-time codes. RedisStorage implements it.
type ConnectStore interface {
	SetNX(key string, value interface{}, ttl time.Duration) (bool, error)
	GetDel(key string) (string, error)
}

// ConnectKeys issues and revokes API keys. APIKeyService implements it.
type ConnectKeys interface {
	CreateAPIKey(userID int64, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error)
	ValidateAPIKey(key string) (*models.APIKey, error)
	DeleteAPIKey(id uuid.UUID, userID int64) error
}

// ConnectService implements the authorize, token and revoke steps.
type ConnectService struct {
	store   ConnectStore
	keys    ConnectKeys
	clients map[string]*ConnectClient
	now     func() time.Time
}

var (
	chromiumAppHost = regexp.MustCompile(`^[a-p]{32}\.chromiumapp\.org$`)
	pkceChallenge   = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	pkceVerifier    = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)
)

// CanvaRedirectURI is the only redirect Canva's OAuth broker uses.
const CanvaRedirectURI = "https://www.canva.com/apps/oauth/authorized"

func isLocalHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "[::1]" || h == "::1"
}

// cleanRedirect rejects redirects that could carry credentials or confuse the
// consent page: user info, fragments, and anything that is not http(s).
func cleanRedirect(u *url.URL) bool {
	if u == nil || u.User != nil || u.Fragment != "" || u.Opaque != "" || u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		return isLocalHost(u.Hostname())
	}
	return false
}

// ConnectOptions configures the clients that depend on a deployment.
type ConnectOptions struct {
	// CanvaSecret is the client secret shared with Canva; empty disables Canva.
	CanvaSecret string
	// ShopifyRedirects are the exact callback URLs of the Shopify app
	// deployments (<SHOPIFY_APP_URL>/auth/trunc/callback); empty disables it.
	ShopifyRedirects []string
}

// NewConnectService registers the plugins. Canva and Shopify are only
// registered when their deployment settings are configured.
func NewConnectService(store ConnectStore, keys ConnectKeys, opts ConnectOptions) *ConnectService {
	s := &ConnectService{store: store, keys: keys, clients: map[string]*ConnectClient{}, now: time.Now}
	s.clients["trunc-chrome"] = &ConnectClient{
		ID: "trunc-chrome", Name: "Trunc for Chrome",
		allow: func(u *url.URL) bool {
			return u.Scheme == "https" && u.Port() == "" && chromiumAppHost.MatchString(u.Hostname())
		},
	}
	s.clients["trunc-wordpress"] = &ConnectClient{
		ID: "trunc-wordpress", Name: "Trunc for WordPress", AnyHost: true,
		allow: func(u *url.URL) bool {
			return strings.Contains(u.Path, "/wp-admin/")
		},
	}
	// The Shopify app is one deployment with a fixed callback, so only that
	// exact URL is accepted (anything wider would let a phishing link pass).
	shopify := map[string]bool{}
	for _, r := range opts.ShopifyRedirects {
		if u, err := url.Parse(strings.TrimSpace(r)); err == nil && u.Scheme == "https" && cleanRedirect(u) {
			shopify[u.String()] = true
		}
	}
	if len(shopify) > 0 {
		s.clients["trunc-shopify"] = &ConnectClient{
			ID: "trunc-shopify", Name: "Trunc for Shopify",
			allow: func(u *url.URL) bool { return shopify[u.String()] },
		}
	}
	if opts.CanvaSecret != "" {
		s.clients["trunc-canva"] = &ConnectClient{
			ID: "trunc-canva", Name: "Trunc for Canva", Confidential: true, secret: opts.CanvaSecret,
			allow: func(u *url.URL) bool { return u.String() == CanvaRedirectURI },
		}
	}
	return s
}

// Client checks that clientID is registered and that redirectURI is one it may
// use. It returns the client and the parsed redirect.
func (s *ConnectService) Client(clientID, redirectURI string) (*ConnectClient, *url.URL, error) {
	c, ok := s.clients[clientID]
	if !ok {
		return nil, nil, connectErr(http.StatusBadRequest, "invalid_client", "Unknown app.")
	}
	u, err := url.Parse(redirectURI)
	if err != nil || len(redirectURI) > 2048 || !cleanRedirect(u) || !c.allow(u) {
		return nil, nil, connectErr(http.StatusBadRequest, "invalid_request", "This app may not send you back to that address.")
	}
	return c, u, nil
}

// AuthorizeParams are the values from the /connect page.
type AuthorizeParams struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	State               string `json:"state"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	// Decision is "allow" or "deny".
	Decision string `json:"decision"`
}

type storedCode struct {
	UserID      int64  `json:"uid"`
	ClientID    string `json:"cid"`
	RedirectURI string `json:"ru"`
	Challenge   string `json:"cc,omitempty"`
	Scope       string `json:"sc"`
	IssuedAt    int64  `json:"iat"`
}

func withQuery(u *url.URL, vals map[string]string) string {
	out := *u
	q := out.Query()
	for k, v := range vals {
		if v != "" {
			q.Set(k, v)
		}
	}
	out.RawQuery = q.Encode()
	return out.String()
}

func codeKey(code string) string {
	sum := sha256.Sum256([]byte(code))
	return connectCodePrefix + hex.EncodeToString(sum[:])
}

// Authorize records the user's decision and returns where to send the browser.
// Errors that mean the redirect itself can't be trusted are returned as errors;
// everything else goes back to the app as error= on the redirect.
func (s *ConnectService) Authorize(userID int64, p AuthorizeParams) (string, error) {
	c, u, err := s.Client(p.ClientID, p.RedirectURI)
	if err != nil {
		return "", err
	}
	if len(p.State) > 512 {
		return "", connectErr(http.StatusBadRequest, "invalid_request", "state is too long.")
	}
	back := func(code, desc string) (string, error) {
		return withQuery(u, map[string]string{"error": code, "error_description": desc, "state": p.State}), nil
	}
	if p.Decision != "allow" {
		return back("access_denied", "You cancelled the connection.")
	}
	if p.ResponseType != "code" {
		return back("unsupported_response_type", "Only response_type=code is supported.")
	}
	if p.Scope != "" && p.Scope != ConnectScope {
		return back("invalid_scope", "The only scope is "+ConnectScope+".")
	}
	if p.CodeChallenge != "" || !c.Confidential {
		if p.CodeChallengeMethod != "S256" || !pkceChallenge.MatchString(p.CodeChallenge) {
			return back("invalid_request", "A PKCE code_challenge with method S256 is required.")
		}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(raw)
	val, _ := json.Marshal(storedCode{
		UserID: userID, ClientID: c.ID, RedirectURI: p.RedirectURI,
		Challenge: p.CodeChallenge, Scope: ConnectScope, IssuedAt: s.now().Unix(),
	})
	ok, err := s.store.SetNX(codeKey(code), string(val), connectCodeTTL)
	if err != nil || !ok {
		return "", connectErr(http.StatusServiceUnavailable, "temporarily_unavailable", "Please try again.")
	}
	return withQuery(u, map[string]string{"code": code, "state": p.State}), nil
}

// TokenParams are the values a plugin posts to /api/v1/oauth/token.
type TokenParams struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
}

// TokenResult is the token response (RFC 6749 section 5.1).
type TokenResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func (s *ConnectService) checkSecret(c *ConnectClient, secret string) bool {
	return c.Confidential && secret != "" && subtle.ConstantTimeCompare([]byte(c.secret), []byte(secret)) == 1
}

func pkceMatches(challenge, verifier string) bool {
	if !pkceVerifier.MatchString(verifier) {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	return subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(challenge)) == 1
}

// Exchange turns a code into a new API key for the user who approved it.
func (s *ConnectService) Exchange(p TokenParams) (*TokenResult, error) {
	if p.GrantType != "authorization_code" {
		return nil, connectErr(http.StatusBadRequest, "unsupported_grant_type", "Only authorization_code is supported.")
	}
	c, ok := s.clients[p.ClientID]
	if !ok {
		return nil, connectErr(http.StatusUnauthorized, "invalid_client", "Unknown app.")
	}
	if c.Confidential && !s.checkSecret(c, p.ClientSecret) {
		return nil, connectErr(http.StatusUnauthorized, "invalid_client", "Client authentication failed.")
	}
	if p.Code == "" {
		return nil, connectErr(http.StatusBadRequest, "invalid_request", "code is required.")
	}
	raw, err := s.store.GetDel(codeKey(p.Code))
	if err != nil {
		return nil, connectErr(http.StatusServiceUnavailable, "temporarily_unavailable", "Please try again.")
	}
	var sc storedCode
	if raw == "" || json.Unmarshal([]byte(raw), &sc) != nil {
		return nil, connectErr(http.StatusBadRequest, "invalid_grant", "The code is invalid, expired or already used.")
	}
	if sc.ClientID != c.ID || sc.RedirectURI != p.RedirectURI {
		return nil, connectErr(http.StatusBadRequest, "invalid_grant", "The code was issued to a different app or redirect.")
	}
	if s.now().Sub(time.Unix(sc.IssuedAt, 0)) > connectCodeTTL {
		return nil, connectErr(http.StatusBadRequest, "invalid_grant", "The code has expired.")
	}
	if sc.Challenge != "" && !pkceMatches(sc.Challenge, p.CodeVerifier) {
		return nil, connectErr(http.StatusBadRequest, "invalid_grant", "The code_verifier does not match.")
	}
	if sc.Challenge == "" && !c.Confidential {
		return nil, connectErr(http.StatusBadRequest, "invalid_grant", "PKCE is required for this app.")
	}

	host := ""
	if u, err := url.Parse(sc.RedirectURI); err == nil {
		host = u.Hostname()
	}
	desc := fmt.Sprintf("%s from %s on %s.", connectKeyMarker, host, s.now().UTC().Format("2 Jan 2006"))
	res, err := s.keys.CreateAPIKey(sc.UserID, &models.CreateAPIKeyRequest{
		Name:        c.Name,
		Description: &desc,
		Permissions: models.DefaultAPIKeyPermissions,
	})
	if err != nil {
		return nil, connectErr(http.StatusInternalServerError, "server_error", "Could not create a key.")
	}
	return &TokenResult{AccessToken: res.PlainKey, TokenType: "Bearer", Scope: ConnectScope}, nil
}

// ErrConnectClient means client authentication failed during revocation.
var ErrConnectClient = errors.New("invalid client")

// Revoke deletes the key behind token (RFC 7009). Unknown tokens succeed
// silently, as the RFC asks. Confidential clients must authenticate.
func (s *ConnectService) Revoke(token, clientID, clientSecret string) error {
	if c, ok := s.clients[clientID]; ok && c.Confidential && !s.checkSecret(c, clientSecret) {
		return ErrConnectClient
	}
	if token == "" {
		return nil
	}
	key, err := s.keys.ValidateAPIKey(token)
	if err != nil || key == nil {
		return nil
	}
	return s.keys.DeleteAPIKey(key.ID, key.UserID)
}
