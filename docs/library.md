# Library Guide

Integration reference for the Go library. Authoritative API docs are generated from source at [pkg.go.dev](https://pkg.go.dev/github.com/marco-montesines/haveibeenpwned); this page adds integration guidance around them.

## Install

```bash
go get github.com/marco-montesines/haveibeenpwned
```

```go
import hibp "github.com/marco-montesines/haveibeenpwned"
```

## Constructing a client

```go
client := hibp.New(apiKey) // apiKey may be "" for unauthenticated endpoints
```

With options:

```go
client := hibp.New(apiKey,
	hibp.WithHTTPClient(customHTTPClient),        // your *http.Client (timeouts, proxies, instrumentation)
	hibp.WithUserAgent("my-app/1.0"),             // identify your project (please do)
	hibp.WithBaseURL(proxyURL),                   // main API base URL — tests, corporate proxy
	hibp.WithPwnedPasswordsBaseURL(mirrorURL),    // Pwned Passwords base URL — tests, mirror
)
```

The client is safe to construct once and reuse; every method takes a `context.Context` for cancellation and deadlines.

## Methods

| Method | HIBP endpoint | API key | Notes |
| ------ | ------------- | :-----: | ----- |
| `PwnedPasswordCount(ctx, password) (int, error)` | `range/{first5}` | no | k-anonymity; returns how often the password appears in breaches. `0` = not found. |
| `GetBreaches(ctx, domain) ([]*Breach, error)` | `breaches` | no | `domain=""` returns the full catalogue. |
| `GetBreachedSite(ctx, name) (*Breach, error)` | `breach/{name}` | no | Name is the breach identifier, e.g. `"Adobe"`. |
| `GetDataClasses(ctx) (*DataClasses, error)` | `dataclasses` | no | All data classes HIBP tracks. |
| `GetBreachedAccount(ctx, email, domain, truncate, unverified) ([]Breach, error)` | `breachedaccount/{account}` | **yes** | `truncate=true` returns names only; `unverified=true` includes unverified breaches. |
| `GetPastedAccount(ctx, email) ([]*Paste, error)` | `pasteaccount/{account}` | **yes** | Pastes mentioning the account. |

## The "not breached" convention (important)

HIBP answers account lookups for unknown accounts with HTTP `404`. This library translates that to **`(nil, nil)`** — a `nil` result with a `nil` error:

```go
breaches, err := client.GetBreachedAccount(ctx, email, "", true, true)
if err != nil {
	return err                       // a real failure (auth, rate limit, network)
}
if breaches == nil {
	fmt.Println("not found in any breach")   // the good outcome
}
```

Don't treat `nil` as an error, and don't expect an empty slice — check for `nil` explicitly. The same applies to `GetPastedAccount`.

## Error handling & rate limits

All non-2xx API responses (other than the 404 case above) come back as `*hibp.HIBPErrorResponse`, preserving the HTTP status code and, on `429`, the `Retry-After` value:

```go
var apiErr *hibp.HIBPErrorResponse
if errors.As(err, &apiErr) {
	switch apiErr.Code {
	case http.StatusUnauthorized: // 401 — missing/invalid API key
	case http.StatusTooManyRequests: // 429 — back off
		time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
	}
}
```

The client **never retries automatically** — the HIBP acceptable use policy expects consumers to back off deliberately, so the decision is yours.

## Data models

- **`Breach`** — `Name`, `Title`, `Domain`, `BreachDate`, `AddedDate`, `ModifiedDate`, `PwnCount`, `Description`, `LogoPath`, `DataClasses`, and the flags `IsVerified`, `IsFabricated`, `IsSensitive`, `IsRetired`, `IsSpamList`, `IsMalware`, `IsSubscriptionFree`, `IsStealerLog`. JSON field names match the [HIBP breach model](https://haveibeenpwned.com/API/v3#BreachModel) (PascalCase).
- **`Paste`** — `Source`, `Id`, `Title`, `Date`, `EmailCount`.
- **`DataClasses`** — `[]string` of compromised data attributes.

## Testing your integration

Inject mock servers via the base-URL options — no network, no global state:

```go
srv := httptest.NewServer(yourMockHandler)
defer srv.Close()
u, _ := url.Parse(srv.URL)
client := hibp.New("test-key", hibp.WithBaseURL(u), hibp.WithPwnedPasswordsBaseURL(u))
```

This is exactly how the library's own test suite works ([`haveibeenpwned_test.go`](https://github.com/marco-montesines/haveibeenpwned/blob/master/haveibeenpwned_test.go)) — useful as a reference for mock response shapes. Handy fixture: `SHA1("password") = 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8` (range prefix `5BAA6`).

## Privacy properties of `PwnedPasswordCount`

- The password is SHA-1-hashed locally; only the **first five hex characters** of the hash are sent to the API.
- Responses are requested with padding, so observers cannot infer the queried hash from response sizes.
- The full password or full hash never leaves your process.
