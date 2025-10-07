package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dvcrn/go-1password-cli/op"
)

func main() {
	client := op.NewOpClient()

	var vault *op.Vault
	vaultEnv := os.Getenv("VAULT")
	var err error
	if vaultEnv != "" {
		vault, err = client.Vault(vaultEnv)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		vaults, err := client.Vaults()
		if err != nil {
			log.Fatal(err)
		}
		if len(vaults) == 0 {
			fmt.Println("No vaults found")
			return
		}
		vault = vaults[0]
	}

	args := os.Args[1:]
	var specifiers []string
	if len(args) > 0 {
		specifiers = args
	} else {
		list, err := client.ItemsByVault(vault.ID)
		if err != nil {
			log.Fatal(err)
		}
		if len(list) == 0 {
			fmt.Printf("No items found in vault: %s\n", vault.Name)
			return
		}
		n := len(list)
		if n > 3 {
			n = 3
		}
		specifiers = make([]string, 0, n)
		for i := 0; i < n; i++ {
			specifiers = append(specifiers, list[i].ID)
		}
	}

	// Batch call: single op invocation
	items, err := client.VaultItems(specifiers, vault.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Vault: %s (%s)\n", vault.Name, vault.ID)
	fmt.Printf("CMD: op item get - --vault %s --format json\n", vault.ID)
	fmt.Println("STDIN JSON:")
	fmt.Print("[")
	for i, s := range specifiers {
		if i > 0 {
			fmt.Print(", ")
		}
		// naive JSON quoting for display-only; library sends proper JSON
		b, _ := json.Marshal(map[string]string{"id": s})
		fmt.Print(string(b))
	}
	fmt.Println("]")
	fmt.Printf("Requested: %v\n", specifiers)
	fmt.Printf("Fetched %d items\n", len(items))
	for _, it := range items {
		fmt.Printf("- %s [%s] id=%s\n", it.Title, it.Category, it.ID)
		if len(it.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", it.Tags)
		}
	}
}
