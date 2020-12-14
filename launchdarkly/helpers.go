package launchdarkly

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	ldapi "github.com/launchdarkly/api-client-go"
)

func GenericLDKeyWithAtRegex(name string) string {
	return fmt.Sprintf("(?P<%s>\\w(([\\w-._]+)?\\w)?)", name)
}

// validateFields verifies that no bad arguments were given to the request.
func validateFields(req *logical.Request, data *framework.FieldData) error {
	var unknownFields []string
	for k := range req.Data {
		if _, ok := data.Schema[k]; !ok {
			unknownFields = append(unknownFields, k)
		}
	}

	if len(unknownFields) > 0 {
		// Sort since this is a human error
		sort.Strings(unknownFields)

		return fmt.Errorf("unknown fields: %q", unknownFields)
	}

	return nil
}

const (
	//APIVersion = "20191212"
	APIVersion = "beta"
)

// Client is used by the provider to access the ld API.
type Client struct {
	apiKey  string
	apiHost string
	ld      *ldapi.APIClient
	ctx     context.Context
}

func newClient(config *launchdarklyConfig, oauth bool) (*Client, error) {
	basePath := fmt.Sprintf(`%s/api/v2`, config.BaseUri)

	cfg := &ldapi.Configuration{
		BasePath:      basePath,
		DefaultHeader: make(map[string]string),
		UserAgent:     fmt.Sprintf("launchdarkly-vault-provider/%s", Version),
	}

	cfg.AddDefaultHeader("LD-API-Version", APIVersion)

	ctx := context.WithValue(context.Background(), ldapi.ContextAPIKey, ldapi.APIKey{
		Key: config.AccessToken,
	})
	if oauth {
		ctx = context.WithValue(context.Background(), ldapi.ContextAccessToken, config.AccessToken)
	}

	return &Client{
		apiKey:  config.AccessToken,
		apiHost: basePath,
		ld:      ldapi.NewAPIClient(cfg),
		ctx:     ctx,
	}, nil
}

func currentMillis() int64 {
	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000
	expire := 60*60*12*1000 + millis

	return expire
}

const (
	MAX_409_RETRIES = 5
	MAX_429_RETRIES = 10
)

func handleRateLimit(apiCall func() (interface{}, *http.Response, error)) (interface{}, *http.Response, error) {
	obj, res, err := apiCall()
	for retryCount := 0; res != nil && res.StatusCode == http.StatusTooManyRequests && retryCount < MAX_429_RETRIES; retryCount++ {
		log.Println("[DEBUG] received a 429 Too Many Requests error. retrying")
		resetStr := res.Header.Get("X-RateLimit-Reset")
		resetInt, parseErr := strconv.ParseInt(resetStr, 10, 64)
		if parseErr != nil {
			log.Println("[DEBUG] could not parse X-RateLimit-Reset header. Sleeping for a random interval.")
			randomRetrySleep()
		} else {
			resetTime := time.Unix(0, resetInt*int64(time.Millisecond))
			sleepDuration := time.Until(resetTime)

			// We have observed situations where LD-s retry header results in a negative sleep duration. In this case,
			// multiply the duration by -1 and add a random 200-500ms
			if sleepDuration <= 0 {
				log.Printf("[DEBUG] received a negative rate limit retry duration of %s. Sleeping for an additional 200-500ms", sleepDuration)
				sleepDuration = -1*sleepDuration + getRandomSleepDuration()
			}
			log.Println("[DEBUG] sleeping", sleepDuration)
			time.Sleep(sleepDuration)
		}
		obj, res, err = apiCall()
	}
	return obj, res, err

}

var randomRetrySleepSeeded = false

// Sleep for a random interval between 200ms and 500ms
func getRandomSleepDuration() time.Duration {
	if !randomRetrySleepSeeded {
		rand.Seed(time.Now().UnixNano())
	}
	n := rand.Intn(300) + 200
	return time.Duration(n) * time.Millisecond
}

// Sleep for a random interval between 200ms and 500ms
func randomRetrySleep() {
	if !randomRetrySleepSeeded {
		rand.Seed(time.Now().UnixNano())
	}
	n := rand.Intn(300) + 200
	time.Sleep(time.Duration(n) * time.Millisecond)
}

// handleLdapiErr extracts the error message and body from a ldapi.GenericSwaggerError or simply returns the
// error  if it is not a ldapi.GenericSwaggerError
func handleLdapiErr(err error) error {
	if err == nil {
		return nil
	}
	if swaggerErr, ok := err.(ldapi.GenericSwaggerError); ok {
		return fmt.Errorf("%s: %s", swaggerErr.Error(), string(swaggerErr.Body()))
	}
	return err
}

func configCheck(config *launchdarklyConfig) error {
	if config == nil {
		return errors.New("Please write your AccessToken to launchdarkly/config")
	}
	if config.AccessToken == "" && config.BaseUri == "" {
		return errors.New("Access Token and BaseUri need to be set")
	}
	if config.AccessToken == "" {
		return errors.New("LaunchDarkly Access Token needs to be set")
	}
	if config.BaseUri == "" {
		return errors.New("LaunchDarkly BaseUri needs to be set")
	}
	return nil
}

func getConfig(b *backend, ctx context.Context, storage logical.Storage) (*launchdarklyConfig, error) {
	config, err := b.config(ctx, storage)
	if err != nil {
		return nil, err
	}

	if err := configCheck(config); err != nil {
		return nil, err
	}

	return config, nil
}
