package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kajikentaro/spotify-fbc/models"
	service_compares "github.com/kajikentaro/spotify-fbc/services/compares"
	"github.com/kajikentaro/spotify-fbc/services/interfaces"
	"github.com/kajikentaro/spotify-fbc/services/uniques"
)

type service struct {
	repository interfaces.Repository
}

func NewService(repository interfaces.Repository) service {
	return service{repository: repository}
}

func getFileStem(fileName string) (string, error) {
	r := regexp.MustCompile(`(.*)\.txt$`)
	fileStem := r.FindStringSubmatch(fileName)
	if len(fileStem) < 2 {
		return "", fmt.Errorf("property file_name is invalid. must end with .txt: '%s'", fileName)
	}
	return fileStem[1], nil
}

func (m *service) recreateTrackTxt(usedFileStem *uniques.Unique, playlist models.PlaylistContent, res []models.TrackContent) {
	// 現在存在する楽曲txtの一覧を作成
	for _, w := range res {
		fileStem, _ := getFileStem(w.FileName)
		usedFileStem.Add(fileStem)
	}

	for _, w := range res {
		fmt.Println("  +", w.Name)
		/* 成功した場合は楽曲txtを作り直す */

		// 削除
		err := m.repository.RemoveTrackContent(playlist.DirName, w)
		if err != nil {
			// 削除に失敗した場合
			fmt.Fprintln(os.Stderr, "failed to remove the old track content: ", filepath.Join(playlist.DirName, w.FileName))
			continue
		}
		fileStem, _ := getFileStem(w.FileName)
		usedFileStem.Delete(fileStem)

		// 作成
		stemName := replaceBannedCharacter(w.Name)
		w.FileName = usedFileStem.Take(stemName) + ".txt"
		err = m.repository.CreateTrackContent(playlist.DirName, w)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create a new track content: ", filepath.Join(playlist.DirName, w.FileName))
		}
	}
}

func (m *service) addRemoteTrack(playlist models.PlaylistContent, tracks []models.TrackContent) error {
	// 曲をリモートのプレイリストに追加
	c := make(chan []models.TrackContent)

	go func() {
		usedFileStem := uniques.NewUnique()
		for cc := range c {
			// 楽曲txtを作り直す
			m.recreateTrackTxt(usedFileStem, playlist, cc)
		}
	}()
	err := m.repository.AddRemoteTrack(playlist.Id, tracks, c)
	close(c)
	if err != nil {
		return err
	}
	return nil
}

func (m *service) syncLocalPlaylistWithRemote(v service_compares.PlaylistTrackDiff) (changed bool, err error) {
	pl := v.Playlist

	// プレイリストの作成/削除
	if pl.DiffState == service_compares.LocalOnly {
		// プレイリストをリモートに作成
		resPlaylist, err := m.repository.CreateRemotePlaylist(pl.V.DirName)
		if err != nil {
			return false, err
		}
		pl.V = resPlaylist
		// プレイリストファイルをローカルに新規に生成する
		resPlaylist.DirName = pl.V.DirName
		m.repository.CreatePlaylistContent(resPlaylist)
		changed = true
		fmt.Println("+", pl.V.DirName)
	}
	if pl.DiffState == service_compares.RemoteOnly {
		// プレイリストをリモートから削除
		err := m.repository.RemoveRemotePlaylist(pl.V)
		if err != nil {
			return false, err
		}
		changed = true
		fmt.Println("-", pl.V.Name)

		// 削除の場合はここで終わり
		return changed, nil
	}
	if pl.DiffState == service_compares.Both {
		fmt.Println(" ", pl.V.Name)
	}

	localOnlyTracks := []models.TrackContent{}
	remoteOnlyTracks := []models.TrackContent{}
	for _, w := range v.Tracks {
		if w.DiffState == service_compares.LocalOnly {
			localOnlyTracks = append(localOnlyTracks, w.V)
		}
		if w.DiffState == service_compares.RemoteOnly {
			remoteOnlyTracks = append(remoteOnlyTracks, w.V)
		}
	}

	// 曲をプレイリストに追加
	if err := m.addRemoteTrack(pl.V, localOnlyTracks); err != nil {
		return false, err
	}
	for _, w := range localOnlyTracks {
		fmt.Println("  +", w.FileName)
		changed = true
	}

	// 曲をリモートのプレイリストから削除
	err = m.repository.RemoveRemoteTrack(pl.V, remoteOnlyTracks)
	if err != nil {
		return false, err
	}
	for _, w := range remoteOnlyTracks {
		fmt.Println("  -", w.Name)
		changed = true
	}

	return changed, nil
}

func (m *service) OverwritePlaylists() error {
	fmt.Fprintln(os.Stderr, "now loading ...")
	changed := false

	// プレイリストの差分を検出
	compare := service_compares.NewCompare(m.repository)
	diff, err := compare.CompareAllPlaylistWithRemote()
	if err != nil {
		return err
	}

	for _, v := range diff {
		_changed, err := m.syncLocalPlaylistWithRemote(v)
		if err != nil {
			return err
		}
		changed = changed || _changed
	}

	// 後片付け: 不要なプレイリストテキストを消去
	deleted, err := m.repository.CleanUpPlaylistContent()
	for _, d := range deleted {
		fmt.Fprintln(os.Stderr, d, "was deleted.")
	}
	if err != nil {
		return err
	}

	if !changed {
		fmt.Println("\nthere was no change on remote")
	}
	return nil
}

func (m *service) CreatePlaylistDirectory(playlist models.PlaylistContent) error {
	// generate a playlist detail file
	err := m.repository.CreatePlaylistContent(playlist)
	if err != nil {
		return err
	}

	// generate a playlist directory
	err = m.repository.CreatePlaylistDirectory(playlist)
	if err != nil {
		return err
	}

	playlistTrack, err := m.repository.FetchRemotePlaylistTrack(playlist.Id)
	if err != nil {
		return err
	}

	// generate a track file in the directory
	usedTrackNames := uniques.NewUnique()
	for _, track := range playlistTrack {
		fileStem := replaceBannedCharacter(track.Name)
		track.FileName = usedTrackNames.Take(fileStem) + ".txt"
		m.repository.CreateTrackContent(playlist.DirName, track)
	}
	return nil
}

func replaceBannedCharacter(path string) string {
	reg := regexp.MustCompile("[\\\\/:*?\"<>|]")
	return reg.ReplaceAllString(path, " ")
}

func (m *service) PullPlaylists() error {
	playlists, err := m.repository.FetchRemotePlaylistContent()
	if err != nil {
		return err
	}
	if err := m.repository.CreateRootDir(); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	usedPlaylistName := uniques.NewUnique()
	for _, v := range playlists {
		// define a unduplicated directory name
		name := replaceBannedCharacter(v.Name)
		uniqueName := usedPlaylistName.Take(name)
		v.DirName = uniqueName

		err := m.CreatePlaylistDirectory(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *service) PushSpecificPlaylist(playlistName string) error {
	fmt.Fprintln(os.Stderr, "now loading ...")

	// 一旦プレイリストだけのの差分を検出
	compare := service_compares.NewCompare(m.repository)
	allPlaylists, err := compare.CalcDiffPlaylist()
	if err != nil {
		return err
	}

	//　該当プレイリストを検索
	var playlist *service_compares.WithDiffState[models.PlaylistContent]
	for _, v := range allPlaylists {
		if v.V.DirName == playlistName {
			playlist = &v
			break
		}
	}
	if playlist == nil {
		return fmt.Errorf("playlist '%s' not found", playlistName)
	}

	diff, err := compare.CompareSinglePlaylistWithRemote(*playlist)
	if err != nil {
		return err
	}

	changed, err := m.syncLocalPlaylistWithRemote(diff)
	if err != nil {
		return err
	}

	if !changed {
		fmt.Println("\nthere was no change on remote")
	}
	return nil
}
