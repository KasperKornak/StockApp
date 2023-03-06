package main

import (
	"fmt"

	"github.com/piquette/finance-go/equity"
)

func main() {
	// A single quote.
	// ---------------
	q, err := equity.Get("AAPL")
	if err != nil {
		// Uh-oh!
		panic(err)
	}
	// All good.
	fmt.Println(q)

	// Multiple quotes.
	// ----------------
	symbols := []string{"AAPL", "GOOG", "MSFT"}
	iter := equity.List(symbols)

	// Iterate over results. Will exit upon any error.
	for iter.Next() {
		q := iter.Equity()
		fmt.Println(q)
	}

	// Catch an error, if there was one.
	if iter.Err() != nil {
		// Uh-oh!
		panic(err)
	}
}
