package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginHandler(t *testing.T) {
	// Switch Gin to test mode to keep logs clean
	gin.SetMode(gin.TestMode)

	// Setup a mock router
	r := gin.Default()
	r.POST("/login", Login)

	t.Run("Empty Request Body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("{}")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Email and password are required")
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "not-an-email",
			"password": "password123",
		}
		body, _ := json.Marshal(loginData)

		// Use bytes.NewBuffer to convert []byte to an io.Reader
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
