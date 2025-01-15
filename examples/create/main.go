package main

import (
	"fmt"
	"log"

	"github.com/dvcrn/go-1password-cli/op"
)

func main() {
	client := op.NewOpClient()

	// Check if vault already exists
	vaults, err := client.Vaults()
	if err != nil {
		log.Fatal(err)
	}

	var vault *op.Vault
	for _, v := range vaults {
		if v.Name == "My Vault" {
			vault = v
			fmt.Println("Using existing vault")
			fmt.Printf("ID: %s, Name: %s, ContentVersion: %d\n", vault.ID, vault.Name, vault.ContentVersion)
			break
		}
	}

	if vault == nil {
		// Create a new vault
		var err error
		vault, err = client.CreateVault("My Vault",
			op.WithVaultDescription("My vault description"),
			op.WithVaultIcon("treasure-chest"),
			op.WithVaultAllowAdminsToManage(true),
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Created new vault")
		fmt.Printf("ID: %s, Name: %s, ContentVersion: %d\n", vault.ID, vault.Name, vault.ContentVersion)
	}

	// Create a new login item
	item, err := client.CreateItem(vault.ID, "login", "My Login",
		op.WithItemURL("https://example.com"),
		op.WithItemGeneratePassword("20,letters,digits"),
		op.WithItemFavorite(true),
		op.WithItemTags([]string{"personal", "work"}),
		op.WithItemAssignments([]op.Assignment{
			{Name: "username", Value: "user@example.com"},
			{Name: "notes", Value: "Some notes"},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created new item")
	fmt.Printf("ID: %s, Title: %s, Category: %s\n", item.ID, item.Title, item.Category)
}
