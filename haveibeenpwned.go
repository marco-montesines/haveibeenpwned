package haveibeenpwned

import (
	"context"
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
	UserAgent     = "haveibeenpwned-client/v0.1"
	Accept        = "application/json"
	Endpoint      = "https://haveibeenpwned.com/api/v3/"
	Domain        = `^(?:[_\-a-z0-9]+\.)*([\-a-z0-9]+\.)[\-a-z0-9]{2,63}$`
	DomainUnicode = `^(?:[_\-\p{L}\d]+\.)*([\-\p{L}\d]+\.)[\-\p{L}\d]{2,63}$`
	useAccessKey  = "useAccessKey"
)

var (
	rDomain    = regexp.MustCompile(Domain)
	baseURL, _ = url.Parse(Endpoint)
)

type HaveIBeenPwned struct {
	accessKey string
	client    *http.Client
}

func New(accessKey string) *HaveIBeenPwned {
	tr := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 20,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
	ans := HaveIBeenPwned{
		accessKey: accessKey,
		client:    netClient,
	}
	return &ans
}

type Breach struct {
	Name         string      `json:"Name,omitempty"`
	Title        string      `json:"Title,omitempty"`
	Domain       string      `json:"Domain,omitempty"`
	BreachDate   string      `json:"BreachDate,omitempty"`
	AddedDate    time.Time   `json:"AddedDate,omitempty"`
	ModifiedDate time.Time   `json:"ModifiedDate,omitempty"`
	PwnCount     int         `json:"PwnCount,omitempty"`
	Description  string      `json:"Description,omitempty"`
	LogoPath     string      `json:"LogoPath,omitempty"`
	DataClasses  DataClasses `json:"DataClasses,omitempty"`
	IsVerified   bool        `json:"IsVerified,omitempty"`
	IsFabricated bool        `json:"IsFabricated,omitempty"`
	IsSensitive  bool        `json:"IsSensitive,omitempty"`
	IsRetired    bool        `json:"IsRetired,omitempty"`
	IsSpamList   bool        `json:"IsSpamList,omitempty"`
}

type Paste struct {
	Source     string    `json:"Source,omitempty"`
	Id         string    `json:"Id,omitempty"`
	Title      string    `json:"Title,omitempty"`
	Date       time.Time `json:"Date,omitempty"`
	EmailCount int       `json:"EmailCount,omitempty"`
}

type DataClasses []string

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

	resp, err := o.request(ctx, fmt.Sprintf("breachedaccount/%s", email), params, useAccessKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, o.formatError(resp)
	}

	var breaches []Breach
	return breaches, json.NewDecoder(resp.Body).Decode(&breaches)
}

func (o *HaveIBeenPwned) GetBreaches(ctx context.Context, domain string) ([]*Breach, error) {
	params := url.Values{}
	if domain != "" {
		params.Set("domain", domain)
	}
	resp, err := o.request(ctx, "breaches", params)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, o.formatError(resp)
	}

	var breaches []*Breach
	return breaches, json.NewDecoder(resp.Body).Decode(&breaches)
}

func (o *HaveIBeenPwned) GetBreachedSite(ctx context.Context, site string) (*Breach, error) {
	resp, err := o.request(ctx, fmt.Sprintf("breach/%s", site), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, o.formatError(resp)
	}

	var breach *Breach
	return breach, json.NewDecoder(resp.Body).Decode(&breach)
}

func (o *HaveIBeenPwned) GetDataClasses(ctx context.Context) (*DataClasses, error) {
	resp, err := o.request(ctx, "dataclasses", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, o.formatError(resp)
	}

	var dataClasses *DataClasses
	return dataClasses, json.NewDecoder(resp.Body).Decode(&dataClasses)
}

func (o *HaveIBeenPwned) GetPastedAccount(ctx context.Context, email string) ([]*Paste, error) {
	if err := o.validateEmail(email); err != nil {
		return nil, err
	}
	resp, err := o.request(ctx, fmt.Sprintf("pasteaccount/%s", email), nil, useAccessKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, o.formatError(resp)
	}

	var pastes []*Paste
	return pastes, json.NewDecoder(resp.Body).Decode(&pastes)
}

func (o *HaveIBeenPwned) request(ctx context.Context, resource string, params url.Values, opt ...string) (*http.Response, error) {
	u, err := baseURL.Parse(resource)
	if err != nil {
		return nil, err
	}

	target := u.String()
	if params != nil {
		target = fmt.Sprintf("%s?%s", target, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", Accept)
	req.Header.Set("user-agent", UserAgent)

	opts := strings.Join(opt, ",")
	if len(opts) > 0 && strings.Contains(opts, useAccessKey) {
		req.Header.Set("hibp-api-key", o.accessKey)
	}

	req.Close = true

	return o.client.Do(req)
}

func (o *HaveIBeenPwned) formatError(r *http.Response) *HIBPErrorResponse {
	var ans *HIBPErrorResponse
	if err := json.NewDecoder(r.Body).Decode(&ans); err != nil {
		code := r.StatusCode
		ans = &HIBPErrorResponse{
			Message:     err.Error(),
			Description: http.StatusText(code),
			Code:        code,
		}
	}
	if retryAfter, _ := strconv.Atoi(r.Header.Get("retry-after")); retryAfter > 0 && ans != nil {
		ans.RetryAfter = retryAfter
	}
	return ans
}

func (o *HaveIBeenPwned) validateEmail(value string) error {
	if len(value) < 6 || len(value) > 150 {
		return errors.New("email length must be from 6 to 150")
	}

	_, err := mail.ParseAddress(value)
	if err != nil {
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

type HIBPErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"-"`
	Code        int    `json:"statusCode"`
	RetryAfter  int    `json:"-"`
}

func (e *HIBPErrorResponse) Error() string {
	return fmt.Sprintf("%d: %s %s %d", e.Code, e.Description, e.Message, e.RetryAfter)
}
