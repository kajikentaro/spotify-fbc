package models

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	dirNameToPL := map[string]PlaylistContent{}
	for _, e := range entries {
		reText := regexp.MustCompile(".txt$")
		if !reText.MatchString(e.Name()) || e.IsDir() {
			// .txtで終わらないファイル, ディレクトリの場合
			continue
		}

		// .txtで終わる名前のファイルの場合
		b, err := os.ReadFile(filepath.Join(fbcPath, e.Name()))
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", e.Name(), err)
		}
		p := unmarshalPlaylistContent(string(b))
		dirNameToPL[p.DirName] = p
	}

	localPLs := []PlaylistContent{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if v, isExist := dirNameToPL[e.Name()]; isExist {
			// プレイリスト情報txtが存在する場合
			localPLs = append(localPLs, v)
		} else {
			// プレイリスト情報txtが存在しない場合
			localPLs = append(localPLs, PlaylistContent{Name: e.Name(), DirName: e.Name()})
		}
	}

	remotePLs := []PlaylistContent{}
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	for _, v := range playlists.Playlists {
		remotePLs = append(remotePLs, PlaylistContent{Id: v.ID.String(), Name: v.Name})
	}

	// 新規作成するべきプレイリストの検索
	toAddPLs := []PlaylistContent{}
	for _, v := range localPLs {
		if v.Id == "" {
			toAddPLs = append(toAddPLs, v)
			continue
		}

		isRemoteExist := false
		for _, r := range remotePLs {
			if v.Id == r.Id {
				isRemoteExist = true
				break
			}
		}
		if !isRemoteExist {
			toAddPLs = append(toAddPLs, v)
			continue
		}
	}

	// 削除するべきプレイリストの検索
	toRemovePLs := []PlaylistContent{}
	for _, v := range remotePLs {
		isLocalExist := false
		for _, l := range localPLs {
			if v.Id == l.Id {
				isLocalExist = true
				break
			}
		}
		if !isLocalExist {
			toRemovePLs = append(toRemovePLs, v)
			continue
		}
	}

	fmt.Println("### ADD ###")
	for _, v := range toAddPLs {
		fmt.Println(v)
	}
	fmt.Println("\n### DELETE ###")
	for _, v := range toRemovePLs {
		fmt.Println(v)
	}

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

func (m *model) CreatePlaylistDirectory(playlist PlaylistContent) error {
	// generate a playlist detail file
	textContent, err := playlist.marshal()
	if err != nil {
		return err
	}
	ioutil.WriteFile(SPOTIFY_PLAYLIST_ROOT+"/"+playlist.DirName+".txt", []byte(textContent), 0666)

	// generate a playlist directory
	err = os.Mkdir(SPOTIFY_PLAYLIST_ROOT+"/"+playlist.DirName, os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	}

	// generate a track file in the directory
	playlistItemPage, err := m.client.GetPlaylistItems(m.ctx, spotify.ID(playlist.Id))
	if err != nil {
		return err
	}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		trackContent := TrackContent{
			Id:      track.ID.String(),
			Name:    track.Name,
			Artist:  joinArtistText(track.Artists),
			Album:   track.Album.Name,
			Seconds: strconv.Itoa(track.Duration),
			Isrc:    track.ExternalIDs["isrc"],
		}
		textContent, err := trackContent.marshal()
		if err != nil {
			return err
		}
		ioutil.WriteFile(SPOTIFY_PLAYLIST_ROOT+"/"+playlist.DirName+"/"+replaceBannedCharacter(track.Name)+".txt", []byte(textContent), 0666)
	}
	return nil
}

func replaceBannedCharacter(path string) string {
	reg := regexp.MustCompile("[\\\\/:*?\"<>|]")
	return reg.ReplaceAllString(path, " ")
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

		err := m.CreatePlaylistDirectory(PlaylistContent{Id: v.ID.String(), Name: v.Name, DirName: uniqueName})
		if err != nil {
			return err
		}
	}
	return nil
}
