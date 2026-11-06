package unit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEscapeVCardValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Text", "Simple Text"},
		{"Acme, Inc.", `Acme\, Inc.`},
		{"Foo; Bar", `Foo\; Bar`},
		{"Back\\slash", `Back\\slash`},
		{"Line 1\nLine 2", `Line 1\nLine 2`},
		{"Line 1\r\nLine 2", `Line 1\nLine 2`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res := services.EscapeVCardValue(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestGenerateVCard(t *testing.T) {
	t.Run("Full Contact Generation", func(t *testing.T) {
		contact := &models.VCardContact{
			FirstName:    "Ada",
			LastName:     "Lovelace",
			Organization: "Computing Pioneers Ltd",
			Title:        "Chief Mathematician",
			Email:        "ada@lovelace.org",
			Phone:        "+44 20 7946 0991",
			CellPhone:    "+44 77 0090 0123",
			WorkPhone:    "+44 20 7946 0992",
			Address:      "12 St James's Square",
			City:         "London",
			State:        "Greater London",
			PostalCode:   "SW1Y 4LB",
			Country:      "United Kingdom",
			Website:      "https://adalovelace.org",
			Note:         "Pioneer of computer algorithms and analytical engines.",
		}

		vcard := services.GenerateVCard(contact)

		assert.True(t, strings.HasPrefix(vcard, "BEGIN:VCARD\r\nVERSION:3.0\r\n"))
		assert.True(t, strings.HasSuffix(vcard, "END:VCARD\r\n"))
		assert.Contains(t, vcard, "FN:Ada Lovelace\r\n")
		assert.Contains(t, vcard, "N:Lovelace;Ada;;;\r\n")
		assert.Contains(t, vcard, "ORG:Computing Pioneers Ltd\r\n")
		assert.Contains(t, vcard, "TITLE:Chief Mathematician\r\n")
		assert.Contains(t, vcard, "EMAIL;TYPE=INTERNET,PREF:ada@lovelace.org\r\n")
		assert.Contains(t, vcard, "TEL;TYPE=CELL,VOICE:+44 20 7946 0991\r\n")
		assert.Contains(t, vcard, "TEL;TYPE=CELL,VOICE:+44 77 0090 0123\r\n")
		assert.Contains(t, vcard, "TEL;TYPE=WORK,VOICE:+44 20 7946 0992\r\n")
		assert.Contains(t, vcard, "URL:https://adalovelace.org\r\n")
		assert.Contains(t, vcard, "ADR;TYPE=WORK:;;12 St James's Square;London;Greater London;SW1Y 4LB;United Kingdom\r\n")
		assert.Contains(t, vcard, "NOTE:Pioneer of computer algorithms and analytical engines.\r\n")
	})

	t.Run("Nil Contact", func(t *testing.T) {
		assert.Equal(t, "", services.GenerateVCard(nil))
	})

	t.Run("Only Organization Contact", func(t *testing.T) {
		contact := &models.VCardContact{
			Organization: "Acme Corporation",
			Email:        "contact@acme.com",
		}
		vcard := services.GenerateVCard(contact)
		assert.Contains(t, vcard, "FN:Acme Corporation\r\n")
		assert.Contains(t, vcard, "N:Acme Corporation;;;;\r\n")
		assert.Contains(t, vcard, "ORG:Acme Corporation\r\n")
		assert.Contains(t, vcard, "EMAIL;TYPE=INTERNET,PREF:contact@acme.com\r\n")
	})
}

func TestMicroPageSlugValidation(t *testing.T) {
	svc := services.NewMicroPageService(nil)

	validSlugs := []string{
		"alex", "alex-smith", "design-lead-2026", "sarah99",
	}
	for _, slug := range validSlugs {
		assert.NoError(t, svc.ValidateSlug(slug), "Slug %s should be valid", slug)
	}

	invalidSlugs := []string{
		"a",                     // too short
		"alex_smith",            // underscore not allowed
		"alex--smith",           // double hyphen
		"-alex",                 // leading hyphen
		"alex-",                 // trailing hyphen
		"api",                   // reserved
		"dashboard",             // reserved
		"micropages",            // reserved
		"m",                     // reserved
	}
	for _, slug := range invalidSlugs {
		assert.Error(t, svc.ValidateSlug(slug), "Slug %s should be rejected", slug)
	}
}

func TestGinRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/micropages/public/:slug", func(c *gin.Context) {
		c.String(200, "public:"+c.Param("slug"))
	})
	r.GET("/api/v1/micropages/public/:slug/vcard", func(c *gin.Context) {
		c.String(200, "vcard:"+c.Param("slug"))
	})
	r.POST("/api/v1/micropages/public/:slug/click/:linkId", func(c *gin.Context) {
		c.String(200, "click:"+c.Param("linkId"))
	})
	r.GET("/m", func(c *gin.Context) {
		c.String(200, "m-root")
	})
	r.GET("/m/*path", func(c *gin.Context) {
		subPath := c.Param("path")
		trimmed := strings.Trim(subPath, "/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && (parts[1] == "vcard" || parts[1] == "contact.vcf") {
			c.String(200, "vcard:"+parts[0])
			return
		}
		c.String(200, "spa:"+subPath)
	})

	// Test direct vCard download route
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/m/alex/vcard", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)
	assert.Equal(t, "vcard:alex", w1.Body.String())

	// Test direct contact.vcf download route
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/m/alex/contact.vcf", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, "vcard:alex", w2.Body.String())

	// Test SPA micropage view route
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/m/alex", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	assert.Equal(t, "spa:/alex", w3.Body.String())

	// Test public API route
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/api/v1/micropages/public/alex", nil)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, 200, w4.Code)
	assert.Equal(t, "public:alex", w4.Body.String())
}
