package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var redirectURI = "http://localhost:8080/callback"

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	// rand.Seed(time.Now().UnixNano())
}

func login(auth *spotifyauth.Authenticator, ctx context.Context) (*oauth2.Token, error) {
	state := getRandomStr()
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	var code string
	fmt.Println("Please enter your code:")
	fmt.Scan(&code)

	token, err := auth.Exchange(ctx, code)
	if err != nil {
		return token, err
	}

	return token, nil
}

func main() {
	/* set up variables */
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)
	ctx := context.Background()

	/* OAuth login */
	var token *oauth2.Token
	if isCacheExist() {
		var err error
		token, err = readCache()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		var err error
		token, err = login(auth, ctx)
		if err != nil {
			log.Fatalln(err)
		}

		err = saveCache(token)
		if err != nil {
			log.Println("failed to save cache: ", err)
		} else {
			cachePath, _ := getCachePath()
			log.Println("token cache was saved to ", cachePath)
		}
	}

	/* prepare Client */
	httpClient := auth.Client(ctx, token)
	client := spotify.New(httpClient)

	/* use API */
	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range playlists.Playlists {
		createPlaylistDirectory(ctx, client, v)
	}
}

func createPlaylistDirectory(ctx context.Context, client *spotify.Client, playlist spotify.SimplePlaylist) error {
	// TODO: playlist.Name に /が含まれる場合を除く
	err := os.Mkdir(playlist.Name, os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	}

	playlistItemPage, err := client.GetPlaylistItems(ctx, playlist.ID)
	if err != nil {
		return err
	}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		// TODO: track.Name に /が含まれる場合を除く
		ioutil.WriteFile(playlist.Name+"/"+track.Name+".txt", []byte(fmt.Sprintf("%v\n", track)), 0666)
	}
	return nil
}

func saveCache(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cachePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func isCacheExist() bool {
	cachePath, err := getCachePath()
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

func readCache() (*oauth2.Token, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(cachePath)
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

func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir + "/.spotify-file-based-client.json", nil
}

func getRandomStr() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
