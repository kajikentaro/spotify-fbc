package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kajikentaro/spotify-fbc/logins"
	"github.com/kajikentaro/spotify-fbc/repositories"
	"github.com/kajikentaro/spotify-fbc/services"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var SPOTIFY_PLAYLIST_ROOT = "spotify-fbc"

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
	Short: "Clean up unused playlist entity txt",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		deleted, err := repository.CleanUpPlaylistContent()
		for d := range deleted {
			fmt.Fprintln(os.Stderr, d, "wad deleted.")
		}
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var rootCmd = &cobra.Command{
	Use:   "spotify-fbc",
	Short: "Spotify file-based client",
	Long: `Spotify file-based client: 
Edit your playlists by moving directories and file locations`,
}

// TODO: 特定プレイリストのみのpush機能
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Synchronize your local files and directories with your spotify account",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		model := services.NewService(repository)
		if !askForConfirmation("WARNING: Your remote spotify playlist will be replaced") {
			return
		}
		if err := model.PushPlaylists(); err != nil {
			log.Fatalln(err)
		}
	},
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare local playlists with your spotify account and print the difference",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		service := services.NewService(repository)
		if err := service.Compare(); err != nil {
			log.Fatalln(err)
		}
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete user-specific data such as OAuth token and Client ID excluding music txt",
	Run: func(cmd *cobra.Command, args []string) {
		logins.RemoveCache()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your spotify account excluding API keys",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		_, login := setup(ctx)
		if err := login.Logout(); err != nil {
			log.Fatalln(err)
		}
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Perform login process",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		setup(ctx)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of spotify-fbc",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.1.2")
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download playlists that your spotify account has. All of your existing local playlists will be overwritten",
	Long: `Download playlists that your spotify account has.
All of your existing local playlists will be overwritten.
If you have local-specific files, It will be remained`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, _ := setup(ctx)
		repository := repositories.NewRepository(client, ctx, SPOTIFY_PLAYLIST_ROOT)
		model := services.NewService(repository)
		if err := model.PullPlaylists(); err != nil {
			log.Fatalln(err)
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
			fmt.Fprintln(os.Stderr, "failed to save cache: ", err)
		} else {
			cachePath, _ := logins.GetCachePath()
			fmt.Fprintln(os.Stderr, "token cache was saved to ", cachePath)
		}
	}

	return login.GetClient(), login
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
