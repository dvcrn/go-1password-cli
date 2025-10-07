package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dvcrn/go-1password-cli/op"
)

func main() {
	client := op.NewOpClient()

	args := os.Args[1:]
	var refs []string

	if len(args) > 0 {
		refs = args
	} else {
		refs = []string{
			"op://chainenv/RCLONE_ENCRYPTION_PASSWORD_GDRIVE/password",
			"op://chainenv/HOMESERVER_WG_ORBIT_MACBOOK_PUBLIC_KEY/password",
		}
	}

	fmt.Printf("Reading %d secret references in a single batch...\n", len(refs))
	results, err := client.ReadMulti(refs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully fetched %d secrets:\n", len(results))
	for ref, value := range results {
		fmt.Printf("- %s: %s\n", ref, value)
	}
}
