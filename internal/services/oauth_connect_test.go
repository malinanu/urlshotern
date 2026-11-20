package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/URLshorter/url-shortener/internal/models"
)

type memStore struct {
	mu sync.Mutex
	m  map[string]string
}

func (s *memStore) SetNX(key string, value interface{}, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[key]; ok {
		return false, nil
	}
	s.m[key] = value.(string)
	return true, nil
}

func (s *memStore) GetDel(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.m[key]
	delete(s.m, key)
	return v, nil
}

type memKeys struct {
	created []*models.CreateAPIKeyRequest
	owner   []int64
	keys    map[string]*models.APIKey
	deleted []uuid.UUID
}

func (k *memKeys) CreateAPIKey(userID int64, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	k.created = append(k.created, req)
	k.owner = append(k.owner, userID)
	plain := "us_test_" + uuid.NewString()
	key := &models.APIKey{ID: uuid.New(), UserID: userID, Name: req.Name}
	k.keys[plain] = key
	return &models.CreateAPIKeyResponse{APIKey: key, PlainKey: plain}, nil
}

func (k *memKeys) ValidateAPIKey(key string) (*models.APIKey, error) {
	if v, ok := k.keys[key]; ok {
		return v, nil
	}
	return nil, errors.New("invalid")
}

func (k *memKeys) DeleteAPIKey(id uuid.UUID, _ int64) error {
	k.deleted = append(k.deleted, id)
	for p, v := range k.keys {
		if v.ID == id {
			delete(k.keys, p)
		}
	}
	return nil
}

const (
	testVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	chromeRedir  = "https://abcdefghijklmnopabcdefghijklmnop.chromiumapp.org/cb"
	shopifyRedir = "https://trunc-app.example.com/auth/trunc/callback"
)

func challengeFor(v string) string {
	sum := sha256.Sum256([]byte(v))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newConnect() (*ConnectService, *memStore, *memKeys) {
	st := &memStore{m: map[string]string{}}
	ks := &memKeys{keys: map[string]*models.APIKey{}}
	return NewConnectService(st, ks, ConnectOptions{CanvaSecret: "canva-secret", ShopifyRedirects: []string{shopifyRedir}}), st, ks
}

func codeFrom(t *testing.T, redirect string) string {
	t.Helper()
	u, err := url.Parse(redirect)
	if err != nil {
		t.Fatal(err)
	}
	if e := u.Query().Get("error"); e != "" {
		t.Fatalf("redirect carries error %q: %s", e, redirect)
	}
	return u.Query().Get("code")
}

func allow(clientID, redirect string) AuthorizeParams {
	return AuthorizeParams{
		ClientID: clientID, RedirectURI: redirect, ResponseType: "code", State: "st-1",
		Scope: ConnectScope, CodeChallenge: challengeFor(testVerifier), CodeChallengeMethod: "S256", Decision: "allow",
	}
}

func TestConnectRedirectRules(t *testing.T) {
	s, _, _ := newConnect()
	ok := map[string][]string{
		"trunc-chrome":    {chromeRedir, "https://ponmlkjihgfedcbaponmlkjihgfedcba.chromiumapp.org/"},
		"trunc-wordpress": {"https://blog.example.com/wp-admin/admin.php?page=trunc", "http://localhost:8881/wp-admin/admin.php"},
		"trunc-shopify":   {shopifyRedir},
		"trunc-canva":     {CanvaRedirectURI},
	}
	bad := map[string][]string{
		"trunc-chrome": {
			"https://abcdefghijklmnopabcdefghijklmnop.chromiumapp.org.evil.com/",
			"https://evil.com/abcdefghijklmnopabcdefghijklmnop.chromiumapp.org",
			"https://abcdefghijklmnopabcdefghijklmnoz.chromiumapp.org/",
			"http://abcdefghijklmnopabcdefghijklmnop.chromiumapp.org/",
			"https://x@abcdefghijklmnopabcdefghijklmnop.chromiumapp.org/",
			"https://abcdefghijklmnopabcdefghijklmnop.chromiumapp.org:8443/",
		},
		"trunc-wordpress": {
			"https://blog.example.com/admin.php", "http://blog.example.com/wp-admin/", "javascript:alert(1)//wp-admin/",
			"https://blog.example.com/wp-admin/#frag", "https://u:p@blog.example.com/wp-admin/",
		},
		"trunc-shopify": {
			"https://trunc-app.example.com.evil.com/auth/trunc/callback",
			"https://evil.example/auth/trunc/callback",
			"https://trunc-app.example.com/auth/trunc/callback/x",
			"https://trunc-app.example.com/other",
			"http://trunc-app.example.com/auth/trunc/callback",
			"https://trunc-app.example.com/auth/trunc/callback#x",
			"data:text/html,hi",
		},
		"trunc-canva": {"https://www.canva.com/apps/oauth/authorized/x", "https://canva.com/apps/oauth/authorized"},
		"unknown":     {chromeRedir},
	}
	for id, list := range ok {
		for _, r := range list {
			if _, _, err := s.Client(id, r); err != nil {
				t.Errorf("%s should allow %s: %v", id, r, err)
			}
		}
	}
	for id, list := range bad {
		for _, r := range list {
			if _, _, err := s.Client(id, r); err == nil {
				t.Errorf("%s must reject %s", id, r)
			}
		}
	}
}

func TestConnectCanvaDisabledWithoutSecret(t *testing.T) {
	s := NewConnectService(&memStore{m: map[string]string{}}, &memKeys{keys: map[string]*models.APIKey{}}, ConnectOptions{})
	if _, _, err := s.Client("trunc-canva", CanvaRedirectURI); err == nil {
		t.Fatal("Canva must be unavailable without a client secret")
	}
	if _, _, err := s.Client("trunc-shopify", shopifyRedir); err == nil {
		t.Fatal("Shopify must be unavailable without a configured callback")
	}
}

func TestConnectFullFlowIssuesKey(t *testing.T) {
	s, _, ks := newConnect()
	redirect, err := s.Authorize(42, allow("trunc-chrome", chromeRedir))
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(redirect)
	if u.Query().Get("state") != "st-1" || !strings.HasPrefix(redirect, chromeRedir+"?") {
		t.Fatalf("unexpected redirect %s", redirect)
	}
	res, err := s.Exchange(TokenParams{GrantType: "authorization_code", Code: codeFrom(t, redirect), RedirectURI: chromeRedir, ClientID: "trunc-chrome", CodeVerifier: testVerifier})
	if err != nil {
		t.Fatal(err)
	}
	if res.TokenType != "Bearer" || res.AccessToken == "" || res.Scope != ConnectScope {
		t.Fatalf("bad token response %+v", res)
	}
	if len(ks.created) != 1 || ks.owner[0] != 42 || ks.created[0].Name != "Trunc for Chrome" {
		t.Fatalf("key not issued to the approving user: %+v %v", ks.created, ks.owner)
	}
	if d := ks.created[0].Description; d == nil || !strings.Contains(*d, "abcdefghijklmnopabcdefghijklmnop.chromiumapp.org") {
		t.Fatalf("description should name the redirect host: %v", d)
	}
}

func TestConnectCodeIsSingleUseAndBound(t *testing.T) {
	s, _, _ := newConnect()
	exchange := func(code, redirect, client, verifier string) error {
		_, err := s.Exchange(TokenParams{GrantType: "authorization_code", Code: code, RedirectURI: redirect, ClientID: client, CodeVerifier: verifier})
		return err
	}
	expectGrant := func(name string, err error) {
		t.Helper()
		var ce *ConnectError
		if !errors.As(err, &ce) || (ce.Code != "invalid_grant" && ce.Code != "invalid_client") {
			t.Errorf("%s: want invalid_grant/invalid_client, got %v", name, err)
		}
	}

	r1, _ := s.Authorize(1, allow("trunc-chrome", chromeRedir))
	c1 := codeFrom(t, r1)
	if err := exchange(c1, chromeRedir, "trunc-chrome", testVerifier); err != nil {
		t.Fatal(err)
	}
	expectGrant("reused code", exchange(c1, chromeRedir, "trunc-chrome", testVerifier))

	r2, _ := s.Authorize(1, allow("trunc-chrome", chromeRedir))
	expectGrant("wrong verifier", exchange(codeFrom(t, r2), chromeRedir, "trunc-chrome", strings.Repeat("a", 43)))

	r3, _ := s.Authorize(1, allow("trunc-chrome", chromeRedir))
	expectGrant("different redirect", exchange(codeFrom(t, r3), chromeRedir+"x", "trunc-chrome", testVerifier))

	r4, _ := s.Authorize(1, allow("trunc-chrome", chromeRedir))
	expectGrant("different client", exchange(codeFrom(t, r4), chromeRedir, "trunc-shopify", testVerifier))

	expectGrant("made-up code", exchange("not-a-code", chromeRedir, "trunc-chrome", testVerifier))

	// Expired: the stored issue time is checked even if the store kept the key.
	r5, _ := s.Authorize(1, allow("trunc-chrome", chromeRedir))
	s.now = func() time.Time { return time.Now().Add(11 * time.Minute) }
	expectGrant("expired code", exchange(codeFrom(t, r5), chromeRedir, "trunc-chrome", testVerifier))
}

func TestConnectAuthorizeErrorsGoBackToApp(t *testing.T) {
	s, _, _ := newConnect()
	cases := map[string]func(*AuthorizeParams){
		"access_denied":             func(p *AuthorizeParams) { p.Decision = "deny" },
		"unsupported_response_type": func(p *AuthorizeParams) { p.ResponseType = "token" },
		"invalid_scope":             func(p *AuthorizeParams) { p.Scope = "admin" },
		"invalid_request":           func(p *AuthorizeParams) { p.CodeChallenge = ""; p.CodeChallengeMethod = "" },
	}
	for want, mut := range cases {
		p := allow("trunc-wordpress", "https://blog.example.com/wp-admin/admin.php?page=trunc")
		mut(&p)
		redirect, err := s.Authorize(7, p)
		if err != nil {
			t.Fatalf("%s: %v", want, err)
		}
		u, _ := url.Parse(redirect)
		if u.Query().Get("error") != want || u.Query().Get("code") != "" || u.Query().Get("state") != "st-1" {
			t.Errorf("%s: got %s", want, redirect)
		}
		if u.Query().Get("page") != "trunc" {
			t.Errorf("%s: existing query lost: %s", want, redirect)
		}
	}
	// A plain method is refused for public clients.
	p := allow("trunc-shopify", shopifyRedir)
	p.CodeChallengeMethod = "plain"
	r, _ := s.Authorize(7, p)
	if !strings.Contains(r, "error=invalid_request") {
		t.Errorf("plain PKCE must be refused: %s", r)
	}
	// An untrusted redirect is never followed.
	if _, err := s.Authorize(7, allow("trunc-chrome", "https://evil.example.com/")); err == nil {
		t.Error("authorize must fail without redirecting to an unregistered URI")
	}
}

func TestConnectCanvaNeedsSecret(t *testing.T) {
	s, _, _ := newConnect()
	p := allow("trunc-canva", CanvaRedirectURI)
	p.CodeChallenge, p.CodeChallengeMethod = "", "" // confidential client, PKCE optional
	r, err := s.Authorize(9, p)
	if err != nil {
		t.Fatal(err)
	}
	code := codeFrom(t, r)
	if _, err := s.Exchange(TokenParams{GrantType: "authorization_code", Code: code, RedirectURI: CanvaRedirectURI, ClientID: "trunc-canva", ClientSecret: "wrong"}); err == nil {
		t.Fatal("wrong secret must fail")
	}
	// The failed attempt did not consume the code.
	res, err := s.Exchange(TokenParams{GrantType: "authorization_code", Code: code, RedirectURI: CanvaRedirectURI, ClientID: "trunc-canva", ClientSecret: "canva-secret"})
	if err != nil || res.AccessToken == "" {
		t.Fatalf("exchange with secret: %v", err)
	}
}

func TestConnectRevoke(t *testing.T) {
	s, _, ks := newConnect()
	r, _ := s.Authorize(3, allow("trunc-chrome", chromeRedir))
	res, err := s.Exchange(TokenParams{GrantType: "authorization_code", Code: codeFrom(t, r), RedirectURI: chromeRedir, ClientID: "trunc-chrome", CodeVerifier: testVerifier})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(res.AccessToken, "trunc-chrome", ""); err != nil {
		t.Fatal(err)
	}
	if len(ks.deleted) != 1 {
		t.Fatal("key not deleted")
	}
	if _, err := ks.ValidateAPIKey(res.AccessToken); err == nil {
		t.Fatal("revoked key still valid")
	}
	if err := s.Revoke("unknown-token", "trunc-chrome", ""); err != nil {
		t.Fatalf("unknown token must succeed silently: %v", err)
	}
	if err := s.Revoke("x", "trunc-canva", "nope"); !errors.Is(err, ErrConnectClient) {
		t.Fatalf("canva revoke without secret: %v", err)
	}
}
