package logins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type login struct {
	ctx  context.Context
	auth *spotifyauth.Authenticator
}

func NewLogin(ctx context.Context, auth *spotifyauth.Authenticator) login {
	return login{ctx: ctx, auth: auth}
}

func (l *login) Login() (*oauth2.Token, error) {
	state := getRandomStr()
	url := l.auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	var code string
	fmt.Println("Please enter your code:")
	fmt.Scan(&code)

	token, err := l.auth.Exchange(l.ctx, code)
	if err != nil {
		return token, err
	}

	return token, nil
}

func (l *login) GetClient(token *oauth2.Token) *spotify.Client {
	httpClient := l.auth.Client(l.ctx, token)
	client := spotify.New(httpClient)
	return client
}

func (l *login) SaveCache(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	cachePath, err := l.GetCachePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(cachePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (l *login) IsCacheExist() bool {
	cachePath, err := l.GetCachePath()
	if err != nil {
		return false
	}

	if _, err := os.Stat(cachePath); errors.Is(err, os.ErrNotExist) {
		// authorized token does not yet exist (maybe on their first use)
		return false
	} else if err != nil {
		// error occured
		return false
	} else {
		return true
	}
}

func (l *login) ReadCache() (*oauth2.Token, error) {
	cachePath, err := l.GetCachePath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	err = json.Unmarshal(b, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (l *login) GetCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".spotify-file-based-client.json"), nil
}

func getRandomStr() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
