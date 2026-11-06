package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupStaticRoutesNoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Should not panic with wildcard conflict
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("setupStaticRoutes panicked: %v", r)
		}
	}()

	setupStaticRoutes(router, nil, "/microsites")
}
