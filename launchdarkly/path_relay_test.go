package launchdarkly

import (
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestAccessRelayCreds(t *testing.T) {

	acceptanceTestEnv, err := newTestAccEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.addConfig)
	t.Run("write relay policy", acceptanceTestEnv.writeRelayPolicy)
	t.Run("read relay token", acceptanceTestEnv.readRelayToken)
	t.Run("read relay no path", acceptanceTestEnv.readNonExistantRelayToken)
}

func (e *testEnv) readRelayToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "relay/testVault",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Data["token"] == "" || !strings.HasPrefix(resp.Data["token"].(string), "rel-") {
		t.Fatal("token does not match expected format")
	}
}

func (e *testEnv) readNonExistantRelayToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "relay/test-no-key",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp != nil {
		t.Fatal("expected a response")
	}
}

func (e *testEnv) writeRelayPolicy(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "relay/policy",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"name":          "testVault",
			"inline_policy": `{"resources":["proj/*:env/*"], "actions": ["*"], "effect":"allow"}`,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}

	if !strings.Contains(resp.Data["inline_policy"].(string), `{"resources":["proj/*:env/*"], "actions": ["*"], "effect":"allow"}`) {
		t.Fatal("policy does not match")
	}
}
