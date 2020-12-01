package launchdarkly

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type launchdarklyConfig struct {
	AccessToken string `json:"access_token"`
	BaseUri     string `json:"base_uri`
}

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Validate we didn't get extraneous fields
	if err := validateFields(req, data); err != nil {
		return nil, logical.CodedError(422, err.Error())
	}

	accessToken := data.Get("access_token").(string)
	if !strings.HasPrefix(accessToken, "api-") {
		return nil, errors.New("Token should start with `api-`")
	}

	config, err := b.config(ctx, req.Storage)

	if err != nil {
		return nil, err
	}
	if config == nil {
		config = &launchdarklyConfig{
			BaseUri:     "https://app.launchdarkly.com",
			AccessToken: "",
		}
	}

	if err := config.Update(data); err != nil {
		return logical.ErrorResponse(fmt.Sprintf("could not update config: %v", err)), nil
	}

	entry, err := logical.StorageEntryJSON("config", config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// Invalidate existing clients so they read the new configuration
	b.Close()

	return nil, nil
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	resp := make(map[string]interface{})

	if v := config.AccessToken; v != "" {
		resp["access_token"] = v
	}

	if v := config.BaseUri; v != "" {
		resp["base_uri"] = v
	}

	return &logical.Response{
		Data: resp,
	}, nil
}

func (config *launchdarklyConfig) Update(data *framework.FieldData) error {
	accessToken := data.Get("access_token").(string)
	if len(accessToken) > 0 {
		config.AccessToken = accessToken
	}

	baseUri := data.Get("base_uri").(string)
	if len(baseUri) > 0 {
		config.BaseUri = baseUri
	}
	return nil
}

func (b *backend) config(ctx context.Context, s logical.Storage) (*launchdarklyConfig, error) {
	config := &launchdarklyConfig{}
	entry, err := s.Get(ctx, "config")

	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	if err := entry.DecodeJSON(config); err != nil {
		return nil, err
	}
	return config, nil
}
