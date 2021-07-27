package haveibeenpwned

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

var (
	expectedMethod  = "GET"
	expectedHeaders = map[string]string{
		"user-agent":   UserAgent,
		"hibp-api-key": "1234",
	}

	mockHandler *http.ServeMux
	mockServer  *httptest.Server
)

func init() {
	mockHandler = http.NewServeMux()
	mockServer = httptest.NewServer(mockHandler)
	baseURL, _ = url.Parse(mockServer.URL)
}

func checkHeader(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Fatalf("Expected %s for request method, got %s", expectedMethod, r.Method)
		}

		for k, v := range expectedHeaders {
			header := r.Header.Get(k)
			if header != v {
				t.Fatalf("Expected %s for request header, got %s", v, header)
			}
		}
	}
}

func TestHaveIBeenPwned_GetBreachedAccount(t *testing.T) {
	assert := assert.New(t)

	mockHandler.HandleFunc("/breachedaccount/info@example.com", func(w http.ResponseWriter, r *http.Request) {
		checkHeader(t)(w, r)
		fmt.Fprint(w, `[{"Name":"Adobe"},{"Name":"Apollo"},{"Name":"LeadHunter"},{"Name":"VerificationsIO"}]`)
	})

	ctx := context.Background()
	hibp := New("1234")

	account, err := hibp.GetBreachedAccount(ctx, "info@example.com", "",
		true, false)
	if err != nil {
		t.Fatalf("[Get Breaches For Account] returned error: %v", err)
	}

	want := []Breach{
		{Name: "Adobe"},
		{Name: "Apollo"},
		{Name: "LeadHunter"},
		{Name: "VerificationsIO"},
	}
	assert.Equal(want, account, "Expected equal value")
}

func TestHaveIBeenPwned_GetBreaches(t *testing.T) {
	assert := assert.New(t)
	expectedHeaders = map[string]string{
		"user-agent": UserAgent,
	}
	mockHandler.HandleFunc("/breaches", func(w http.ResponseWriter, r *http.Request) {
		checkHeader(t)(w, r)
		fmt.Fprint(w, `[{"Name": "Adobe"}]`)
	})

	ctx := context.Background()
	hibp := New("1234")

	account, err := hibp.GetBreaches(ctx, "adobe.com")
	if err != nil {
		t.Fatalf("[Get Breached Sites] returned error: %v", err)
	}

	want := []*Breach{
		{Name: "Adobe"},
	}
	assert.Equal(want, account, "they should be the same output.")
}

func TestHaveIBeenPwned_GetBreachSite(t *testing.T) {
	assert := assert.New(t)
	expectedHeaders = map[string]string{
		"user-agent": UserAgent,
	}
	mockHandler.HandleFunc("/breach/Adobe", func(w http.ResponseWriter, r *http.Request) {
		checkHeader(t)(w, r)
		fmt.Fprint(w, `{"Name": "Adobe"}`)
	})

	ctx := context.Background()
	hibp := New("1234")

	account, err := hibp.GetBreachedSite(ctx, "Adobe")
	if err != nil {
		t.Fatalf("[Get Breached Site] returned error: %v", err)
	}

	want := &Breach{
		Name: "Adobe",
	}
	assert.Equal(want, account, "they should be the same output.")
}

func TestHaveIBeenPwned_GetDataClasses(t *testing.T) {
	assert := assert.New(t)
	expectedHeaders = map[string]string{
		"user-agent": UserAgent,
	}
	mockHandler.HandleFunc("/dataclasses", func(w http.ResponseWriter, r *http.Request) {
		checkHeader(t)(w, r)
		fmt.Fprint(w, `["Account balances","Age groups"]`)
	})

	ctx := context.Background()
	hibp := New("1234")

	account, err := hibp.GetDataClasses(ctx)
	if err != nil {
		t.Fatalf("[Get All Data Classes] returned error: %v", err)
	}

	want := &DataClasses{
		"Account balances",
		"Age groups",
	}
	assert.Equal(want, account, "Expected equal value")
}

func TestHaveIBeenPwned_GetPastedAccount(t *testing.T) {
	assert := assert.New(t)

	mockHandler.HandleFunc("/pasteaccount/info@example.com", func(w http.ResponseWriter, r *http.Request) {
		checkHeader(t)(w, r)
		fmt.Fprint(w, `[{"Source":"Pastebin","Id":"Ab2ZYrq4","EmailCount":48},{"Source":"Pastebin","Id":"46g62dvD","EmailCount":1670}]`)
	})

	ctx := context.Background()
	hibp := New("1234")

	account, err := hibp.GetPastedAccount(ctx, "info@example.com")
	if err != nil {
		t.Fatalf("[Get Pastes For Account] returned error: %v", err)
	}

	want := []*Paste{
		{
			Source:     "Pastebin",
			Id:         "Ab2ZYrq4",
			EmailCount: 48,
		},
		{
			Source:     "Pastebin",
			Id:         "46g62dvD",
			EmailCount: 1670,
		},
	}
	assert.Equal(want, account, "Expected equal value")
}
