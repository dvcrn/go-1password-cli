package main

import (
	"fmt"
	"log"

	"github.com/dvcrn/go-1password-cli/op"
)

func main() {
	client := op.NewOpClient()

	// List all vaults
	vaults, err := client.Vaults()
	if err != nil {
		log.Fatal(err)
	}

	if len(vaults) == 0 {
		fmt.Println("No vaults found")
		return
	}

	// List items in the first vault
	vault := vaults[0]
	fmt.Printf("Listing items in vault: %s\n", vault.Name)

	// List all items in the vault
	items, err := client.ItemsByVault(vault.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d items\n", len(items))
	for _, item := range items {
		fmt.Printf("- %s (%s)\n", item.Title, item.Category)
		if len(item.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", item.Tags)
		}
	}

	// List items with specific tags
	fmt.Println("\nListing items with tag 'personal'")
	taggedItems, err := client.ItemsByVault(vault.ID, op.WithTags([]string{"personal"}))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d items with tag 'personal'\n", len(taggedItems))
	for _, item := range taggedItems {
		fmt.Printf("- %s (%s)\n", item.Title, item.Category)
	}
}