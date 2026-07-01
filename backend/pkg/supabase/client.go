package supabase

import (
	"context"
	"fmt"

	"github.com/supabase-community/supabase-go"
)

type Client struct {
	client *supabase.Client
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Aud      string `json:"aud"`
	Role     string `json:"role"`
}

func NewClient(supabaseURL, supabaseKey string) *Client {
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create supabase client: %v", err))
	}
	return &Client{client: client}
}

func (c *Client) VerifyJWT(ctx context.Context, token string) (*User, error) {
	user, err := c.client.Auth.WithToken(token).GetUser()
	if err != nil {
		return nil, fmt.Errorf("verify JWT: %w", err)
	}

	return &User{
		ID:    user.ID.String(),
		Email: user.Email,
		Aud:   user.Aud,
		Role:  user.Role,
	}, nil
}
