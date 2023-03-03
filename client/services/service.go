package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
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

	// プレイリスト情報txtファイルを読み込み
	dirNameToPL := map[string]models.PlaylistContent{}
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
		p := models.UnmarshalPlaylistContent(string(b))
		dirNameToPL[p.DirName] = p
	}

	// ディレクトリを "プレイリスト情報txtファイル" の情報と関連付けて, 配列として保存
	localPLs := []models.PlaylistContent{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if v, isExist := dirNameToPL[e.Name()]; isExist {
			// プレイリスト情報txtが存在する場合
			localPLs = append(localPLs, v)
		} else {
			// プレイリスト情報txtが存在しない場合
			localPLs = append(localPLs, models.PlaylistContent{Name: e.Name(), DirName: e.Name()})
		}
	}

	// リモートのプレイリストを配列で取得
	remotePLs := []models.PlaylistContent{}
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	for _, v := range playlists.Playlists {
		remotePLs = append(remotePLs, models.PlaylistContent{Id: v.ID.String(), Name: v.Name})
	}

	idToPlaylist := map[string]models.PlaylistContent{}
	for _, v := range localPLs {
		idToPlaylist[v.Id] = v
	}
	for _, v := range remotePLs {
		idToPlaylist[v.Id] = v
	}

	// localにあれば +1, remoteにあれば +2する
	// 1ならlocalのみ, 2ならremoteのみ, 3なら両方に存在することになる
	playlistState := map[string]int{}
	localOnly := []models.PlaylistContent{}
	for _, v := range localPLs {
		if v.Id == "" {
			localOnly = append(localOnly, v)
			continue
		}
		playlistState[v.Id] += 1
	}
	for _, v := range remotePLs {
		playlistState[v.Id] += 2
	}

	toAddPLs := localOnly
	toRemovePLs := []models.PlaylistContent{}
	indefinitePLs := []models.PlaylistContent{}
	for id, bit := range playlistState {
		if bit == 1 {
			toAddPLs = append(toAddPLs, idToPlaylist[id])
		}
		if bit == 2 {
			toRemovePLs = append(toRemovePLs, idToPlaylist[id])
		}
		if bit == 3 {
			indefinitePLs = append(indefinitePLs, idToPlaylist[id])
		}
	}

	for _, v := range toAddPLs {
		fmt.Println("+", v.Name)
		tracks, err := readLocalPlaylistTrack(filepath.Join(fbcPath, v.DirName))
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  +", w.FileName)
		}
	}

	for _, v := range toRemovePLs {
		fmt.Println("-", v.Name)
		tracks, err := readLocalPlaylistTrack(filepath.Join(fbcPath, v.DirName))
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  -", w.FileName)
		}
	}

	for _, v := range indefinitePLs {
		fmt.Println("?", v.Name)
		tracks, err := readLocalPlaylistTrack(filepath.Join(fbcPath, v.DirName))
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  ?", w.FileName)
		}
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

func readLocalPlaylistTrack(dirPath string) ([]models.TrackContent, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", dirPath, err)
	}

	// プレイリスト情報txtファイルを読み込み
	tracks := []models.TrackContent{}
	for _, e := range entries {
		reText := regexp.MustCompile(".txt$")
		if !reText.MatchString(e.Name()) || e.IsDir() {
			// .txtで終わらないファイル, ディレクトリの場合
			continue
		}
		content, err := os.ReadFile(filepath.Join(dirPath, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %w", filepath.Join(dirPath, e.Name()), err)
		}
		t := models.UnmarshalTrackContent(string(content))
		if t.FileName == "" {
			// ユーザーが新規作成したTrackのtxtにはおそらくfile_nameプロパティが無い
			t.FileName = e.Name()
		}
		if t.FileName != e.Name() {
			log.Printf("Warning: a file_name property was incorrect. The property in the file was '%s', but path was '%s'.", t.FileName, e.Name())
			t.FileName = e.Name()
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func (m *model) CreatePlaylistDirectory(playlist models.PlaylistContent) error {
	// generate a playlist detail file
	textContent := playlist.Marshal()
	os.WriteFile(filepath.Join(SPOTIFY_PLAYLIST_ROOT, playlist.DirName+".txt"), []byte(textContent), 0666)

	// generate a playlist directory
	err := os.Mkdir(filepath.Join(SPOTIFY_PLAYLIST_ROOT, playlist.DirName), os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	}

	// generate a track file in the directory
	playlistItemPage, err := m.client.GetPlaylistItems(m.ctx, spotify.ID(playlist.Id))
	if err != nil {
		return err
	}
	usedTrackNames := map[string]struct{}{}
	for _, playlistItem := range playlistItemPage.Items {
		track := playlistItem.Track.Track
		fileName := unique(&usedTrackNames, replaceBannedCharacter(track.Name)) + ".txt"
		trackContent := models.TrackContent{
			Id:       track.ID.String(),
			Name:     track.Name,
			Artist:   joinArtistText(track.Artists),
			Album:    track.Album.Name,
			Seconds:  strconv.Itoa(track.Duration),
			Isrc:     track.ExternalIDs["isrc"],
			FileName: fileName,
		}
		textContent := trackContent.Marshal()
		os.WriteFile(filepath.Join(SPOTIFY_PLAYLIST_ROOT, playlist.DirName, fileName), []byte(textContent), 0666)
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
		uniqueName := unique(&usedPlaylistName, name)

		err := m.CreatePlaylistDirectory(models.PlaylistContent{Id: v.ID.String(), Name: v.Name, DirName: uniqueName})
		if err != nil {
			return err
		}
	}
	return nil
}

// stemNameがすでにusedSetに存在する場合は末尾に連番の数字を足したものを返す
/* 例:
 * usedSet := map[string]struct{}{}
 * res := unique(usedSet, "hoge")
 * // res is "hoge"
 * res := unique(usedSet, "hoge")
 * // res is "hoge 2"
 */
func unique(usedSet *map[string]struct{}, stemName string) string {
	uniqueName := stemName
	for i := 2; i < 1e7; i++ {
		if _, isDuplicated := (*usedSet)[uniqueName]; isDuplicated {
			uniqueName = stemName + " " + strconv.Itoa(i)
		} else {
			break
		}
	}
	(*usedSet)[uniqueName] = struct{}{}
	return uniqueName
}
