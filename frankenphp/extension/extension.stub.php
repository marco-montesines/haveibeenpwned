<?php

/**
 * IDE / documentation stubs for the hibp FrankenPHP extension.
 * The real implementations live in extension.go (compiled into FrankenPHP).
 *
 * @generate-class-entries
 */

/**
 * Number of times the password appears in the Pwned Passwords corpus.
 * Uses the k-anonymity range API: only the first five characters of the
 * password's SHA-1 hash are transmitted. Returns -1 on lookup failure.
 */
function hibp_pwned_password_count(string $password): int {}

/**
 * JSON-encoded array of breaches, optionally filtered by domain
 * (empty string = all breaches). Returns {"error": "..."} on failure.
 */
function hibp_breaches(string $domain = ""): string {}
