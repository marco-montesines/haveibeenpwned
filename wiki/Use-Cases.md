# Use Cases

Real-world scenarios this project is built for. All of them are **defensive**: checking your own users, accounts, and domains against known breach data.

## 1. Signup / password-change policy with breach checking

The canonical use case. A registration or password-change form enforces a policy like:

> Your password must be:
>
> - At least 10 characters
> - Contain at least 8 unique characters
> - Different from your name or email address
>
> We encourage you not to use the same password you have used for another site, and will check your new password against datasets like HaveIBeenPwned.

This mirrors [NIST SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html), which recommends verifiers check new passwords against lists of compromised credentials. The k-anonymity range API makes the check safe: **the plaintext password never leaves your process** — only the first five characters of its SHA-1 hash are sent, and responses are padded so the queried hash cannot be inferred.

### In Go

```go
func validateNewPassword(ctx context.Context, client *hibp.HaveIBeenPwned, password, name, email string) error {
	if len(password) < 10 {
		return errors.New("password must be at least 10 characters")
	}
	unique := make(map[rune]struct{})
	for _, r := range password {
		unique[r] = struct{}{}
	}
	if len(unique) < 8 {
		return errors.New("password must contain at least 8 unique characters")
	}
	lower := strings.ToLower(password)
	if lower == strings.ToLower(name) || lower == strings.ToLower(email) {
		return errors.New("password must differ from your name and email address")
	}
	count, err := client.PwnedPasswordCount(ctx, password)
	if err != nil {
		// Fail open or closed depending on your risk appetite — but never
		// block signups just because HIBP is briefly unreachable.
		log.Printf("hibp check unavailable: %v", err)
		return nil
	}
	if count > 0 {
		return fmt.Errorf("this password appears in %d known data breaches — please choose another", count)
	}
	return nil
}
```

### In PHP (via the FrankenPHP extension)

```php
if (hibp_pwned_password_count($_POST['password']) > 0) {
    throw new RuntimeException('This password appears in known data breaches.');
}
```

### In any other language (via the HTTP API)

```bash
curl -s -X POST http://hibp-api:8080/v1/pwnedpassword -d '{"password":"P@ssw0rd"}'
# {"count":6421042,"pwned":true}
```

See [[HTTP API Reference|HTTP-API-Reference]] — the check is a `POST` so the password never appears in URLs or access logs.

## 2. Breach exposure check for your own accounts

Check whether your accounts appear in known breaches or public pastes (requires an [API key](https://haveibeenpwned.com/API/Key)):

```bash
export HIBP_API_KEY=...
hibp account info@example.com          # breaches containing this account
hibp pastes info@example.com           # pastes containing this account
```

A `404` from HIBP means "not found in any breach" — the client returns an empty result, not an error, so scripting is straightforward.

## 3. Ops / IT: monitor breaches affecting a domain

Security or IT teams can watch whether services their organization relies on have been breached — no API key needed for the breach catalogue:

```bash
hibp breaches -domain adobe.com | jq '.[].Title'
```

Or on a schedule from any service via the HTTP API (`GET /v1/breaches?domain=...`), alerting when a new breach appears for a vendor domain. The `Breach` model includes `BreachDate`, `PwnCount`, `DataClasses` (what kind of data leaked), and flags like `IsVerified` and `IsMalware` for filtering.

## 4. Incident response: what data was exposed?

When a breach hits a service your users share credentials with, pull the details to scope your response:

```bash
hibp breach Adobe | jq '{date: .BreachDate, count: .PwnCount, data: .DataClasses}'
```

`DataClasses` tells you whether passwords, emails, or more sensitive data classes were involved — which drives whether you force resets or just notify.

## 5. Non-Go platforms: one shared breach-check service

Run `hibp serve` (or the published container image) once, and every service in your stack — PHP, Python, Node, shell — gets breach checking over plain HTTP JSON, without each team reimplementing a HIBP client or handling SHA-1/k-anonymity themselves. See [[Deployment and IaC|Deployment-and-IaC]].

## A note on responsible use

Account-level lookups are for accounts you own or are authorized to check. Respect the [HIBP acceptable use policy](https://haveibeenpwned.com/API/v3#AcceptableUse), identify your application with `WithUserAgent`, and if you display breach or paste data to users, visibly attribute [Have I Been Pwned](https://haveibeenpwned.com) (the data is CC BY 4.0; password range checks carry no attribution requirement).
