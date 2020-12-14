package launchdarkly

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ldapi "github.com/launchdarkly/api-client-go"
)

func (b *backend) pathRelayWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	name := data.Get("name").(string)
	if name == "" {
		return nil, errors.New("name is required")
	}

	policy := data.Get("inline_policy").(string)
	if policy == "" {
		return nil, errors.New("name is required")
	}

	var tokenPolicy ldapi.Policy
	if err := json.Unmarshal([]byte(policy), &tokenPolicy); err != nil {
		return nil, err
	}

	newEntry, err := logical.StorageEntryJSON("relay/policy/"+name, tokenPolicy)
	if err != nil {
		return nil, err
	}

	req.Storage.Put(ctx, newEntry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"inline_policy": policy,
		},
	}, nil
}

func (b *backend) pathRelayRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	name := data.Get("name").(string)

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var tokenPolicy ldapi.Policy

	policyEntry, err := req.Storage.Get(ctx, "relay/policy/"+name)
	if policyEntry == nil {
		return nil, nil
	}
	if policyEntry != nil {
		if err := policyEntry.DecodeJSON(&tokenPolicy); err != nil {
			return nil, err
		}
	}

	token, err := CreateRelayToken(config, name, tokenPolicy)
	if err != nil {
		return nil, handleLdapiErr(err)
	}

	resp := b.Secret(programmaticAPIKey).Response(map[string]interface{}{
		"token": token.FullKey,
	}, map[string]interface{}{
		"api_key_id":      token.Id,
		"credential_type": "rac",
		"secret_type":     "relay",
	})
	resp.Secret.MaxTTL = config.MaxTTL
	resp.Secret.TTL = config.TTL

	return resp, nil
}

// CreatelaunchdarklyToken uses LaunchDarkly API to create a Relay Auto Config token
func CreateRelayToken(config *launchdarklyConfig, name string, policy ldapi.Policy) (*ldapi.RelayProxyConfig, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})

	// Prepare request
	newToken := ldapi.RelayProxyConfigBody{
		Name:   name,
		Policy: []ldapi.Policy{policy},
	}
	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	tokenRaw, _, err := handleRateLimit(func() (interface{}, *http.Response, error) {
		return client.ld.RelayProxyConfigurationsApi.PostRelayAutoConfig(client.ctx, newToken)
	})
	if err != nil {
		return nil, handleLdapiErr(err)
	}
	token := tokenRaw.(ldapi.RelayProxyConfig)

	return &token, nil
}

func (b *backend) pathRelayDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	name := data.Get("name").(string)

	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var relayEntry ldapi.RelayProxyConfig
	entry, err := req.Storage.Get(ctx, "relayEntry/"+name)

	if entry != nil {
		if err := entry.DecodeJSON(&relayEntry); err != nil {
			return nil, err
		}
	}

	if entry != nil {
		err := DeleteRelayToken(config, relayEntry.Id)
		if err != nil {
			return nil, handleLdapiErr(err)
		}

		err = req.Storage.Delete(ctx, "relayEntry/"+name)
		if err != nil {
			return nil, handleLdapiErr(err)
		}
	}
	return nil, nil
}

// DeleteRelayToken uses the LaunchDarkly API to delete a Relay Auto Congfig token
func DeleteRelayToken(config *launchdarklyConfig, tokenId string) error {
	//logger := hclog.New(&hclog.LoggerOptions{})

	client, err := newClient(config, false)
	if err != nil {
		return handleLdapiErr(err)
	}

	_, _, err = handleRateLimit(func() (interface{}, *http.Response, error) {
		res, err := client.ld.RelayProxyConfigurationsApi.DeleteRelayProxyConfig(client.ctx, tokenId)
		return nil, res, err
	})
	if err != nil {
		return handleLdapiErr(err)
	}

	return nil
}
