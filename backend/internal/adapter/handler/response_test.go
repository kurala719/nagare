package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nagare/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type apiResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func TestRespondErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		err      error
		status   int
		errorMsg string
	}{
		{domain.ErrNotFound, http.StatusNotFound, "resource not found"},
		{domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{domain.ErrTimeout, http.StatusGatewayTimeout, "operation timed out"},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		respondError(c, tc.err)

		if w.Code != tc.status {
			t.Fatalf("expected status %d, got %d", tc.status, w.Code)
		}

		var body apiResponse
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}
		if body.Success {
			t.Fatalf("expected success=false")
		}
		if body.Error != tc.errorMsg {
			t.Fatalf("expected error %q, got %q", tc.errorMsg, body.Error)
		}
	}
}
