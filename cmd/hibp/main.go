// Command hibp is a command line interface and HTTP API server for Troy
// Hunt's Have I Been Pwned API v3, built on the
// github.com/marco-montesines/haveibeenpwned library.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	hibp "github.com/marco-montesines/haveibeenpwned"
)

const usage = `hibp - query Troy Hunt's Have I Been Pwned API v3

Usage:
  hibp [global flags] <command> [command flags] [arguments]

Commands:
  breaches               List all breaches, optionally filtered by domain
  breach <name>          Show a single breach by name (e.g. "Adobe")
  account <email>        List breaches for an account (requires API key)
  pastes <email>         List pastes for an account (requires API key)
  dataclasses            List all data classes
  password [<password>]  Count occurrences in Pwned Passwords (reads stdin if omitted)
  serve                  Run an HTTP JSON API exposing the endpoints above

Global flags:
  -api-key string        HIBP API key (default: $HIBP_API_KEY)
  -base-url string       Override the HIBP API base URL
  -timeout duration      Request timeout (default 30s)

Examples:
  hibp breaches -domain adobe.com
  hibp breach Adobe
  HIBP_API_KEY=... hibp account info@example.com -truncate=false
  echo -n 'P@ssw0rd' | hibp password
  hibp serve -addr :8080
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "hibp:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	global := flag.NewFlagSet("hibp", flag.ExitOnError)
	global.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	apiKey := global.String("api-key", os.Getenv("HIBP_API_KEY"), "HIBP API key")
	baseURL := global.String("base-url", "", "override the HIBP API base URL")
	timeout := global.Duration("timeout", 30*time.Second, "request timeout")
	global.Parse(args)

	if global.NArg() == 0 {
		global.Usage()
		os.Exit(2)
	}

	var opts []hibp.Option
	if *baseURL != "" {
		u, err := url.Parse(*baseURL)
		if err != nil {
			return fmt.Errorf("invalid -base-url: %w", err)
		}
		opts = append(opts, hibp.WithBaseURL(u))
	}
	client := hibp.New(*apiKey, opts...)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	command, rest := global.Arg(0), global.Args()[1:]
	switch command {
	case "breaches":
		fs := flag.NewFlagSet("breaches", flag.ExitOnError)
		domain := fs.String("domain", "", "filter breaches by domain")
		fs.Parse(rest)
		breaches, err := client.GetBreaches(ctx, *domain)
		if err != nil {
			return err
		}
		return printJSON(breaches)

	case "breach":
		if len(rest) != 1 {
			return fmt.Errorf("usage: hibp breach <name>")
		}
		breach, err := client.GetBreachedSite(ctx, rest[0])
		if err != nil {
			return err
		}
		return printJSON(breach)

	case "account":
		fs := flag.NewFlagSet("account", flag.ExitOnError)
		domain := fs.String("domain", "", "filter results by breach domain")
		truncate := fs.Bool("truncate", true, "return breach names only")
		unverified := fs.Bool("unverified", true, "include unverified breaches")
		email, err := popArg(&rest, "usage: hibp account <email> [flags]")
		if err != nil {
			return err
		}
		fs.Parse(rest)
		breaches, err := client.GetBreachedAccount(ctx, email, *domain, *truncate, *unverified)
		if err != nil {
			return err
		}
		if breaches == nil {
			fmt.Fprintln(os.Stderr, "good news: account not found in any breach")
			return printJSON([]hibp.Breach{})
		}
		return printJSON(breaches)

	case "pastes":
		email, err := popArg(&rest, "usage: hibp pastes <email>")
		if err != nil {
			return err
		}
		pastes, err := client.GetPastedAccount(ctx, email)
		if err != nil {
			return err
		}
		if pastes == nil {
			fmt.Fprintln(os.Stderr, "good news: account not found in any paste")
			return printJSON([]hibp.Paste{})
		}
		return printJSON(pastes)

	case "dataclasses":
		dataClasses, err := client.GetDataClasses(ctx)
		if err != nil {
			return err
		}
		return printJSON(dataClasses)

	case "password":
		password, err := readPassword(rest)
		if err != nil {
			return err
		}
		count, err := client.PwnedPasswordCount(ctx, password)
		if err != nil {
			return err
		}
		return printJSON(map[string]any{"pwned": count > 0, "count": count})

	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		addr := fs.String("addr", ":8080", "listen address")
		fs.Parse(rest)
		return serve(*addr, client)

	default:
		global.Usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

// popArg removes and returns the first positional argument so that command
// flags may follow it (e.g. "hibp account info@example.com -truncate=false").
func popArg(rest *[]string, usageMsg string) (string, error) {
	if len(*rest) == 0 || strings.HasPrefix((*rest)[0], "-") {
		return "", fmt.Errorf("%s", usageMsg)
	}
	arg := (*rest)[0]
	*rest = (*rest)[1:]
	return arg, nil
}

// readPassword takes the password from the argument list or, preferably,
// from stdin so it does not end up in the shell history.
func readPassword(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	stat, err := os.Stdin.Stat()
	if err == nil && stat.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprint(os.Stderr, "password: ")
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("reading password from stdin: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
