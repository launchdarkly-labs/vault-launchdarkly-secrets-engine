package launchdarkly

import (
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestProjectKeys(t *testing.T) {

	acceptanceTestEnv, err := newTestAccEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.addConfig)
	t.Run("read project keys", acceptanceTestEnv.readProjectKeys)
	t.Run("read project keys", acceptanceTestEnv.resetSdkKey)
}

const relayTestPath = "project/vault-integration/test"

func (e *testEnv) readProjectKeys(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      relayTestPath,
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Data["sdk"] == "" || !strings.HasPrefix(resp.Data["sdk"].(string), "sdk-") {
		t.Fatal("environment sdk key does not match expected format")
	}

	if resp.Data["mobile"] == "" || !strings.HasPrefix(resp.Data["mobile"].(string), "mob-") {
		t.Fatal("environment mobile key does not match expected format")
	}

	if resp.Data["client_id"] == "" {
		t.Fatal("client_id is empty")
	}

}

func (e *testEnv) resetSdkKey(t *testing.T) {
	reqCurrent := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      relayTestPath,
		Storage:   e.Storage,
	}
	respCurrent, err := e.Backend.HandleRequest(e.Context, reqCurrent)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", respCurrent, err)
	}
	if respCurrent == nil {
		t.Fatal("expected a response")
	}
	currentSdkKey := respCurrent.Data["sdk"]

	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      relayTestPath + "/reset/sdk",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Data["sdk"] == currentSdkKey {
		t.Fatal("sdk key not reset")
	}
}

func (e *testEnv) resetMobileKey(t *testing.T) {
	reqCurrent := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      relayTestPath,
		Storage:   e.Storage,
	}
	respCurrent, err := e.Backend.HandleRequest(e.Context, reqCurrent)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", respCurrent, err)
	}
	if respCurrent == nil {
		t.Fatal("expected a response")
	}
	currentSdkKey := respCurrent.Data["mobile"]

	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      relayTestPath + "/reset/mobile",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Data["mobile"] == currentSdkKey {
		t.Fatal("mobile key not reset")
	}
}
