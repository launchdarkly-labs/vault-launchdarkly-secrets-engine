package launchdarkly

import (
	"context"
	"errors"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ldapi "github.com/launchdarkly/api-client-go"
)

type TokenKey struct {
	Name string
}

type RoleName struct {
	Name string
}

func (b *backend) pathRoleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	roleName := data.Get("customrole").(string)
	tokenName := data.Get("name").(string)
	if roleName == "" {
		return nil, errors.New("name is required")
	}

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var roleEntry ldapi.Token
	entry, err := req.Storage.Get(ctx, "role/"+roleName)
	if entry != nil {
		if err := entry.DecodeJSON(&roleEntry); err != nil {
			return nil, err
		}
	}

	token, err := CreateRoleToken(config, roleName, tokenName)
	if err != nil {
		return nil, err
	}

	resp := b.Secret(programmaticAPIKey).Response(map[string]interface{}{
		"token": token.Token,
	}, map[string]interface{}{
		"api_key_id":      token.Id,
		"credential_type": "api",
		"secret_type":     "role",
	})

	return resp, nil
}

func (b *backend) readRoleCredentials(ctx context.Context, s logical.Storage, credentialName string) (*tokenCredentialEntry, error) {
	var roleEntry tokenCredentialEntry
	entry, err := s.Get(ctx, "role/"+credentialName)
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

func (b *backend) pathRoleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	roleName := data.Get("customrole").(string)
	tokenName := data.Get("name").(string)
	if roleName == "" {
		return nil, errors.New("name is required")
	}

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var roleEntry ldapi.Token
	entry, err := req.Storage.Get(ctx, "role/"+roleName)
	if entry != nil {
		if err := entry.DecodeJSON(&roleEntry); err != nil {
			return nil, err
		}
	}

	token, err := CreateRoleToken(config, roleName, tokenName)
	if err != nil {
		return nil, err
	}

	newEntry, err := logical.StorageEntryJSON("role/"+roleName, token)
	if err != nil {
		return nil, err
	}

	req.Storage.Put(ctx, newEntry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"token": token.Token,
		},
	}, nil
}

// CreatelaunchdarklyToken uses launchdarkly API to create an API token
func CreateRoleToken(config *launchdarklyConfig, role string, name string) (*ldapi.Token, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})

	// Prepare request
	newToken := ldapi.TokenBody{
		Name:          name,
		CustomRoleIds: []string{role},
		ServiceToken:  true,
	}
	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	token, _, err := client.ld.AccessTokensApi.PostToken(client.ctx, newToken)
	if err != nil {
		return nil, handleLdapiErr(err)
	}

	return &token, nil
}

func DeleteRoleToken(config *launchdarklyConfig, id string) (*ldapi.Token, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})

	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	_, err = client.ld.AccessTokensApi.DeleteToken(client.ctx, id)
	if err != nil {
		return nil, handleLdapiErr(err)
	}

	return nil, nil
}
