package op

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Client struct{}

type Vault struct {
	ContentVersion int    `json:"content_version"`
	ID             string `json:"id"`
	Name           string `json:"name"`
}

type Field struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Purpose   string `json:"purpose,omitempty"`
	Label     string `json:"label"`
	Reference string `json:"reference"`
	Value     string `json:"value,omitempty"`
}

type Item struct {
	AdditionalInformation string    `json:"additional_information,omitempty"`
	Category              string    `json:"category"`
	CreatedAt             time.Time `json:"created_at"`
	Favorite              bool      `json:"favorite,omitempty"`
	Fields                []Field   `json:"fields"`
	ID                    string    `json:"id"`
	LastEditedBy          string    `json:"last_edited_by"`
	Tags                  []string  `json:"tags,omitempty"`
	Title                 string    `json:"title"`
	UpdatedAt             time.Time `json:"updated_at"`
	Urls                  []struct {
		Href    string `json:"href"`
		Label   string `json:"label,omitempty"`
		Primary bool   `json:"primary,omitempty"`
	} `json:"urls,omitempty"`
	Vault struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vault"`
	Version int `json:"version"`
}

func NewOpClient() *Client {
	return &Client{}
}

func (c *Client) runOp(opCmd string, args []string) ([]byte, error) {
	cmdArgs := []string{opCmd}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--format", "json")

	cmd := exec.Command("op", cmdArgs...)
	errBuf := bytes.NewBuffer(nil)
	cmd.Stderr = errBuf

	out, err := cmd.Output()
	if err != nil {
		if errBuf.String() != "" {
			return nil, fmt.Errorf("op returned err: %s", errBuf.String())
		}
		return nil, err
	}

	return out, nil
}

func (c *Client) runOpAndUnmarshal(opCmd string, args []string, unmarshalInto any) error {
	out, err := c.runOp(opCmd, args)
	if err != nil {
		return err
	}

	err = json.Unmarshal(out, unmarshalInto)
	if err != nil {
		return err
	}

	return nil
}

// Vaults returns a list of all vaults in the current 1Password account
func (c *Client) Vaults() ([]*Vault, error) {
	var out []*Vault
	err := c.runOpAndUnmarshal("vault", []string{"list"}, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Vault retrieves a vault by its ID or name
// If you have a Vault named "Private", you can specify this as either "Private" or with its ID
func (c *Client) Vault(vaultIDOrName string) (*Vault, error) {
	var out Vault
	err := c.runOpAndUnmarshal("vault", []string{"get", vaultIDOrName}, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// Item returns an item by its ID or name, across all Vaults the user has access to
// To get items scoped to a specific Vault, use VaultItem()
func (c *Client) Item(itemIDOrName string) (*Item, error) {
	var out Item
	err := c.runOpAndUnmarshal("item", []string{"get", itemIDOrName}, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// VaultItem returns an item by it's ID or name, within the specified Vault
// To get items across all Vaults, use Item()
func (c *Client) VaultItem(itemIDOrName string, vaultIDOrName string) (*Item, error) {
	var out Item
	err := c.runOpAndUnmarshal("item", []string{"get", itemIDOrName, "--vault", vaultIDOrName}, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// ReadItemField does a lookup of a specific field within an item, within a vault
// `lookupIdentifier` is a string in the format `op://<vault>/<item>/<field>`
// This is equivalent to `op read op://<vault>/<item>/<field>`
func (c *Client) Read(lookupIdentifier string) (string, error) {
	out, err := c.runOp("read", []string{lookupIdentifier})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// ReadItemField does a lookup of a specific field within an item, within a vault
// This is equivalent to `op read op://<vault>/<item>/<field>`
func (c *Client) ReadItemField(vaultIdOrName string, itemIdOrName string, fieldName string) (string, error) {
	lookupString := fmt.Sprintf("op://%s/%s/%s", vaultIdOrName, itemIdOrName, fieldName)
	return c.Read(lookupString)
}

// EditItemField edits the fields of a specific item.
// This is equivalent to `op item edit <itemID> assignment ...
func (c *Client) EditItemField(vaultIdOrName string, itemIdOrName string, assignments ...Assignment) (*Item, error) {

	if len(assignments) == 0 {
		return nil, errors.New("no assignments specified")
	}

	item, err := c.VaultItem(itemIdOrName, vaultIdOrName)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, len(assignments)+2)
	args = append(args, "edit", item.ID)
	for _, assignment := range assignments {
		args = append(args, fmt.Sprintf("%s=%s", assignment.Name, assignment.Value))
	}

	var out Item
	err = c.runOpAndUnmarshal("item", args, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Assignment of a field value to the item. used in EditItemField.
type Assignment struct {
	Name  string
	Value string
}
