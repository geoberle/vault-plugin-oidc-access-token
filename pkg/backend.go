package accesstoken

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend
	tokenGenerator TokenGetter
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	return Backend(ctx, conf, GetOIDCAccessToken)
}

func Backend(ctx context.Context, conf *logical.BackendConfig, tokenGenerator TokenGetter) (*backend, error) {
	var b backend

	b.tokenGenerator = tokenGenerator
	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(backendHelp),
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"client/",
			},
		},
		Paths: []*framework.Path{
			pathLisClients(&b),
			pathClients(&b),
			pathAccessTokens(&b),
		},
		Secrets: []*framework.Secret{},
	}

	if conf == nil {
		return nil, fmt.Errorf("configuration passed into backend is nil")
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return &b, nil
}

const backendHelp = `
The OIDC access token backend generates valid access tokens for an OIDC client.
`
