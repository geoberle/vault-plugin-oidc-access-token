package accesstoken

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
	"golang.org/x/oauth2"
)

func TestBackend_clientCreateAndReadRoundtrip(t *testing.T) {
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}
	b, _ := Backend(context.Background(), config, func(ctx context.Context, clientId string, clientSecret string, tokenUrl string) (*oauth2.Token, error) {
		return &oauth2.Token{
			AccessToken: "token",
		}, nil
	})

	clientId := "client"
	clientSecret := "clientSecret"
	tokenUrl := "http://localhost:8080/token"
	_, err := b.HandleRequest(context.Background(), createClientRequest(&config.StorageView, clientId, clientSecret, tokenUrl))
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	resp, err := b.HandleRequest(context.Background(), getClientRequest(&config.StorageView, "client"))
	if resp == nil {
		t.Fatalf("failed to read client: %s", err)
	}
	if resp.Data["client_id"].(string) != clientId {
		t.Fatalf("client_id is not correct: %s", err)
	}
	if resp.Data["client_secret"].(string) != clientSecret {
		t.Fatalf("client_secret is not correct: %s", err)
	}
	if resp.Data["token_url"].(string) != tokenUrl {
		t.Fatalf("token_url is not correct: %s", err)
	}
}

func TestBackend_clientCreate(t *testing.T) {
	tests := []struct {
		Name                        string
		ClientId                    string
		ClientSecret                string
		TokenUrl                    string
		FailGetToken                bool
		SleepSecondsTokenGeneration int
		Fail                        bool
	}{
		{
			"happy path",
			"client",
			"secret",
			"https://localhost:8080/token",
			false,
			0,
			false,
		},
		{
			"token generation fails",
			"client",
			"secret",
			"https://localhost:8080/token",
			true,
			0,
			true,
		},
		{
			"timeout token generation validation",
			"client",
			"secret",
			"https://localhost:8080/token",
			false,
			6,
			true,
		},
	}
	for _, tc := range tests {

		config := logical.TestBackendConfig()
		config.StorageView = &logical.InmemStorage{}
		b, _ := Backend(context.Background(), config, func(ctx context.Context, clientId string, clientSecret string, tokenUrl string) (*oauth2.Token, error) {
			if tc.SleepSecondsTokenGeneration > 0 {
				<-time.After(time.Duration(tc.SleepSecondsTokenGeneration) * time.Second)
			}
			if tc.FailGetToken {
				return nil, fmt.Errorf("failed to get token")
			}
			return &oauth2.Token{
				AccessToken: "token",
			}, nil
		})

		resp, err := b.HandleRequest(context.Background(), createClientRequest(&config.StorageView, tc.ClientId, tc.ClientSecret, tc.TokenUrl))
		if tc.Fail {
			if err == nil {
				t.Fatalf("expected an error for test %q", tc.Name)
			}
			continue
		} else if err != nil || (resp != nil && resp.IsError()) {
			t.Fatalf("bad: test name: %q\nresp: %#v\nerr: %v", tc.Name, resp, err)
		}
	}
}

func createClientRequest(storage *logical.Storage, clientId string, clientSecret string, tokenUrl string) *logical.Request {
	return &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("clients/%s", clientId),
		Storage:   *storage,
		Data: map[string]interface{}{
			"name":          clientId,
			"client_id":     clientId,
			"client_secret": clientSecret,
			"token_url":     tokenUrl,
		},
	}
}

func getClientRequest(storage *logical.Storage, clientId string) *logical.Request {
	return &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("clients/%s", clientId),
		Storage:   *storage,
	}
}
