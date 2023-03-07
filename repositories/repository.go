package repositories

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

type Repository struct {
	client   *spotify.Client
	ctx      context.Context
	rootPath string
}

func NewRepository(client *spotify.Client, ctx context.Context, rootPath string) *Repository {
	return &Repository{client: client, ctx: ctx, rootPath: rootPath}
}
