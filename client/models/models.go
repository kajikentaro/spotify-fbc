package models

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zmb3/spotify/v2"
)

type model struct {
	client *spotify.Client
	ctx    context.Context
}

var SPOTIFY_PLAYLIST_ROOT = "spotify-fbc"

func NewModel(client *spotify.Client, ctx context.Context) model {
	return model{client: client, ctx: ctx}
}

func (m *model) ComparePlaylists(fbcPath string) error {
	entries, err := os.ReadDir(fbcPath)
	if err != nil {
		return err
	}

	localDirsSet := map[string]struct{}{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		localDirsSet[e.Name()] = struct{}{}
	}

	remoteDirsSet := map[string]struct{}{}
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	for _, v := range playlists.Playlists {
		remoteDirsSet[v.Name] = struct{}{}
	}

	fmt.Println(localDirsSet)
	fmt.Println(remoteDirsSet)

	/*
		playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
		if err != nil {
			return err
		}
		for _, v := range playlists.Playlists[:1] {
			err := m.CreatePlaylistDirectory(v)
			if err != nil {
				return err
			}
		}
	*/
	return nil
}

func (m *model) CreatePlaylistDirectory(uniqueName string, playlist spotify.SimplePlaylist) error {
	// generate a playlist detail file
	textContent := getPlaylistContent(uniqueName, playlist)
	ioutil.WriteFile(SPOTIFY_PLAYLIST_ROOT+"/"+uniqueName+".txt", []byte(textContent), 0666)

	// generate a playlist directory
	err := os.Mkdir(SPOTIFY_PLAYLIST_ROOT+"/"+replaceBannedCharacter(playlist.Name), os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	}

	// generate a track file in the directory
	playlistItemPage, err := m.client.GetPlaylistItems(m.ctx, playlist.ID)
	if err != nil {
		return err
	}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		textContent := getTrackContent(track)
		ioutil.WriteFile(SPOTIFY_PLAYLIST_ROOT+"/"+uniqueName+"/"+replaceBannedCharacter(track.Name)+".txt", []byte(textContent), 0666)
	}
	return nil
}

func replaceBannedCharacter(path string) string {
	reg := regexp.MustCompile("[\\\\/:*?\"<>|]")
	return reg.ReplaceAllString(path, " ")
}

func getPlaylistContent(dirName string, playlist spotify.SimplePlaylist) string {
	dd := [][]string{
		{"id", playlist.ID.String()},
		{"name", playlist.Name},
		{"dir_name", dirName},
	}

	text := ""
	for _, d := range dd {
		text += d[0] + " " + d[1] + "\n"
	}

	return text
}

func getTrackContent(track *spotify.FullTrack) string {
	dd := [][]string{
		{"id", track.ID.String()},
		{"name", track.Name},
		{"artist", joinArtistText(track.Artists)},
		{"album", track.Album.Name},
		{"seconds", strconv.Itoa(track.Duration)},
		{"isrc", track.ExternalIDs["isrc"]},
	}

	text := ""
	for _, d := range dd {
		text += d[0] + " " + d[1] + "\n"
	}

	return text
}

func joinArtistText(artists []spotify.SimpleArtist) string {
	text := []string{}
	for _, a := range artists {
		text = append(text, a.Name)
	}
	return strings.Join(text, ", ")
}

func (m *model) PullPlaylists() error {
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	os.Mkdir(SPOTIFY_PLAYLIST_ROOT, os.ModePerm)
	usedPlaylistName := map[string]struct{}{}

	for _, v := range playlists.Playlists[:] {
		// define a unduplicated directory name
		name := replaceBannedCharacter(v.Name)
		uniqueName := name
		for i := 2; i < 1e7; i++ {
			if _, isDuplicated := usedPlaylistName[uniqueName]; isDuplicated {
				uniqueName = name + " " + strconv.Itoa(i)
			} else {
				break
			}
		}
		usedPlaylistName[name] = struct{}{}

		err := m.CreatePlaylistDirectory(uniqueName, v)
		if err != nil {
			return err
		}
	}
	return nil
}
