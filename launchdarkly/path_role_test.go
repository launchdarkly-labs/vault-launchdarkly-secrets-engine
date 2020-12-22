package launchdarkly

import (
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestRoleToken(t *testing.T) {

	acceptanceTestEnv, err := newTestAccEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.addConfig)
	t.Run("read role token", acceptanceTestEnv.readRoleToken)
}

func (e *testEnv) readRoleToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "role/test-vault-role",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Data["token"] == "" || !strings.HasPrefix(resp.Data["token"].(string), "api-") {
		t.Fatal("token does not match expected format")
	}
}
