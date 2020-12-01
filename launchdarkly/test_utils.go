package launchdarkly

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	LDApiKey = "LD_API_KEY"
)

type testEnv struct {
	AccessToken string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage
}

func newTestAccEnv() (*testEnv, error) {
	ctx := context.Background()
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: time.Hour,
			MaxLeaseTTLVal:     time.Hour,
		},
	}
	b, err := Factory(ctx, conf)
	if err != nil {
		return nil, err
	}

	return &testEnv{
		AccessToken: os.Getenv(LDApiKey),
		Backend:     b,
		Context:     ctx,
		Storage:     &logical.InmemStorage{},
	}, nil
}

func (e *testEnv) addConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"access_token": e.AccessToken,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
}
