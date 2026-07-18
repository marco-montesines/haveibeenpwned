package haveibeenpwned

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient returns a client pointed at a mock server that serves the
// given handler, plus the server itself for cleanup.
func newTestClient(t *testing.T, handler http.Handler) *HaveIBeenPwned {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	return New("1234", WithBaseURL(u), WithPwnedPasswordsBaseURL(u))
}

// checkHeaders fails the request handler if method or headers do not match.
func checkHeaders(t *testing.T, r *http.Request, expected map[string]string) {
	t.Helper()
	assert.Equal(t, http.MethodGet, r.Method, "unexpected request method")
	for k, v := range expected {
		assert.Equal(t, v, r.Header.Get(k), "unexpected value for header %q", k)
	}
}

func TestHaveIBeenPwned_GetBreachedAccount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breachedaccount/info@example.com", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{
			"user-agent":   UserAgent,
			"hibp-api-key": "1234",
		})
		assert.Equal(t, "true", r.URL.Query().Get("truncateResponse"))
		assert.Equal(t, "false", r.URL.Query().Get("includeUnverified"))
		fmt.Fprint(w, `[{"Name":"Adobe"},{"Name":"Apollo"},{"Name":"LeadHunter"},{"Name":"VerificationsIO"}]`)
	})
	hibp := newTestClient(t, mux)

	account, err := hibp.GetBreachedAccount(context.Background(), "info@example.com", "", true, false)
	require.NoError(t, err)

	want := []Breach{
		{Name: "Adobe"},
		{Name: "Apollo"},
		{Name: "LeadHunter"},
		{Name: "VerificationsIO"},
	}
	assert.Equal(t, want, account)
}

func TestHaveIBeenPwned_GetBreachedAccount_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breachedaccount/clean@example.com", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	hibp := newTestClient(t, mux)

	account, err := hibp.GetBreachedAccount(context.Background(), "clean@example.com", "", true, false)
	require.NoError(t, err, "a 404 means the account is not breached, not an error")
	assert.Nil(t, account)
}

func TestHaveIBeenPwned_GetBreachedAccount_InvalidEmail(t *testing.T) {
	hibp := New("1234")
	for _, email := range []string{"", "short", "no-at-sign.example.com", "two words@example.com"} {
		_, err := hibp.GetBreachedAccount(context.Background(), email, "", true, false)
		assert.Error(t, err, "expected validation error for %q", email)
	}
}

func TestHaveIBeenPwned_GetBreaches(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breaches", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{"user-agent": UserAgent})
		assert.Empty(t, r.Header.Get("hibp-api-key"), "unauthenticated endpoint must not send the API key")
		assert.Equal(t, "adobe.com", r.URL.Query().Get("domain"))
		fmt.Fprint(w, `[{"Name": "Adobe"}]`)
	})
	hibp := newTestClient(t, mux)

	breaches, err := hibp.GetBreaches(context.Background(), "adobe.com")
	require.NoError(t, err)
	assert.Equal(t, []*Breach{{Name: "Adobe"}}, breaches)
}

func TestHaveIBeenPwned_GetBreachedSite(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breach/Adobe", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{"user-agent": UserAgent})
		fmt.Fprint(w, `{"Name": "Adobe"}`)
	})
	hibp := newTestClient(t, mux)

	breach, err := hibp.GetBreachedSite(context.Background(), "Adobe")
	require.NoError(t, err)
	assert.Equal(t, &Breach{Name: "Adobe"}, breach)
}

func TestHaveIBeenPwned_GetDataClasses(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dataclasses", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{"user-agent": UserAgent})
		fmt.Fprint(w, `["Account balances","Age groups"]`)
	})
	hibp := newTestClient(t, mux)

	dataClasses, err := hibp.GetDataClasses(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &DataClasses{"Account balances", "Age groups"}, dataClasses)
}

func TestHaveIBeenPwned_GetPastedAccount(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pasteaccount/info@example.com", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{
			"user-agent":   UserAgent,
			"hibp-api-key": "1234",
		})
		fmt.Fprint(w, `[{"Source":"Pastebin","Id":"Ab2ZYrq4","EmailCount":48},{"Source":"Pastebin","Id":"46g62dvD","EmailCount":1670}]`)
	})
	hibp := newTestClient(t, mux)

	pastes, err := hibp.GetPastedAccount(context.Background(), "info@example.com")
	require.NoError(t, err)

	want := []*Paste{
		{Source: "Pastebin", Id: "Ab2ZYrq4", EmailCount: 48},
		{Source: "Pastebin", Id: "46g62dvD", EmailCount: 1670},
	}
	assert.Equal(t, want, pastes)
}

func TestHaveIBeenPwned_PwnedPasswordCount(t *testing.T) {
	// SHA-1("password") = 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
	mux := http.NewServeMux()
	mux.HandleFunc("/range/5BAA6", func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r, map[string]string{
			"user-agent":  UserAgent,
			"Add-Padding": "true",
		})
		fmt.Fprint(w, "0018A45C4D1DEF81644B54AB7F969B88D65:2\r\n"+
			"1E4C9B93F3F0682250B6CF8331B7EE68FD8:10437277\r\n"+
			"FED23A38AF2AC2B0A20BAF80012DC6E2726:0\r\n")
	})
	hibp := newTestClient(t, mux)

	count, err := hibp.PwnedPasswordCount(context.Background(), "password")
	require.NoError(t, err)
	assert.Equal(t, 10437277, count)
}

func TestHaveIBeenPwned_PwnedPasswordCount_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "0018A45C4D1DEF81644B54AB7F969B88D65:2\r\n")
	})
	hibp := newTestClient(t, mux)

	count, err := hibp.PwnedPasswordCount(context.Background(), "password")
	require.NoError(t, err)
	assert.Zero(t, count, "a suffix miss must report a count of zero")
}

func TestHaveIBeenPwned_PwnedPasswordCount_EmptyPassword(t *testing.T) {
	hibp := New("")
	_, err := hibp.PwnedPasswordCount(context.Background(), "")
	assert.Error(t, err)
}

func TestHaveIBeenPwned_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breachedaccount/info@example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("retry-after", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"statusCode":429,"message":"Rate limit is exceeded. Try again in 7 seconds."}`)
	})
	hibp := newTestClient(t, mux)

	_, err := hibp.GetBreachedAccount(context.Background(), "info@example.com", "", true, false)
	require.Error(t, err)

	var apiErr *HIBPErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.Code)
	assert.Equal(t, 7, apiErr.RetryAfter)
	assert.Contains(t, apiErr.Message, "Rate limit")
}

func TestHaveIBeenPwned_APIError_NonJSONBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/breaches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "upstream unavailable")
	})
	hibp := newTestClient(t, mux)

	_, err := hibp.GetBreaches(context.Background(), "")
	require.Error(t, err)

	var apiErr *HIBPErrorResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusServiceUnavailable, apiErr.Code)
}

func TestValidateEmail(t *testing.T) {
	hibp := New("")

	valid := []string{
		"info@example.com",
		"first.last@sub.example.co.uk",
		"user+tag@example.com",
		"info@münchen.de", // IDN domains are punycode-checked
	}
	for _, email := range valid {
		assert.NoError(t, hibp.validateEmail(email), "expected %q to be valid", email)
	}

	invalid := []string{
		"",
		"a@b.c",
		"plainaddress",
		"missing-domain@",
		"@missing-local.com",
	}
	for _, email := range invalid {
		assert.Error(t, hibp.validateEmail(email), "expected %q to be invalid", email)
	}
}
