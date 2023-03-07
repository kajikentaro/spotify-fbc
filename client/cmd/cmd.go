package cmd

import (
	"bufio"
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
	godotenv.Load()

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean up unused playlist entity txt",
	Long:  `clean up unused playlist entity txt`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		deleted, err := repository.CleanUpPlaylistContent()
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
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		model := services.NewModel(repository)
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
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		model := services.NewModel(repository)
		if err := model.ComparePlaylists(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete user-specific data such as OAuth token and Client ID without music txt.",
	Long:  `Delete user-specific data such as OAuth token and Client ID without music txt.`,
	Run: func(cmd *cobra.Command, args []string) {
		logins.RemoveCache()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "TODO",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		_, login := setup(ctx)
		if err := login.Logout(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "TODO",
	Long:  `login`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		setup(ctx)
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
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		model := services.NewModel(repository)
		if err := model.PullPlaylists(); err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func setup(ctx context.Context) (*spotify.Client, logins.Login) {
	login, isOk := logins.NewFromCache(ctx)

	// APIの各キーが登録されていない場合
	if !isOk {
		// set up variables
		clientID := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		redirectUri := os.Getenv("REDIRECT_URI")
		if clientID == "" {
			fmt.Println("Please visit https://developer.spotify.com/dashboard/applications and do 'CREATE AN APP'.")
			fmt.Println("Enter your Client ID:")
			clientID = readLine()
			fmt.Println("Enter your Cilent Secret:")
			clientSecret = readLine()
			fmt.Println("Enter your Redirect URI: (default http://localhost:8080/callback)")
			redirectUri = readLine()
		}
		if redirectUri == "" {
			redirectUri = "http://localhost:8080/callback"
		}
		login = logins.NewLogin(ctx, clientID, clientSecret, redirectUri, nil)
	}

	// ログアウト状態の場合
	if !login.IsLogin() {
		err := login.Login()
		if err != nil {
			log.Fatalln(err)
		}

		// save cache
		err = login.SaveCache()
		if err != nil {
			log.Println("failed to save cache: ", err)
		} else {
			cachePath, _ := logins.GetCachePath()
			log.Println("token cache was saved to ", cachePath)
		}
	}

	return login.GetClient(), login
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}
