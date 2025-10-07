package op

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
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
    if os.Getenv("OP_GO_DEBUG") != "" {
        fmt.Fprintf(os.Stderr, "[op-go] cmd: op %s\n", strings.Join(cmdArgs, " "))
    }
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

// runOpWithInput runs an op CLI command and writes the provided stdin.
func (c *Client) runOpWithInput(opCmd string, args []string, stdin []byte) ([]byte, error) {
    cmdArgs := []string{opCmd}
    cmdArgs = append(cmdArgs, args...)
    cmdArgs = append(cmdArgs, "--format", "json")

    cmd := exec.Command("op", cmdArgs...)
    cmd.Stdin = bytes.NewReader(stdin)
    if os.Getenv("OP_GO_DEBUG") != "" {
        preview := string(stdin)
        if len(preview) > 512 {
            preview = preview[:512] + "..."
        }
        fmt.Fprintf(os.Stderr, "[op-go] cmd: op %s\n[op-go] stdin:\n%s\n", strings.Join(cmdArgs, " "), preview)
    }

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

// VaultItems returns items by IDs or names within the specified vault using
// a single `op item get -` invocation. Input is provided on stdin as a JSON
// array of objects with an `id` key, e.g.:
//   [{"id":"Slack (angr)"}, {"id":"angr.slack.com"}]
// If the batch call fails, the error is returned without fallback.
func (c *Client) VaultItems(itemIDsOrNames []string, vaultIDOrName string) ([]*Item, error) {
    if len(itemIDsOrNames) == 0 {
        return nil, errors.New("no items specified")
    }

    // Build JSON array payload: [{"id": specifier}, ...]
    type spec struct{ ID string `json:"id"` }
    payload := make([]spec, 0, len(itemIDsOrNames))
    for _, s := range itemIDsOrNames {
        payload = append(payload, spec{ID: s})
    }
    stdinBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    // Prepare args with optional vault scoping
    args := []string{"get", "-"}
    if vaultIDOrName != "" {
        args = append(args, "--vault", vaultIDOrName)
    }

    out, err := c.runOpWithInput("item", args, stdinBytes)
    if err != nil {
        return nil, err
    }

    var arr []*Item
    if err := json.Unmarshal(out, &arr); err == nil {
        return arr, nil
    }

    dec := json.NewDecoder(bytes.NewReader(out))
    var items []*Item
    for {
        var it Item
        if err := dec.Decode(&it); err != nil {
            if errors.Is(err, io.EOF) {
                break
            }
            return nil, err
        }
        itCopy := it
        items = append(items, &itCopy)
    }
    return items, nil
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

// CreateVault creates a new vault with the given name and optional parameters
func (c *Client) CreateVault(name string, opts ...VaultCreateOption) (*Vault, error) {
	args := []string{"create", name}

	// Apply options
	for _, opt := range opts {
		opt.apply(&args)
	}

	var out Vault
	err := c.runOpAndUnmarshal("vault", args, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// CreateItem creates a new item in the specified vault
func (c *Client) CreateItem(vaultIDOrName string, category string, title string, opts ...ItemCreateOption) (*Item, error) {
	args := []string{"create",
		"--vault", vaultIDOrName,
		"--category", category,
		"--title", title,
	}

	// Apply options
	for _, opt := range opts {
		opt.apply(&args)
	}

	var out Item
	err := c.runOpAndUnmarshal("item", args, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// VaultCreateOption represents options for vault creation
type VaultCreateOption interface {
	apply(*[]string)
}

// ItemCreateOption represents options for item creation
type ItemCreateOption interface {
	apply(*[]string)
}

// Vault creation options
type vaultDescription string
type vaultIcon string
type vaultAllowAdminsToManage bool

func (d vaultDescription) apply(args *[]string) {
	*args = append(*args, "--description", string(d))
}

func (i vaultIcon) apply(args *[]string) {
	*args = append(*args, "--icon", string(i))
}

func (a vaultAllowAdminsToManage) apply(args *[]string) {
	*args = append(*args, "--allow-admins-to-manage", fmt.Sprintf("%t", bool(a)))
}

// WithVaultDescription sets the vault description
func WithVaultDescription(description string) VaultCreateOption {
	return vaultDescription(description)
}

// WithVaultIcon sets the vault icon
func WithVaultIcon(icon string) VaultCreateOption {
	return vaultIcon(icon)
}

// WithVaultAllowAdminsToManage sets whether administrators can manage the vault
func WithVaultAllowAdminsToManage(allow bool) VaultCreateOption {
	return vaultAllowAdminsToManage(allow)
}

// Item creation options
type itemURL string
type itemGeneratePassword string
type itemFavorite bool
type itemTags []string
type itemAssignments []Assignment

func (u itemURL) apply(args *[]string) {
	*args = append(*args, "--url", string(u))
}

func (g itemGeneratePassword) apply(args *[]string) {
	*args = append(*args, "--generate-password="+string(g))
}

func (f itemFavorite) apply(args *[]string) {
	if bool(f) {
		*args = append(*args, "--favorite")
	}
}

func (t itemTags) apply(args *[]string) {
	*args = append(*args, "--tags", strings.Join([]string(t), ","))
}

func (a itemAssignments) apply(args *[]string) {
	for _, assignment := range a {
		*args = append(*args, fmt.Sprintf("%s=%s", assignment.Name, assignment.Value))
	}
}

// WithItemURL sets the item URL
func WithItemURL(url string) ItemCreateOption {
	return itemURL(url)
}

// WithItemGeneratePassword adds a generated password with optional recipe
func WithItemGeneratePassword(recipe string) ItemCreateOption {
	return itemGeneratePassword(recipe)
}

// WithItemFavorite marks the item as favorite
func WithItemFavorite(favorite bool) ItemCreateOption {
	return itemFavorite(favorite)
}

// WithItemTags sets the item tags
func WithItemTags(tags []string) ItemCreateOption {
	return itemTags(tags)
}

// WithItemAssignments sets field assignments for the item
func WithItemAssignments(assignments []Assignment) ItemCreateOption {
	return itemAssignments(assignments)
}

// ItemListOption represents options for listing items
type ItemListOption interface {
	apply(*[]string)
}

// Item list options
type itemListTags []string

func (t itemListTags) apply(args *[]string) {
	*args = append(*args, "--tags", strings.Join([]string(t), ","))
}

// WithTags filters items by tags when listing
func WithTags(tags []string) ItemListOption {
	return itemListTags(tags)
}

// ItemsByVault returns a list of items in the specified vault
func (c *Client) ItemsByVault(vaultIDOrName string, opts ...ItemListOption) ([]*Item, error) {
	args := []string{"list", "--vault", vaultIDOrName}

	// Apply options
	for _, opt := range opts {
		opt.apply(&args)
	}

	var out []*Item
	err := c.runOpAndUnmarshal("item", args, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
