package launchdarkly

// func testBackend(tb testing.TB) (*backend, logical.Storage) {
// 	tb.Helper()

// 	config := logical.TestBackendConfig()
// 	config.StorageView = &logical.InmemStorage{}

// 	b, err := Factory(context.Background(), config)
// 	if err != nil {
// 		tb.Fatal(err)
// 	}
// 	return b.(*backend), config.StorageView
// }

// func testConfigCreate(t *testing.T, d map[string]interface{}) logicaltest.TestStep {
// 	return logicaltest.TestStep{
// 		Operation: logical.CreateOperation,
// 		Path:      "config",
// 		Data:      d,
// 	}
// }

// func TestBackend(t *testing.T) {
// t.Run("info", func(t *testing.T) {
// 	t.Parallel()

// 	b, storage := testBackend(t)
// 	resp, err := b.HandleRequest(context.Background(), &logical.Request{
// 		Storage:   storage,
// 		Operation: logical.ReadOperation,
// 		Path:      "info",
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if v, exp := resp.Data["commit"].(string), version.GitCommit; v != exp {
// 		t.Errorf("expected %q to be %q", v, exp)
// 	}
// })

// 	t.Run("config", func(t *testing.T) {
// 		t.Parallel()

// 		data := map[string]interface{}{
// 			"accessToken": "123456789",
// 			"baseUri":     "https://app.launchdarkly.com",
// 		}

// 		b, storage := testBackend(t)
// 		_, err := b.HandleRequest(context.Background(), &logical.Request{
// 			Storage:   storage,
// 			Operation: logical.UpdateOperation,
// 			Path:      "config",
// 			Data:      data,
// 		})
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		resp, err := b.HandleRequest(context.Background(), &logical.Request{
// 			Storage:   storage,
// 			Operation: logical.ReadOperation,
// 			Path:      "config",
// 		})
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		if v, exp := resp.Data["accessToken"].(string), "123456789"; v != exp {
// 			t.Errorf("expected %q to be %q", v, exp)
// 		}

// 		if v, exp := resp.Data["baseUri"].(string), "https://app.launchdarkly.com"; v != exp {
// 			t.Errorf("expected %q to be %q", v, exp)
// 		}
// 	})
// 	t.Run("coderefs", func(t *testing.T) {
// 		t.Parallel()

// 		data := map[string]interface{}{
// 			"projectName": "testRepo",
// 		}

// 		b, storage := testBackend(t)
// 		_, err := b.HandleRequest(context.Background(), &logical.Request{
// 			Storage:   storage,
// 			Operation: logical.ReadOperation,
// 			Path:      "coderefs/testRepo",
// 			Data:      data,
// 		})
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		resp, err := b.HandleRequest(context.Background(), &logical.Request{
// 			Storage:   storage,
// 			Operation: logical.ReadOperation,
// 			Path:      "coderefs/testRepo",
// 		})
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		if v, exp := resp.Data["accessToken"].(string), "123456789"; v != exp {
// 			t.Errorf("expected %q to be %q", v, exp)
// 		}

// 		if v, exp := resp.Data["baseUri"].(string), "https://app.launchdarkly.com"; v != exp {
// 			t.Errorf("expected %q to be %q", v, exp)
// 		}
// 	})
// }
