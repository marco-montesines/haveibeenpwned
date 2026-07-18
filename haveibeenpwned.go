// Package haveibeenpwned is a client for Troy Hunt's Have I Been Pwned
// API v3 (https://haveibeenpwned.com/API/v3), including the Pwned
// Passwords k-anonymity range API (https://api.pwnedpasswords.com).
//
// Endpoints that query data for a specific account (GetBreachedAccount,
// GetPastedAccount) require an API key from https://haveibeenpwned.com/API/Key.
// All other endpoints, including PwnedPasswordCount, work without a key.
package haveibeenpwned

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

const (
	// UserAgent is the default User-Agent header sent with every request.
	// The HIBP API rejects requests without a User-Agent.
	UserAgent = "haveibeenpwned-go/1.0 (+https://github.com/marco-montesines/haveibeenpwned)"
	// Accept is the media type requested from the API.
	Accept = "application/json"
	// Endpoint is the default base URL of the HIBP v3 API.
	Endpoint = "https://haveibeenpwned.com/api/v3/"
	// PwnedPasswordsEndpoint is the default base URL of the Pwned Passwords range API.
	PwnedPasswordsEndpoint = "https://api.pwnedpasswords.com/"
	// Domain matches a syntactically valid ASCII domain name.
	Domain = `^(?:[_\-a-z0-9]+\.)*([\-a-z0-9]+\.)[\-a-z0-9]{2,63}$`
	// DomainUnicode matches a syntactically valid internationalized domain name.
	DomainUnicode = `^(?:[_\-\p{L}\d]+\.)*([\-\p{L}\d]+\.)[\-\p{L}\d]{2,63}$`
)

var rDomain = regexp.MustCompile(Domain)

// HaveIBeenPwned is a client for the HIBP v3 and Pwned Passwords APIs.
// It is safe for concurrent use by multiple goroutines.
type HaveIBeenPwned struct {
	accessKey    string
	client       *http.Client
	userAgent    string
	baseURL      *url.URL
	passwordsURL *url.URL
}

// Option configures a HaveIBeenPwned client.
type Option func(*HaveIBeenPwned)

// WithHTTPClient replaces the default *http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *HaveIBeenPwned) { o.client = client }
}

// WithUserAgent replaces the default User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(o *HaveIBeenPwned) { o.userAgent = userAgent }
}

// WithBaseURL replaces the HIBP v3 API base URL, e.g. for tests or proxies.
func WithBaseURL(u *url.URL) Option {
	return func(o *HaveIBeenPwned) { o.baseURL = u }
}

// WithPwnedPasswordsBaseURL replaces the Pwned Passwords API base URL.
func WithPwnedPasswordsBaseURL(u *url.URL) Option {
	return func(o *HaveIBeenPwned) { o.passwordsURL = u }
}

// New returns a client for the HIBP v3 API. The accessKey may be empty if
// only endpoints that do not require authentication are used.
func New(accessKey string, opts ...Option) *HaveIBeenPwned {
	baseURL, _ := url.Parse(Endpoint)
	passwordsURL, _ := url.Parse(PwnedPasswordsEndpoint)
	o := &HaveIBeenPwned{
		accessKey: accessKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 20,
			},
		},
		userAgent:    UserAgent,
		baseURL:      baseURL,
		passwordsURL: passwordsURL,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Breach describes a single breach in the HIBP corpus.
// See https://haveibeenpwned.com/API/v3#BreachModel.
type Breach struct {
	Name               string      `json:"Name,omitempty"`
	Title              string      `json:"Title,omitempty"`
	Domain             string      `json:"Domain,omitempty"`
	BreachDate         string      `json:"BreachDate,omitempty"`
	AddedDate          time.Time   `json:"AddedDate,omitempty"`
	ModifiedDate       time.Time   `json:"ModifiedDate,omitempty"`
	PwnCount           int         `json:"PwnCount,omitempty"`
	Description        string      `json:"Description,omitempty"`
	LogoPath           string      `json:"LogoPath,omitempty"`
	DataClasses        DataClasses `json:"DataClasses,omitempty"`
	IsVerified         bool        `json:"IsVerified,omitempty"`
	IsFabricated       bool        `json:"IsFabricated,omitempty"`
	IsSensitive        bool        `json:"IsSensitive,omitempty"`
	IsRetired          bool        `json:"IsRetired,omitempty"`
	IsSpamList         bool        `json:"IsSpamList,omitempty"`
	IsMalware          bool        `json:"IsMalware,omitempty"`
	IsSubscriptionFree bool        `json:"IsSubscriptionFree,omitempty"`
	IsStealerLog       bool        `json:"IsStealerLog,omitempty"`
}

// Paste describes a single paste containing a searched account.
// See https://haveibeenpwned.com/API/v3#PasteModel.
type Paste struct {
	Source     string    `json:"Source,omitempty"`
	Id         string    `json:"Id,omitempty"`
	Title      string    `json:"Title,omitempty"`
	Date       time.Time `json:"Date,omitempty"`
	EmailCount int       `json:"EmailCount,omitempty"`
}

// DataClasses is the list of data attributes compromised in a breach.
type DataClasses []string

// GetBreachedAccount returns all breaches an account appears in. Requires an
// API key. A nil slice with a nil error means the account was not found in
// any breach. See https://haveibeenpwned.com/API/v3#BreachesForAccount.
func (o *HaveIBeenPwned) GetBreachedAccount(ctx context.Context, email string, domain string,
	truncateResponse bool, includeUnverified bool) ([]Breach, error) {
	if err := o.validateEmail(email); err != nil {
		return nil, err
	}
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	params.Set("truncateResponse", strconv.FormatBool(truncateResponse))
	params.Set("includeUnverified", strconv.FormatBool(includeUnverified))

	resp, err := o.request(ctx, "breachedaccount/"+url.PathEscape(email), params, true)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, o.formatError(resp)
	}

	var breaches []Breach
	return breaches, json.NewDecoder(resp.Body).Decode(&breaches)
}

// GetBreaches returns all breaches in the system, optionally filtered by the
// domain the breach occurred on. No API key required.
// See https://haveibeenpwned.com/API/v3#AllBreaches.
func (o *HaveIBeenPwned) GetBreaches(ctx context.Context, domain string) ([]*Breach, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	resp, err := o.request(ctx, "breaches", params, false)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, o.formatError(resp)
	}

	var breaches []*Breach
	return breaches, json.NewDecoder(resp.Body).Decode(&breaches)
}

// GetBreachedSite returns a single breach by its name (the stable "Name"
// attribute, e.g. "Adobe"). No API key required.
// See https://haveibeenpwned.com/API/v3#SingleBreach.
func (o *HaveIBeenPwned) GetBreachedSite(ctx context.Context, site string) (*Breach, error) {
	resp, err := o.request(ctx, "breach/"+url.PathEscape(site), nil, false)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, o.formatError(resp)
	}

	var breach *Breach
	return breach, json.NewDecoder(resp.Body).Decode(&breach)
}

// GetDataClasses returns all data classes in the system. No API key required.
// See https://haveibeenpwned.com/API/v3#AllDataClasses.
func (o *HaveIBeenPwned) GetDataClasses(ctx context.Context) (*DataClasses, error) {
	resp, err := o.request(ctx, "dataclasses", nil, false)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, o.formatError(resp)
	}

	var dataClasses *DataClasses
	return dataClasses, json.NewDecoder(resp.Body).Decode(&dataClasses)
}

// GetPastedAccount returns all pastes an account appears in. Requires an API
// key. A nil slice with a nil error means the account was not found in any
// paste. See https://haveibeenpwned.com/API/v3#PastesForAccount.
func (o *HaveIBeenPwned) GetPastedAccount(ctx context.Context, email string) ([]*Paste, error) {
	if err := o.validateEmail(email); err != nil {
		return nil, err
	}
	resp, err := o.request(ctx, "pasteaccount/"+url.PathEscape(email), nil, true)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, o.formatError(resp)
	}

	var pastes []*Paste
	return pastes, json.NewDecoder(resp.Body).Decode(&pastes)
}

// PwnedPasswordCount reports how many times a password appears in the Pwned
// Passwords corpus, using the k-anonymity range API: only the first five
// characters of the password's SHA-1 hash ever leave the process. A count of
// zero means the password was not found. No API key required.
// See https://haveibeenpwned.com/API/v3#PwnedPasswords.
func (o *HaveIBeenPwned) PwnedPasswordCount(ctx context.Context, password string) (int, error) {
	if password == "" {
		return 0, errors.New("haveibeenpwned: password must not be empty")
	}
	sum := sha1.Sum([]byte(password))
	hash := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix, suffix := hash[:5], hash[5:]

	target := strings.TrimSuffix(o.passwordsURL.String(), "/") + "/range/" + prefix
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("user-agent", o.userAgent)
	// Ask the API to pad responses so the response size does not leak
	// information about the queried range.
	req.Header.Set("Add-Padding", "true")

	resp, err := o.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("haveibeenpwned: pwned passwords range query failed: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rest, ok := strings.CutPrefix(line, suffix+":")
		if !ok {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil {
			return 0, fmt.Errorf("haveibeenpwned: malformed range response line %q: %w", line, err)
		}
		return count, nil
	}
	return 0, scanner.Err()
}

func (o *HaveIBeenPwned) request(ctx context.Context, resource string, params url.Values, authenticated bool) (*http.Response, error) {
	target := strings.TrimSuffix(o.baseURL.String(), "/") + "/" + resource
	if len(params) > 0 {
		target += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", Accept)
	req.Header.Set("user-agent", o.userAgent)
	if authenticated {
		req.Header.Set("hibp-api-key", o.accessKey)
	}

	return o.client.Do(req)
}

func (o *HaveIBeenPwned) formatError(r *http.Response) *HIBPErrorResponse {
	var ans *HIBPErrorResponse
	if err := json.NewDecoder(r.Body).Decode(&ans); err != nil || ans == nil {
		ans = &HIBPErrorResponse{Code: r.StatusCode}
	}
	if ans.Code == 0 {
		ans.Code = r.StatusCode
	}
	ans.Description = http.StatusText(r.StatusCode)
	if retryAfter, _ := strconv.Atoi(r.Header.Get("retry-after")); retryAfter > 0 {
		ans.RetryAfter = retryAfter
	}
	return ans
}

func (o *HaveIBeenPwned) validateEmail(value string) error {
	if len(value) < 6 || len(value) > 150 {
		return errors.New("email length must be from 6 to 150")
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return errors.New("email not valid")
	}
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return errors.New("email cannot split to local and domain parts")
	}
	local := parts[0]
	if len(local) > 64 {
		return errors.New("length of local part should be max 64 characters")
	}
	if strings.Contains(local, " ") {
		return errors.New("not an email pattern (bad local part)")
	}
	domain := parts[1]
	if !rDomain.MatchString(strings.ToLower(domain)) {
		punyEncoded, err := idna.Punycode.ToASCII(domain)
		if err != nil {
			return errors.New("not an email pattern (bad domain)")
		}
		if !rDomain.MatchString(punyEncoded) {
			return errors.New("not an email pattern (bad domain)")
		}
	}

	return nil
}

func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, body)
	body.Close()
}

// HIBPErrorResponse is the error payload returned by the HIBP API for
// non-2xx responses. It implements the error interface. RetryAfter carries
// the Retry-After header value (in seconds) on 429 responses.
type HIBPErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"-"`
	Code        int    `json:"statusCode"`
	RetryAfter  int    `json:"-"`
}

func (e *HIBPErrorResponse) Error() string {
	msg := fmt.Sprintf("haveibeenpwned: %d %s", e.Code, e.Description)
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if e.RetryAfter > 0 {
		msg += fmt.Sprintf(" (retry after %ds)", e.RetryAfter)
	}
	return msg
}
