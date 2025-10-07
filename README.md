# go-1password-cli

Super thin wrapper around 1passwords `op` CLI tool.

**Note:** This is not for the secret automation API, but for the CLI tool. There is no authentication, you need to have the `op` CLI tool installed and logged in.

1Password will prompt you if you want to allow access on usage.


## Installation
```
go get github.com/dvcrn/go-1password-cli
```

## Usage

```go
package main

import (
    "github.com/dvcrn/go-1password-cli/op"
)

func main() {
	client := op.NewOpClient()

	// get all vaults
	vaults, err := client.Vaults()

	// get a specific vault
	vault, err := client.Vault("Private")
	vault, err := client.Vault("<vaultID>")

	// get item
	item, err := client.Item("Netflix")
	item, err := client.Item("<itemID>")
}
```

## API Reference

- [type Vault](<#type-vault>)
- [type Item](<#type-item>)
- [type OpClient](<#type-opclient>)
  - [func NewOpClient() *OpClient](<#func-newopclient>)
  - [func (c *OpClient) Item(itemIDOrName string) (*Item, error)](<#func-opclient-item>)
  - [func (c *OpClient) LookupField(vaultIdOrName string, itemIdOrName string, fieldName string) (string, error)](<#func-opclient-lookupfield>)
  - [func (c *OpClient) Vault(vaultIDOrName string) (*Vault, error)](<#func-opclient-vault>)
  - [func (c *OpClient) VaultItem(itemIDOrName string, vaultIDOrName string) (*Item, error)](<#func-opclient-vaultitem>)
  - [func (c *OpClient) Vaults() ([]*Vault, error)](<#func-opclient-vaults>)
  - [func (c *Client) VaultItems(itemIDsOrNames []string, vaultIDOrName string) ([]*Item, error)](<#func-client-vaultitems>)
  - [func (c *Client) Items(itemIDsOrNames []string) ([]*Item, error)](<#func-client-items>)


## type Item

```go
type Item struct {
    AdditionalInformation string    `json:"additional_information,omitempty"`
    Category              string    `json:"category"`
    CreatedAt             time.Time `json:"created_at"`
    Favorite              bool      `json:"favorite,omitempty"`
    ID                    string    `json:"id"`
    LastEditedBy          string    `json:"last_edited_by"`
    Tags                  []string  `json:"tags,omitempty"`
    Title                 string    `json:"title"`
    UpdatedAt             time.Time `json:"updated_at"`
    Urls                  []struct {
        Href    string `json:"href"`
        Label   string `json:"label,omitempty"`
        Primary bool   `json:"primary,omitempty"`
    }   `json:"urls,omitempty"`
    Vault struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }   `json:"vault"`
    Version int `json:"version"`
}
```

## type Vault

```go
type Vault struct {
    ContentVersion int    `json:"content_version"`
    ID             string `json:"id"`
    Name           string `json:"name"`
}
```

## type OpClient

```go
type OpClient struct{}
```

### func NewOpClient

```go
func NewOpClient() *OpClient
```

### func \(\*OpClient\) Item

```go
func (c *OpClient) Item(itemIDOrName string) (*Item, error)
```

Item returns an item by its ID or name, across all Vaults the user has access to To get items scoped to a specific Vault, use VaultItem\(\)

```go
func (c *Client) CreateItem(vaultIDOrName string, category string, title string, opts ...ItemCreateOption) (*Item, error) {
```

Creates a new item in the specified Vault

### func \(\*OpClient\) ReadItemField

```go
func (c *OpClient) ReadItemField(vaultIdOrName string, itemIdOrName string, fieldName string) (string, error)
```

ReadItemField does a lookup of a specific field within an item, within a vault This is equivalent to op read op://\<vault\>/\<item\>/\<field\>

### func \(\*OpClient\) EditItemField

```go
func (c *OpClient) EditItemField(vaultIdOrName string, itemIdOrName string, assignments ...Assignment) (*Item, error)
```

EditItemField does a lookup of a specific field within an item, within a vault and edit the fields in the item. This is equivalent to op item edit <item-id> field=value ...


### func \(\*OpClient\) Vault

```go
func (c *OpClient) Vault(vaultIDOrName string) (*Vault, error)
```

```go
func (c *Client) CreateVault(name string, opts ...VaultCreateOption) (*Vault, error) {
```

Creates a new Vault

Vault retrieves a vault by its ID or name If you have a Vault named "Private", you can specify this as either "Private" or with its ID

### func \(\*OpClient\) VaultItem

```go
func (c *OpClient) VaultItem(itemIDOrName string, vaultIDOrName string) (*Item, error)
```

VaultItem returns an item by it's ID or name, within the specified Vault To get items across all Vaults, use Item\(\)

### func \(\*OpClient\) Vaults

```go
func (c *OpClient) Vaults() ([]*Vault, error)
```

Vaults returns a list of all vaults in the current 1Password account

### func (\*Client) VaultItems

```go
func (c *Client) VaultItems(itemIDsOrNames []string, vaultIDOrName string) ([]*Item, error)
```

Fetch multiple items in a single `op` CLI call scoped to a vault. Uses `op item get -` with a JSON array on stdin for speed.

Example:

```go
client := op.NewOpClient()
items, err := client.VaultItems([]string{"Slack (angr)", "angr.slack.com", "Netspotapp"}, "Personal")
```

### func (\*Client) Items

```go
func (c *Client) Items(itemIDsOrNames []string) ([]*Item, error)
```

Fetch multiple items across all accessible vaults in one call. Uses the same batching strategy as `VaultItems` but without a vault filter.

Example:

```go
client := op.NewOpClient()
items, err := client.Items([]string{"Slack (angr)", "angr.slack.com"})
```
