package launchdarkly

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ldapi "github.com/launchdarkly/api-client-go"

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
					"ttl": {
						Type:        framework.TypeDurationSecond,
						Description: "Default lease for generated keys. If <= 0, will use system default.",
					},
					"max_ttl": {
						Type:        framework.TypeDurationSecond,
						Description: "Maximum time a service account key is valid for. If <= 0, will use system default.",
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
		Secrets: []*framework.Secret{
			b.programmaticAPIKeys(),
		},
	}

	return b
}

func (b *backend) Close() {
	b.clientMutex.Lock()
	defer b.clientMutex.Unlock()
}

func (b *backend) programmaticAPIKeys() *framework.Secret {
	return &framework.Secret{
		Type: programmaticAPIKey,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: "Programmatic API Key Public Key",
			},
		},
		Renew:  b.programmaticAPIKeysRenew,
		Revoke: b.programmaticAPIKeyRevoke,
	}
}

func (b *backend) programmaticAPIKeyRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	programmaticAPIKeyIDRaw, ok := req.Secret.InternalData["api_key_id"]
	if !ok {
		return nil, fmt.Errorf("secret is missing programmatic api key id internal data")
	}

	programmaticAPIKeyID, ok := programmaticAPIKeyIDRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing programmatic api key id internal data")
	}

	TypeRaw, ok := req.Secret.InternalData["credential_type"]
	if !ok {
		return nil, fmt.Errorf("secret is missing credential_type internal data")
	}

	KeyType, ok := TypeRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing credential_type internal data")
	}

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	switch KeyType {
	case "api":
		DeleteRoleToken(config, programmaticAPIKeyID)
	case "rac":
		DeleteRelayToken(config, programmaticAPIKeyID)
	}

	return nil, nil
}

func (b *backend) readCredentials(ctx context.Context, s logical.Storage, credentialName string, credentialType string) (*tokenCredentialEntry, error) {
	var roleEntry tokenCredentialEntry
	entry, err := s.Get(ctx, credentialType+"/"+credentialName)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		if err := entry.DecodeJSON(&roleEntry); err != nil {
			return nil, err
		}
		return &roleEntry, nil
	}
	return nil, nil
}

func (b *backend) programmaticAPIKeysRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = config.TTL
	resp.Secret.MaxTTL = config.MaxTTL
	return resp, nil
}

const backendHelp = `
The LaunchDarkly secrets engine generates LaunchDarkly tokens.
`
const programmaticAPIKey = `LDApiKey`

type tokenCredentialEntry struct {
	Token  ldapi.Token   `json:"token"`
	TTL    time.Duration `json:"ttl"`
	MaxTTL time.Duration `json:"max_ttl"`
}

func (r tokenCredentialEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"token":   r.Token,
		"ttl":     r.TTL.Seconds(),
		"max_ttl": r.MaxTTL.Seconds(),
	}
	return respData
}
