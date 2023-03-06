package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kajikentaro/spotify-file-based-client/client/logins"
	"github.com/kajikentaro/spotify-file-based-client/client/repositories"
	"github.com/kajikentaro/spotify-file-based-client/client/services"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var SPOTIFY_PLAYLIST_ROOT = "spotify-fbc"

func Execute() {
	log.Println("start")
	fmt.Println()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println()
	log.Println("done")
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean up unused playlist entity txt",
	Long:  `clean up unused playlist entity txt`,
	Run: func(cmd *cobra.Command, args []string) {
		deleted, err := repositories.CleanUpPlaylistContent(SPOTIFY_PLAYLIST_ROOT)
		for d := range deleted {
			log.Println(d, "was deleted.")
		}
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var rootCmd = &cobra.Command{
	Use:   "spotifyfbc",
	Short: "spotifyfbc: Spotify file-based client",
	Long: `Spotify file-based client
  Edit your playlists by moving directories and file locations`,
}

// TODO: 特定プレイリストのみのpush機能
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Synchronize your local files and directories with your spotify account.",
	Long:  `Synchronize your local files and directories with your spotify account`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := genClient(ctx)
		model := services.NewModel(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		if err := model.PushPlaylists(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "TODO",
	Long:  `login`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := genClient(ctx)
		model := services.NewModel(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		if err := model.ComparePlaylists(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "TODO",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		_, login := genClient(ctx)
		login.RemoveCache()
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "TODO",
	Long:  `login`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		genClient(ctx)
	},
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
		client, _ := genClient(ctx)
		model := services.NewModel(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		if err := model.PullPlaylists(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func genClient(ctx context.Context) (*spotify.Client, logins.Login) {
	var redirectURI = "http://localhost:8080/callback"
	/* set up variables */
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate, spotifyauth.ScopePlaylistReadCollaborative, spotifyauth.ScopePlaylistModifyPrivate),
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
	return client, login
}
