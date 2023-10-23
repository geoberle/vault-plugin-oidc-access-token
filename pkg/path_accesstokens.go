package accesstoken

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathAccessTokens(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "accesstoken/" + framework.GenericNameWithAtRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the client.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathReadAccessToken,
		},

		HelpSynopsis:    pathAccessTokenHelp,
		HelpDescription: pathAccessTokenHelp,
	}
}

func (b *backend) pathReadAccessToken(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)

	// Get the client
	client, err := b.Client(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return logical.ErrorResponse(fmt.Sprintf("unknown client: %s", name)), nil
	}

	// Get access token
	token, err := b.GenerateToken(ctx, client.ClientId, client.ClientSecret, client.TokenUrl, 5)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("failed to aquire OIDC access token for %s from %s", name, client.TokenUrl)), err
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"access_token": token.AccessToken,
			"expiry":       token.Expiry,
		},
	}, nil
}

const pathAccessTokenHelp = `
Request an access token for an OIDC client configuration.
`
