package logins

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Login struct {
	ctx          context.Context
	token        *oauth2.Token
	clientId     string
	clientSecret string
	redirectURI  string
}

func (l *Login) IsLogin() bool {
	return l.token != nil
}

func GetAuth(redirectURI, clientID, clientSecret string) *spotifyauth.Authenticator {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate, spotifyauth.ScopePlaylistReadCollaborative, spotifyauth.ScopePlaylistModifyPrivate, spotifyauth.ScopePlaylistModifyPublic),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)
	return auth
}

func NewFromCache(ctx context.Context) (Login, bool) {
	if IsCacheExist() {
		cache, err := ReadCache()
		if err != nil {
			return Login{}, false
		}
		return NewLogin(ctx, cache.ClientId, cache.ClientSecret, cache.RedirectURI, cache.Token), true
	}
	return Login{}, false
}

func NewLogin(ctx context.Context, clientId, clientSecret, redirectURI string, token *oauth2.Token) Login {
	return Login{ctx: ctx, clientId: clientId, clientSecret: clientSecret, redirectURI: redirectURI, token: token}
}

func (l *Login) Login() error {
	state := getRandomStr()
	auth := GetAuth(l.redirectURI, l.clientId, l.clientSecret)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	ch := make(chan string)
	defer close(ch)

	codeCtx, codeCancel := context.WithCancel(l.ctx)

	if l.redirectURI == "http://localhost:8080/callback" {
		fmt.Println("Waiting a callback. You can also paste code:")
		// 'code'取得用のサーバーを立てる
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			s := r.URL.Query().Get("state")
			if code == "" {
				w.Write([]byte("Error: Cannot take code."))
			} else if state != s {
				w.Write([]byte("Error: query 'state' is wrong."))
			} else {
				// チャネルが閉じてない場合のみ送信する
				select {
				case <-codeCtx.Done():
					return
				default:
					ch <- code
				}
				w.Write([]byte("OAuth login was successful. You can close this window."))
			}
		})
		go func() {
			err := http.ListenAndServe(":8080", nil)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	} else {
		fmt.Println("Please enter your code query:")
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		code := scanner.Text()

		// チャネルが閉じてない場合のみ送信する
		select {
		case <-codeCtx.Done():
			return
		default:
			ch <- code
		}
	}()

	code := <-ch
	codeCancel()

	token, err := auth.Exchange(l.ctx, code)
	if err != nil {
		return err
	}
	l.token = token

	return nil
}

func (l *Login) GetClient() *spotify.Client {
	auth := GetAuth(l.redirectURI, l.clientId, l.clientSecret)
	httpClient := auth.Client(l.ctx, l.token)
	client := spotify.New(httpClient)
	return client
}

type Cache struct {
	Token        *oauth2.Token `json:"token"`
	ClientId     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	RedirectURI  string        `json:"redirect_uri"`
}

func (l *Login) SaveCache() error {
	cache := Cache{Token: l.token, ClientId: l.clientId, ClientSecret: l.clientSecret, RedirectURI: l.redirectURI}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(cachePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (l *Login) Logout() error {
	ctx := context.Background()
	withoutToken := NewLogin(ctx, l.clientId, l.clientSecret, l.redirectURI, nil)
	err := withoutToken.SaveCache()

	if err != nil {
		return err
	}

	return nil
}

func RemoveCache() error {
	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	err = os.Remove(cachePath)
	if err != nil {
		return err
	}

	return nil
}

func IsCacheExist() bool {
	cachePath, err := GetCachePath()
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

func ReadCache() (*Cache, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache Cache
	err = json.Unmarshal(b, &cache)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

func GetCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".spotify-fbc.json"), nil
}

func getRandomStr() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
