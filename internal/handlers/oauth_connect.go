package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/services"
)

// ConnectHandlers serves "Sign in with Trunc" for the plugins.
type ConnectHandlers struct {
	svc *services.ConnectService
}

func NewConnectHandlers(svc *services.ConnectService) *ConnectHandlers {
	return &ConnectHandlers{svc: svc}
}

func writeConnectError(c *gin.Context, err error) {
	var ce *services.ConnectError
	if errors.As(err, &ce) {
		c.JSON(ce.Status, gin.H{"error": ce.Code, "error_description": ce.Description})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "error_description": "Something went wrong."})
}

// GetClient describes an app for the consent page.
// GET /api/v1/oauth/client?client_id=&redirect_uri=
func (h *ConnectHandlers) GetClient(c *gin.Context) {
	client, u, err := h.svc.Client(c.Query("client_id"), c.Query("redirect_uri"))
	if err != nil {
		writeConnectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"client_id": client.ID,
		"name":      client.Name,
		"host":      u.Hostname(),
		"any_host":  client.AnyHost,
		"scope":     services.ConnectScope,
	})
}

// Authorize records the signed-in user's decision and returns the redirect.
// POST /api/v1/oauth/authorize (JWT only: an API key cannot approve new keys)
func (h *ConnectHandlers) Authorize(c *gin.Context) {
	if t, _ := c.Get("auth_type"); t == "api_key" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "error_description": "Sign in to trunc.site to connect an app."})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login_required", "error_description": "Sign in to continue."})
		return
	}
	var p services.AuthorizeParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid request."})
		return
	}
	redirect, err := h.svc.Authorize(userID, p)
	if err != nil {
		writeConnectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect": redirect})
}

// connectParams reads the OAuth parameters from a form or JSON body.
func connectParams(c *gin.Context) map[string]string {
	out := map[string]string{}
	if strings.HasPrefix(c.ContentType(), "application/json") {
		var m map[string]interface{}
		if err := c.ShouldBindJSON(&m); err == nil {
			for k, v := range m {
				if s, ok := v.(string); ok {
					out[k] = s
				}
			}
		}
		return out
	}
	if err := c.Request.ParseForm(); err == nil {
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				out[k] = v[0]
			}
		}
	}
	return out
}

// clientCredentials reads client_secret_basic, falling back to the body.
func clientCredentials(c *gin.Context, p map[string]string) (string, string) {
	if id, secret, ok := c.Request.BasicAuth(); ok {
		return id, secret
	}
	return p["client_id"], p["client_secret"]
}

// Token exchanges a code for an API key.
// POST /api/v1/oauth/token
func (h *ConnectHandlers) Token(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	p := connectParams(c)
	id, secret := clientCredentials(c, p)
	res, err := h.svc.Exchange(services.TokenParams{
		GrantType:    p["grant_type"],
		Code:         p["code"],
		RedirectURI:  p["redirect_uri"],
		ClientID:     id,
		ClientSecret: secret,
		CodeVerifier: p["code_verifier"],
	})
	if err != nil {
		writeConnectError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Revoke deletes a key issued to an app (RFC 7009).
// POST /api/v1/oauth/revoke
func (h *ConnectHandlers) Revoke(c *gin.Context) {
	p := connectParams(c)
	id, secret := clientCredentials(c, p)
	if err := h.svc.Revoke(p["token"], id, secret); err != nil {
		if errors.Is(err, services.ErrConnectClient) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client", "error_description": "Client authentication failed."})
			return
		}
		writeConnectError(c, err)
		return
	}
	c.Status(http.StatusOK)
}
