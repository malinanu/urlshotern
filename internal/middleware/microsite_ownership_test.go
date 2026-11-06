package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
)

type fakeOwnerLookup map[int64]int64

var errFakeUnavailable = errors.New("unavailable")

func (f fakeOwnerLookup) GetOwnerID(id int64) (int64, error) {
	if id == 99 {
		return 0, errFakeUnavailable
	}
	owner, ok := f[id]
	if !ok {
		return 0, errors.New("not found")
	}
	return owner, nil
}

func TestRequireMicrositeOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lookup := fakeOwnerLookup{1: 10, 2: 20}

	cases := []struct {
		name   string
		path   string
		userID int64
		admin  bool
		want   int
	}{
		{"no user", "/p/1", 0, false, http.StatusUnauthorized},
		{"bad id", "/p/abc", 10, false, http.StatusBadRequest},
		{"missing page", "/p/5", 10, false, http.StatusNotFound},
		{"other owner", "/p/2", 10, false, http.StatusForbidden},
		{"owner", "/p/1", 10, false, http.StatusOK},
		{"admin bypass", "/p/2", 10, true, http.StatusOK},
		{"unavailable", "/p/99", 10, false, http.StatusServiceUnavailable},
	}
	for _, tc := range cases {
		r := gin.New()
		r.GET("/p/:id", func(c *gin.Context) {
			if tc.userID != 0 {
				c.Set("user_id", tc.userID)
				c.Set("user", &models.User{IsAdmin: tc.admin})
			}
			c.Next()
		}, RequireMicrositeOwnership(lookup, errFakeUnavailable), func(c *gin.Context) {
			if c.GetInt64("microsite_id") == 0 {
				t.Errorf("%s: microsite_id not set", tc.name)
			}
			c.Status(http.StatusOK)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.name, w.Code, tc.want)
		}
	}
}
