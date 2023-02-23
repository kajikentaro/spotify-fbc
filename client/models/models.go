package models

import (
	"context"
	"fmt"

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
	/*
		err := os.Mkdir(playlist.Name, os.ModePerm)
		if os.IsExist(err) {
			log.Println(playlist.Name, "is already created")
		}
	*/

	playlistItemPage, err := m.client.GetPlaylistItems(m.ctx, playlist.ID)
	if err != nil {
		return err
	}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		// TODO: track.Name に /が含まれる場合を除く
		// ioutil.WriteFile(playlist.Name+"/"+track.Name+".txt", []byte(fmt.Sprintf("%v\n", track)), 0666)
		fmt.Println(track.Name)
	}
	return nil
}

func (m *model) PullPlaylists() error {
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	for _, v := range playlists.Playlists {
		err := m.CreatePlaylistDirectory(v)
		if err != nil {
			return err
		}
	}
	return nil
}
