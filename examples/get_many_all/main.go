package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dvcrn/go-1password-cli/op"
)

// Example that fetches multiple items across all vaults in a single op call.
// Usage:
//
//	go run examples/get_many_all/main.go "Slack (angr)" "angr.slack.com" "Netspotapp"
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: go run examples/get_many_all/main.go <item1> <item2> [...]")
		return
	}

	client := op.NewOpClient()
	items, err := client.Items(args)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Vault scope: ALL\n")
	fmt.Printf("CMD: op item get - --format json\n")
	fmt.Println("STDIN JSON:")
	fmt.Print("[")
	for i, s := range args {
		if i > 0 {
			fmt.Print(", ")
		}
		b, _ := json.Marshal(map[string]string{"id": s})
		fmt.Print(string(b))
	}
	fmt.Println("]")

	fmt.Printf("Requested: %v\n", args)
	fmt.Printf("Fetched %d items\n", len(items))
	for _, it := range items {
		fmt.Printf("- %s [%s] id=%s\n", it.Title, it.Category, it.ID)
		if len(it.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", it.Tags)
		}
	}
}
