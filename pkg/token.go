package accesstoken

import (
	"context"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type TokenGetter func(context.Context, string, string, string) (*oauth2.Token, error)

func GetOIDCAccessToken(ctx context.Context, clientId string, clientSecret string, tokenUrl string) (*oauth2.Token, error) {
	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (b *backend) GenerateToken(ctx context.Context, clientId string, clientSecret string, tokenUrl string, timeoutSeconds int) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	tokenChan := make(chan *oauth2.Token)
	errChan := make(chan error)

	go func() {
		token, err := b.tokenGenerator(ctx, clientId, clientSecret, tokenUrl)
		if err != nil {
			errChan <- err
			return
		}
		tokenChan <- token
	}()

	select {
	case token := <-tokenChan:
		return token, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
