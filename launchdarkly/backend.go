package launchdarkly

import (
	"context"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/pkg/errors"
)

// Factory creates a new usable instance of this secrets engine.
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, errors.Wrap(err, "failed to create factory")
	}
	return b, nil
}

// backend is the actual backend.
type backend struct {
	*framework.Backend
	store map[string][]byte

	clientMutex sync.RWMutex
}

// Backend creates a new backend.
func Backend(c *logical.BackendConfig) *backend {
	//var b backend

	b := &backend{
		store: make(map[string][]byte),
	}

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        backendHelp,
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: []*framework.Path{
			// launchdarkly/info
			&framework.Path{
				Pattern:      "info",
				HelpSynopsis: "Display information about this plugin",
				HelpDescription: `

Displays information about the plugin, such as the plugin version and where to
get help.

`,
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation: b.pathInfo,
				},
			},
			// launchdarkly/config
			&framework.Path{
				Pattern:      "config",
				HelpSynopsis: "Configure LaunchDarkly secret engine.",
				HelpDescription: `

Configure launchdarkly secret engine.

`,
				Fields: map[string]*framework.FieldSchema{
					"access_token": &framework.FieldSchema{
						Type:        framework.TypeString,
						Description: "LaunchDarkly access Token",
						Default:     "",
					},
					"base_uri": &framework.FieldSchema{
						Type:        framework.TypeString,
						Description: "LaunchDarkly baseUri.",
						Default:     "",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation:   b.pathConfigRead,
					logical.UpdateOperation: b.pathConfigWrite,
				},
			},
			&framework.Path{
				Pattern: "role",
				Fields: map[string]*framework.FieldSchema{
					"name": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the role.",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ListOperation: b.pathRoleRead,
				},
			},
			&framework.Path{
				Pattern: "relay/policy",
				Fields: map[string]*framework.FieldSchema{
					"inline_policy": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the custom role to create a token for. Must be an existing custom role.",
						Required:    true,
					},
					"name": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name to be used for the token.",
						Required:    true,
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.CreateOperation: b.pathRelayWrite,
					logical.UpdateOperation: b.pathRelayWrite,
				},
			},
			&framework.Path{
				Pattern: "relay/" + GenericLDKeyWithAtRegex("name"),
				Fields: map[string]*framework.FieldSchema{
					"name": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name to be used for the token.",
						Required:    true,
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation:   b.pathRelayRead,
					logical.DeleteOperation: b.pathRelayDelete,
				},
			},
			&framework.Path{
				//Pattern: "role/" + framework.GenericNameWithAtRegex("customrole"),
				Pattern: "role/" + GenericLDKeyWithAtRegex("customrole"),
				Fields: map[string]*framework.FieldSchema{
					"customrole": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the custom role to create a token for. Must be an existing custom role.",
					},
					"name": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name to be used for the token.",
						Default:     "vault-generated",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation:   b.pathRoleRead,
					logical.DeleteOperation: b.pathRoleDelete,
				},
			},
			&framework.Path{
				Pattern: "project/" + GenericLDKeyWithAtRegex("project") + "/" + GenericLDKeyWithAtRegex("env"),
				Fields: map[string]*framework.FieldSchema{
					"project": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the project.",
					},
					"env": {
						Type:        framework.TypeLowerCaseString,
						Description: "The env of the project.",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation: b.pathProjectEnvRead,
				},
			},
			&framework.Path{
				Pattern: "project/" + GenericLDKeyWithAtRegex("project") + "/" + GenericLDKeyWithAtRegex("env") + "/reset/" + framework.GenericNameWithAtRegex("type"),
				Fields: map[string]*framework.FieldSchema{
					"project": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the project.",
						Required:    true,
					},
					"env": {
						Type:        framework.TypeLowerCaseString,
						Description: "The env of the project.",
						Required:    true,
					},
					"type": {
						Type:          framework.TypeLowerCaseString,
						Description:   "The type of key to reset.",
						Required:      true,
						AllowedValues: []interface{}{"mobile", "sdk"},
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation: b.pathProjectReset,
				},
			},
			&framework.Path{
				Pattern: "coderefs/" + framework.GenericNameWithAtRegex("project"),
				Fields: map[string]*framework.FieldSchema{
					"project": {
						Type:        framework.TypeLowerCaseString,
						Description: "The name of the project.",
						Required:    true,
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation:   b.pathCoderefsRead,
					logical.DeleteOperation: b.pathCoderefsDelete,
				},
			},
		},
	}

	return b
}

func (b *backend) Close() {
	b.clientMutex.Lock()
	defer b.clientMutex.Unlock()
}

const backendHelp = `
The LaunchDarkly secrets engine generates LaunchDarkly tokens.
`
