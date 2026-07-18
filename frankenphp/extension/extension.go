// Package extension exposes Have I Been Pwned lookups to PHP as a native
// FrankenPHP extension written in Go.
//
// It provides the PHP functions:
//
//	hibp_pwned_password_count(string $password): int
//	hibp_breaches(string $domain = ""): string
//
// The extension reads the HIBP API key from the HIBP_API_KEY environment
// variable; both exported functions work without a key.
//
// It must be compiled into FrankenPHP with xcaddy — see frankenphp/Dockerfile
// and frankenphp/README.md in this repository.
package extension

// #cgo CFLAGS: -D_GNU_SOURCE
// #include "extension.h"
import "C"

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/dunglas/frankenphp"
	hibp "github.com/marco-montesines/haveibeenpwned"
)

func init() {
	frankenphp.RegisterExtension(unsafe.Pointer(&C.hibp_module_entry))
}

var (
	clientOnce sync.Once
	client     *hibp.HaveIBeenPwned
)

func hibpClient() *hibp.HaveIBeenPwned {
	clientOnce.Do(func() {
		client = hibp.New(os.Getenv("HIBP_API_KEY"))
	})
	return client
}

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// hibp_pwned_password_count reports how many times the password appears in
// the Pwned Passwords corpus (k-anonymity: only the first five characters of
// the SHA-1 hash are sent). Returns -1 on lookup failure.
//
//export hibp_pwned_password_count
func hibp_pwned_password_count(password *C.zend_string) int64 {
	ctx, cancel := requestContext()
	defer cancel()

	count, err := hibpClient().PwnedPasswordCount(ctx, frankenphp.GoString(unsafe.Pointer(password)))
	if err != nil {
		return -1
	}
	return int64(count)
}

// hibp_breaches returns the breach catalogue as a JSON string, optionally
// filtered by domain (empty string = all breaches). On failure it returns a
// JSON object of the form {"error": "..."}.
//
//export hibp_breaches
func hibp_breaches(domain *C.zend_string) *C.zend_string {
	ctx, cancel := requestContext()
	defer cancel()

	breaches, err := hibpClient().GetBreaches(ctx, frankenphp.GoString(unsafe.Pointer(domain)))

	var payload []byte
	if err != nil {
		payload, _ = json.Marshal(map[string]string{"error": err.Error()})
	} else {
		payload, err = json.Marshal(breaches)
		if err != nil {
			payload, _ = json.Marshal(map[string]string{"error": err.Error()})
		}
	}
	return (*C.zend_string)(frankenphp.PHPString(string(payload), false))
}
