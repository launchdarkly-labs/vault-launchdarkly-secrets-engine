package launchdarkly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestCoderefsToken(t *testing.T) {

	acceptanceTestEnv, err := newTestAccEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.addConfig)
	t.Run("read coderefs token", acceptanceTestEnv.readCoderefsToken)
	t.Run("reset coderefs token", acceptanceTestEnv.resetCoderefsToken)
	t.Run("delete coderefs token", acceptanceTestEnv.deleteCoderefsToken)
}

func (e *testEnv) readCoderefsToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "coderefs/testVaultRepo",
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

func (e *testEnv) resetCoderefsToken(t *testing.T) {
	reqFirst := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "coderefs/testVaultRepo",
		Storage:   e.Storage,
	}
	respFirst, err := e.Backend.HandleRequest(e.Context, reqFirst)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", respFirst, err)
	}
	if respFirst == nil {
		t.Fatal("expected a response")
	}
	if respFirst.Data["token"] == "" || !strings.HasPrefix(respFirst.Data["token"].(string), "api-") {
		t.Fatal("token does not match expected format")
	}
	currentToken := respFirst.Data["token"]

	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "coderefs/testVaultRepo",
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

	if resp.Data["token"] == currentToken {
		t.Fatal("token has not been reset")
	}

}

func (e *testEnv) deleteCoderefsToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "coderefs/testVaultRepo",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	fmt.Println(resp)
	if resp != nil {
		t.Fatal("token not successfully removed")
	}
}
