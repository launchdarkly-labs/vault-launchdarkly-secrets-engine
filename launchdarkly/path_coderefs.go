package launchdarkly

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	token, err := CreateCodeRefsToken(config, projectName)
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

	resp.Secret.MaxTTL = config.MaxTTL * time.Second
	resp.Secret.TTL = config.TTL * time.Second

	return resp, nil
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
