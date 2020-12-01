package launchdarkly

import (
	"context"
	"errors"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ldapi "github.com/launchdarkly/api-client-go"
)

func (b *backend) pathProjectEnvRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	//logger := hclog.New(&hclog.LoggerOptions{})
	projectKey := data.Get("project").(string)
	envKey := data.Get("env").(string)

	if projectKey == "" {
		return nil, errors.New("project is required")
	}

	if envKey == "" {
		return nil, errors.New("env is required")
	}
	config, err := getConfig(b, ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	var envEntry map[string]interface{}
	entry, err := req.Storage.Get(ctx, "project/"+projectKey)
	if entry != nil {
		if err := entry.DecodeJSON(&envEntry); err != nil {
			return nil, err
		}
	}
	if envEntry != nil {
		return &logical.Response{
			Data: envEntry,
		}, nil
	}

	project, _, err := client.ld.ProjectsApi.GetProject(client.ctx, projectKey)
	if err != nil {
		return nil, handleLdapiErr(err)
	}
	env := []ldapi.Environment{}

	// Looking for the environment that matches the path. Only 1 should match.
	for i := range project.Environments {
		if project.Environments[i].Key == envKey {
			env = append(env, project.Environments[i])
			break
		}
	}

	envData := map[string]interface{}{
		"sdk":       env[0].ApiKey,
		"mobile":    env[0].MobileKey,
		"client_id": env[0].Id,
	}

	newEntry, err := logical.StorageEntryJSON("project/"+projectKey, envData)
	if err != nil {
		return nil, err
	}

	req.Storage.Put(ctx, newEntry)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: envData,
	}, nil
}

func (b *backend) pathProjectReset(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	projectKey := data.Get("project").(string)
	envKey := data.Get("env").(string)
	resetType := data.Get("type").(string)
	switch resetType {
	case
		"mobile",
		"sdk":
	default:
		return nil, errors.New("Reset needs to be sdk or mobile")
	}

	if projectKey == "" {
		return nil, errors.New("project is required")
	}
	if envKey == "" {
		return nil, errors.New("env is required")
	}
	config, err := b.config(ctx, req.Storage)
	if err != nil {
		return &logical.Response{
			Data: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}

	client, err := newClient(config, false)
	if err != nil {
		return nil, err
	}

	var env ldapi.Environment
	switch reset := resetType; reset {
	case "mobile":
		env, _, err = client.ld.EnvironmentsApi.ResetEnvironmentMobileKey(client.ctx, projectKey, envKey, nil)
		if err != nil {
			return nil, handleLdapiErr(err)
		}
	case "sdk":
		env, _, err = client.ld.EnvironmentsApi.ResetEnvironmentSDKKey(client.ctx, projectKey, envKey, nil)
		if err != nil {
			return nil, handleLdapiErr(err)
		}
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"sdk":       env.ApiKey,
			"mobile":    env.MobileKey,
			"client_id": env.Id,
		},
	}, nil
}
