package models

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/zmb3/spotify/v2"
)

type model struct {
	client *spotify.Client
	ctx    context.Context
}

func NewModel(client *spotify.Client, ctx context.Context) model {
	return model{client: client, ctx: ctx}
}

func (m *model) CreatePlaylistDirectory(playlist spotify.SimplePlaylist) error {
	// TODO: playlist.Name に /が含まれる場合を除く
	err := os.Mkdir(playlist.Name, os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	}

	playlistItemPage, err := m.client.GetPlaylistItems(m.ctx, playlist.ID)
	if err != nil {
		return err
	}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		// TODO: track.Name に /が含まれる場合を除く
		textContent := getTextContent(track)
		ioutil.WriteFile(playlist.Name+"/"+track.Name+".txt", []byte(textContent), 0666)
	}
	return nil
}

func getTextContent(track *spotify.FullTrack) string {
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
	for _, v := range playlists.Playlists[:1] {
		err := m.CreatePlaylistDirectory(v)
		if err != nil {
			return err
		}
	}
	return nil
}
