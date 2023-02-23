package cmd

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kajikentaro/spotify-file-based-client/client/logins"
	"github.com/kajikentaro/spotify-file-based-client/client/models"
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

func Execute() {
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

	login := logins.NewLogin(ctx, auth)

	/* OAuth login and get client*/
	var token *oauth2.Token
	if login.IsCacheExist() {
		var err error
		token, err = login.ReadCache()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		var err error
		token, err = login.Login()
		if err != nil {
			log.Fatalln(err)
		}

		err = login.SaveCache(token)
		if err != nil {
			log.Println("failed to save cache: ", err)
		} else {
			cachePath, _ := login.GetCachePath()
			log.Println("token cache was saved to ", cachePath)
		}
	}
	client := login.GetClient(token)

	/* generate model */
	model := models.NewModel(client, ctx)

	/* use API via model*/
	model.PullPlaylists()
}
