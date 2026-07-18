package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	hibp "github.com/marco-montesines/haveibeenpwned"
)

// serve runs a small HTTP JSON API in front of the library so non-Go
// consumers (PHP, shell scripts, other services) can use it over HTTP.
func serve(addr string, client *hibp.HaveIBeenPwned) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           logRequests(apiHandler(client)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("hibp API listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Printf("received %s, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

func apiHandler(client *hibp.HaveIBeenPwned) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /v1/breaches", func(w http.ResponseWriter, r *http.Request) {
		breaches, err := client.GetBreaches(r.Context(), r.URL.Query().Get("domain"))
		respond(w, breaches, err)
	})

	mux.HandleFunc("GET /v1/breaches/{name}", func(w http.ResponseWriter, r *http.Request) {
		breach, err := client.GetBreachedSite(r.Context(), r.PathValue("name"))
		respond(w, breach, err)
	})

	mux.HandleFunc("GET /v1/dataclasses", func(w http.ResponseWriter, r *http.Request) {
		dataClasses, err := client.GetDataClasses(r.Context())
		respond(w, dataClasses, err)
	})

	mux.HandleFunc("GET /v1/breachedaccount/{email}", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		truncate := boolParam(q.Get("truncate"), true)
		unverified := boolParam(q.Get("unverified"), true)
		breaches, err := client.GetBreachedAccount(r.Context(), r.PathValue("email"), q.Get("domain"), truncate, unverified)
		if breaches == nil && err == nil {
			breaches = []hibp.Breach{}
		}
		respond(w, breaches, err)
	})

	mux.HandleFunc("GET /v1/pasteaccount/{email}", func(w http.ResponseWriter, r *http.Request) {
		pastes, err := client.GetPastedAccount(r.Context(), r.PathValue("email"))
		if pastes == nil && err == nil {
			pastes = []*hibp.Paste{}
		}
		respond(w, pastes, err)
	})

	// POST keeps the password out of URLs, query strings, and access logs.
	mux.HandleFunc("POST /v1/pwnedpassword", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": `expected JSON body {"password": "..."}`})
			return
		}
		count, err := client.PwnedPasswordCount(r.Context(), body.Password)
		respond(w, map[string]any{"pwned": count > 0, "count": count}, err)
	})

	return mux
}

func respond(w http.ResponseWriter, v any, err error) {
	if err != nil {
		status := http.StatusBadGateway
		var apiErr *hibp.HIBPErrorResponse
		if errors.As(err, &apiErr) && apiErr.Code >= 400 {
			status = apiErr.Code
			if apiErr.RetryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(apiErr.RetryAfter))
			}
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func boolParam(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond)))
	})
}
