package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	hibp "github.com/marco-montesines/haveibeenpwned"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHandler wires the HTTP API to a mock HIBP upstream.
func newTestHandler(t *testing.T, upstream http.Handler) http.Handler {
	t.Helper()
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := hibp.New("test-key", hibp.WithBaseURL(u), hibp.WithPwnedPasswordsBaseURL(u))
	return apiHandler(client)
}

func TestServe_Healthz(t *testing.T) {
	handler := newTestHandler(t, http.NewServeMux())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestServe_Breaches(t *testing.T) {
	upstream := http.NewServeMux()
	upstream.HandleFunc("/breaches", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "adobe.com", r.URL.Query().Get("domain"))
		fmt.Fprint(w, `[{"Name":"Adobe"}]`)
	})
	handler := newTestHandler(t, upstream)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/breaches?domain=adobe.com", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var breaches []hibp.Breach
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &breaches))
	require.Len(t, breaches, 1)
	assert.Equal(t, "Adobe", breaches[0].Name)
}

func TestServe_BreachedAccount_NotFoundMeansEmpty(t *testing.T) {
	upstream := http.NewServeMux()
	upstream.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	handler := newTestHandler(t, upstream)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/breachedaccount/clean@example.com", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `[]`, rec.Body.String())
}

func TestServe_PwnedPassword(t *testing.T) {
	upstream := http.NewServeMux()
	// SHA-1("password") = 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
	upstream.HandleFunc("/range/5BAA6", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "1E4C9B93F3F0682250B6CF8331B7EE68FD8:10437277\r\n")
	})
	handler := newTestHandler(t, upstream)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/pwnedpassword", strings.NewReader(`{"password":"password"}`))
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"pwned":true,"count":10437277}`, rec.Body.String())
}

func TestServe_PwnedPassword_BadRequest(t *testing.T) {
	handler := newTestHandler(t, http.NewServeMux())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/pwnedpassword", strings.NewReader(`{}`))
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestServe_UpstreamErrorMapped(t *testing.T) {
	upstream := http.NewServeMux()
	upstream.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("retry-after", "5")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"statusCode":429,"message":"Rate limit is exceeded."}`)
	})
	handler := newTestHandler(t, upstream)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/breachedaccount/info@example.com", nil))

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "5", rec.Header().Get("Retry-After"))
}
