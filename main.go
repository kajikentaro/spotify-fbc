package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	/* ユーザー固有ではない情報の取得が可能 */
	result, err := client.GetRelatedArtists(ctx, "3WrFJ7ztbogyGnTHbHJFl2")
	if err != nil {
		log.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(result[0:2], "", "  ")
	fmt.Println(string(bytes))

	/* Client Credentials Flow だとユーザー固有の情報は取得できない */
	/*
		simplePlaylistPage, err := client.CurrentUsersPlaylists(ctx)
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Println(simplePlaylistPage)
	*/
}
