package launchdarkly

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/antihax/optional"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ldapi "github.com/launchdarkly/api-client-go"
)

func (b *backend) pathCoderefsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	logger := hclog.New(&hclog.LoggerOptions{})
	projectName := data.Get("project").(string)
	logger.Debug(projectName)
	if projectName == "" {
		return nil, errors.New("project is required")
	}
	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var tokenEntry ldapi.Token
	entry, err := req.Storage.Get(ctx, "coderefs/"+projectName)
	if err != nil {
		return nil, err
	}

	if entry != nil {
		if err := entry.DecodeJSON(&tokenEntry); err != nil {
			return nil, err
		}
	}
	var token *ldapi.Token
	if entry != nil {
		token, err = ResetCodeRefsToken(config, tokenEntry.Id)
		if err != nil {
			return nil, err
		}
	} else {
		token, err = CreateCodeRefsToken(config, projectName)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	newEntry, err := logical.StorageEntryJSON("coderefs/"+projectName, token)
	if err != nil {
		return nil, err
	}

	err = req.Storage.Put(ctx, newEntry)
	if err != nil {
		return nil, err
	}

	resp := b.Secret(programmaticAPIKey).Response(map[string]interface{}{
		"token": token.Token,
	}, map[string]interface{}{
		"api_key_id":      token.Id,
		"credential_type": "api",
		"secret_type":     "coderefs",
	})

	resp.Secret.MaxTTL = config.MaxTTL
	resp.Secret.TTL = config.TTL

	return resp, nil
}

func (b *backend) pathCoderefsDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	name := data.Get("project").(string)

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var tokenEntry ldapi.Token
	entry, err := req.Storage.Get(ctx, "coderefs/"+name)
	if err != nil {
		return nil, err
	}

	if entry != nil {
		if err := entry.DecodeJSON(&tokenEntry); err != nil {
			return nil, err
		}
	}

	if entry != nil {
		err := DeleteAccessToken(config, name, tokenEntry.Id)
		if err != nil {
			return nil, handleLdapiErr(err)
		}

		err = req.Storage.Delete(ctx, "coderefs/"+name)
		if err != nil {
			return nil, handleLdapiErr(err)
		}
	}
	return nil, nil
}

// CreateCodeRefsToken uses launchdarkly API to create an API token
func CreateCodeRefsToken(config *launchdarklyConfig, project string) (*ldapi.Token, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})

	// Prepare request
	resource := fmt.Sprintf(`code-reference-repository/%s`, project)

	statement := ldapi.Statement{
		Resources: []string{resource},
		Actions:   []string{"*"},
		Effect:    "allow",
	}

	newToken := ldapi.TokenBody{
		Name:              project,
		InlineRole:        []ldapi.Statement{statement},
		ServiceToken:      true,
		DefaultApiVersion: 20191212,
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

// ResetCodeRefsToken uses launchdarkly API to create an API token
func ResetCodeRefsToken(config *launchdarklyConfig, id string) (*ldapi.Token, error) {
	expiry := optional.NewInt64(currentMillis())

	opts := ldapi.ResetTokenOpts{
		Expiry: expiry,
	}

	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	token, _, err := client.ld.AccessTokensApi.ResetToken(client.ctx, id, &opts)
	if err != nil {
		return nil, handleLdapiErr(err)
	}

	return &token, nil

}

// DeleteRelayToken uses the LaunchDarkly API to delete a Relay Auto Congfig token
func DeleteAccessToken(config *launchdarklyConfig, name string, tokenId string) error {
	//logger := hclog.New(&hclog.LoggerOptions{})

	client, err := newClient(config, false)
	if err != nil {
		return handleLdapiErr(err)
	}

	_, _, err = handleRateLimit(func() (interface{}, *http.Response, error) {
		res, err := client.ld.AccessTokensApi.DeleteToken(client.ctx, tokenId)
		return nil, res, err
	})
	if err != nil {
		return handleLdapiErr(err)
	}

	return nil
}
