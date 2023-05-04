package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kajikentaro/spotify-fbc/models"
	"github.com/kajikentaro/spotify-fbc/repositories"
	service_compares "github.com/kajikentaro/spotify-fbc/services/compares"
	"github.com/kajikentaro/spotify-fbc/services/uniques"
)

type service struct {
	repository *repositories.Repository
}

func NewService(repository *repositories.Repository) service {
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

func (m *service) Compare() error {
	fmt.Fprintln(os.Stderr, "now loading ...")

	compare := service_compares.NewCompare(m.repository)

	diff, err := compare.CompareAll()
	if err != nil {
		return err
	}

	for _, v := range diff.LocalOnly {
		fmt.Println("+", v.Playlist.DirName)
		for _, w := range v.Tracks {
			fmt.Println("  +", w.FileName)
		}
	}

	for _, v := range diff.RemoteOnly {
		fmt.Println("-", v.Name)
	}

	for _, v := range diff.Both {
		fmt.Println(" ", v.Playlist.Name)

		for _, w := range v.TracksLocalOnly {
			fmt.Println("  +", w.FileName)
		}
		for _, w := range v.TracksRemoteOnly {
			fmt.Println("  -", w.Name)
		}
	}

	if len(diff.RemoteOnly)+len(diff.LocalOnly) == 0 {
		fmt.Println("\nthere is no difference")
	}

	return nil
}

func (m *service) PushPlaylists() error {
	fmt.Fprintln(os.Stderr, "now loading ...")
	isChange := false

	// プレイリストの差分を検出
	compare := service_compares.NewCompare(m.repository)
	diff, err := compare.CompareAll()
	if err != nil {
		return err
	}

	// 新規追加プレイリストをpushする
	for _, v := range diff.LocalOnly {
		// プレイリストをリモートに作成
		resPlaylist, err := m.repository.CreateRemotePlaylist(v.Playlist.DirName)
		if err != nil {
			return err
		}
		isChange = true
		fmt.Println("+", v.Playlist.DirName)

		// プレイリストファイルをローカルに新規に生成する
		resPlaylist.DirName = v.Playlist.DirName
		m.repository.CreatePlaylistContent(resPlaylist)

		// 曲をプレイリストに追加
		if err := m.addRemoteTrack(resPlaylist, v.Tracks); err != nil {
			return err
		}
	}

	for _, v := range diff.RemoteOnly {
		// プレイリストをリモートから削除
		err := m.repository.RemoveRemotePlaylist(v)
		if err != nil {
			return err
		}
		isChange = true
		fmt.Println("-", v.Name)
	}

	for _, v := range diff.Both {
		isChange = true
		fmt.Println(" ", v.Playlist.Name)

		// 曲をプレイリストに追加
		if err := m.addRemoteTrack(v.Playlist, v.TracksLocalOnly); err != nil {
			return err
		}

		// 曲をリモートのプレイリストから削除
		err = m.repository.RemoveRemoteTrack(v.Playlist, v.TracksRemoteOnly)
		if err != nil {
			return err
		}
		for _, w := range diff.RemoteOnly {
			fmt.Println("  -", w.Name)
		}
	}

	// 後片付け: 不要なプレイリストテキストを消去
	deleted, err := m.repository.CleanUpPlaylistContent()
	for _, d := range deleted {
		fmt.Fprintln(os.Stderr, d, "was deleted.")
	}
	if err != nil {
		return err
	}

	if !isChange {
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
