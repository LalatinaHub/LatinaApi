package httputil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetJSON(t *testing.T) {
	type Resp struct {
		Message string `json:"message"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "LatinaApi/1.0", r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resp{Message: "pong"})
	}))
	defer ts.Close()

	client := NewClient(2 * time.Second)
	var resp Resp
	err := client.GetJSON(context.Background(), ts.URL, &resp)
	require.NoError(t, err)
	assert.Equal(t, "pong", resp.Message)
}

func TestClient_GetWithRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	client := NewClient(2 * time.Second)
	resp, err := client.GetWithRetry(context.Background(), ts.URL, 3, 10*time.Millisecond)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts)
}
