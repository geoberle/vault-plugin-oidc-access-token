package accesstoken

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const clientBackendStoragePrefix = "client/"

func pathLisClients(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "clients/?$",

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathClientList,
		},

		HelpSynopsis:    pathClientHelpSyn,
		HelpDescription: pathClientHelpDesc,
	}
}

func pathClients(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "clients/" + framework.GenericNameWithAtRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the client.",
			},
			"client_id": {
				Type:        framework.TypeString,
				Description: "OIDC client ID.",
			},
			"client_secret": {
				Type:        framework.TypeString,
				Description: "OIDC client secret.",
			},
			"token_url": {
				Type:        framework.TypeString,
				Description: "URL to get the token access token from.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathClientRead,
			logical.UpdateOperation: b.pathClientCreate,
			logical.DeleteOperation: b.pathClientDelete,
		},

		HelpSynopsis:    pathClientHelpSyn,
		HelpDescription: pathClientHelpDesc,
	}
}

func (b *backend) Client(ctx context.Context, s logical.Storage, n string) (*clientEntry, error) {
	entry, err := s.Get(ctx, clientBackendStoragePrefix+n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result clientEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathClientRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	client, err := b.Client(ctx, req.Storage, data.Get("name").(string))
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"client_id":     client.ClientId,
			"client_secret": client.ClientSecret,
			"token_url":     client.TokenUrl,
		},
	}, nil
}

func (b *backend) pathClientCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	clientId := data.Get("client_id").(string)
	clientSecret := data.Get("client_secret").(string)
	tokenUrl := data.Get("token_url").(string)

	// check if the configuration data is valid by trying to get an access token
	if _, err := b.GenerateToken(ctx, clientId, clientSecret, tokenUrl, 5); err != nil {
		return logical.ErrorResponse("client configuration seems to be invalid to aquire an access token"), err
	}

	entry, err := logical.StorageEntryJSON(clientBackendStoragePrefix+name, &clientEntry{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		TokenUrl:     tokenUrl,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	var response *logical.Response
	return response, nil
}

func (b *backend) pathClientDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, clientBackendStoragePrefix+data.Get("name").(string))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathClientList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, clientBackendStoragePrefix)
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

type clientEntry struct {
	ClientId     string `json:"client_id" mapstructure:"client_id" structs:"client_id"`
	ClientSecret string `json:"client_secret" mapstructure:"client_secret" structs:"client_secret"`
	TokenUrl     string `json:"token_url" mapstructure:"token_url" structs:"token_url"`
}

const pathClientHelpSyn = `
Mange OIDC client configurations.
`

const pathClientHelpDesc = `
This path lets you manage the OIDC client configurations.

`
