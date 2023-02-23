package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kajikentaro/spotify-file-based-client/client/logins"
	"github.com/kajikentaro/spotify-file-based-client/client/models"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var rootCmd = &cobra.Command{
	Use:   "spotifyfbc",
	Short: "spotifyfbc: Spotify file-based client",
	Long: `Spotify file-based client
  Edit your playlists by moving directories and file locations`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pullCmd)

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of spotifyfbc",
	Long:  `All software has versions. This is spotifyfilefbc's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.1")
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "TODO",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := genClient(ctx)

		/* generate model */
		model := models.NewModel(client, ctx)

		/* use API via model*/
		model.PullPlaylists()
	},
}

func genClient(ctx context.Context) *spotify.Client {
	var redirectURI = "http://localhost:8080/callback"
	/* set up variables */
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

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
	return client
}
