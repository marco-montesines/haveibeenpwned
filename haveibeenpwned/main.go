package main

import (
	"context"
	"fmt"

	"github.com/klopjq/haveibeenpwned"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hibp := haveibeenpwned.New("123")

	// https://haveibeenpwned.com/API/v3#BreachesForAccount
	breaches, err := hibp.GetBreachedAccount(ctx, "info@example.com", "",
		false, false)
	if err != nil {
		fmt.Println(err)
	}
	for i := range breaches {
		fmt.Printf("%#v\n", breaches[i])
	}

	// https://haveibeenpwned.com/API/v3#AllBreaches
	domainBreaches, err := hibp.GetBreaches(ctx, "adobe.com")
	if err != nil {
		fmt.Println(err)
	}
	for i := range domainBreaches {
		fmt.Printf("%#v\n", domainBreaches[i])
	}

	// https://haveibeenpwned.com/API/v3#SingleBreach
	siteBreached, err := hibp.GetBreachedSite(ctx, "Adobe")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", siteBreached)

	// https://haveibeenpwned.com/API/v3#AllDataClasses
	dataClasses, err := hibp.GetDataClasses(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", dataClasses)

	// https://haveibeenpwned.com/API/v3#PastesForAccount
	pastes, err := hibp.GetPastedAccount(ctx, "info@example.com")
	if err != nil {
		fmt.Println(err)
	}
	for i := range pastes {
		fmt.Printf("%#v\n", pastes[i])
	}
}
