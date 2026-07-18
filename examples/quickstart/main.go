// A minimal example of the haveibeenpwned library. The password check and
// breach catalogue work without an API key; account lookups need one from
// https://haveibeenpwned.com/API/Key (set HIBP_API_KEY).
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	hibp "github.com/marco-montesines/haveibeenpwned"
)

func main() {
	ctx := context.Background()
	client := hibp.New(os.Getenv("HIBP_API_KEY"))

	// Check a password against Pwned Passwords (k-anonymity, no API key).
	count, err := client.PwnedPasswordCount(ctx, "P@ssw0rd")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("password seen %d times in breaches\n", count)

	// List breaches for a domain (no API key).
	breaches, err := client.GetBreaches(ctx, "adobe.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range breaches {
		fmt.Printf("%s (%s): %d accounts\n", b.Title, b.BreachDate, b.PwnCount)
	}

	// Look up an account (requires HIBP_API_KEY).
	if os.Getenv("HIBP_API_KEY") != "" {
		accountBreaches, err := client.GetBreachedAccount(ctx, "info@example.com", "", true, true)
		if err != nil {
			log.Fatal(err)
		}
		if accountBreaches == nil {
			fmt.Println("account not found in any breach")
		}
		for _, b := range accountBreaches {
			fmt.Println("breached:", b.Name)
		}
	}
}
