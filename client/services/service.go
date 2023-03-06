package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
	"github.com/kajikentaro/spotify-file-based-client/client/repositories"
	"github.com/kajikentaro/spotify-file-based-client/client/services/uniques"
)

type model struct {
	repository *repositories.Repository
}

func NewModel(repository *repositories.Repository) model {
	return model{repository: repository}
}

func getFileStem(fileName string) (string, error) {
	r := regexp.MustCompile(`(.*)\.txt$`)
	fileStem := r.FindStringSubmatch(fileName)
	if len(fileStem) < 2 {
		return "", fmt.Errorf("property file_name is invalid. must end with .txt: '%s'", fileName)
	}
	return fileStem[1], nil
}

func (m *model) recreateTrackTxt(playlist models.PlaylistContent, res []repositories.EditTrackRes) {

	// 現在存在する楽曲txtの一覧を作成
	usedFileStem := uniques.NewUnique()
	for _, w := range res {
		fileStem, _ := getFileStem(w.FileName)
		usedFileStem.Add(fileStem)
	}

	for _, w := range res {
		if w.IsOk {
			fmt.Println("  +", w.Name)
			/* 成功した場合は楽曲txtを作り直す */

			// 削除
			err := m.repository.RemoveTrackContent(playlist.DirName, w.TrackContent)
			if err != nil {
				// 削除に失敗した場合
				fmt.Println("failed to remove the old track content: ", filepath.Join(playlist.DirName, w.TrackContent.FileName))
				continue
			}
			fileStem, _ := getFileStem(w.FileName)
			usedFileStem.Delete(fileStem)

			// 作成
			stemName := replaceBannedCharacter(w.Name)
			w.FileName = usedFileStem.Take(stemName) + ".txt"
			err = m.repository.CreateTrackContent(playlist.DirName, w.TrackContent)
			if err != nil {
				fmt.Println("failed to create a new track content: ", filepath.Join(playlist.DirName, w.FileName))
			}
		} else {
			fmt.Println("   ", w.FileName, w.Message)
		}
	}
}

func (m *model) PushPlaylists() error {
	// プレイリストの差分を検出
	diff, err := m.calcDiffPlaylist()
	if err != nil {
		return err
	}

	// 新規追加プレイリストをpushする
	for _, v := range diff.localOnly {
		// プレイリストをリモートに作成
		resPlaylist, err := m.repository.CreateRemotePlaylist(v.DirName)
		if err != nil {
			return err
		}
		fmt.Println("+", v.DirName)

		// プレイリストファイルをローカルに新規に生成する
		resPlaylist.DirName = v.DirName
		m.repository.CreatePlaylistContent(resPlaylist)

		// ローカルの曲を取得
		tracks, err := m.repository.FetchLocalPlaylistTrack(v.DirName)
		if err != nil {
			return err
		}

		// 曲をリモートのプレイリストに追加
		res, err := m.repository.AddRemoteTrack(resPlaylist.Id, tracks)
		if err != nil {
			return err
		}
		// 楽曲txtを作り直す
		m.recreateTrackTxt(v, res)
	}

	for _, v := range diff.remoteOnly {
		// プレイリストをリモートから削除
		err := m.repository.RemoveRemotePlaylist(v)
		if err != nil {
			return err
		}
		fmt.Println("-", v.Name)
	}

	for _, v := range diff.both {
		// トラックの差分を検出
		diff, err := m.calcDiffTrack(v)
		if err != nil {
			return err
		}
		// 差分が0の場合は何もしない
		if len(diff.localOnly) == 0 && len(diff.remoteOnly) == 0 {
			continue
		}
		fmt.Println(" ", v.Name)

		// 曲をリモートのプレイリストに追加
		res, err := m.repository.AddRemoteTrack(v.Id, diff.localOnly)
		if err != nil {
			return err
		}
		// 楽曲txtを作り直す
		m.recreateTrackTxt(v, res)

		// 曲をリモートのプレイリストから削除
		err = m.repository.RemoveRemoteTrack(v, diff.remoteOnly)
		if err != nil {
			return err
		}
		for _, w := range diff.remoteOnly {
			fmt.Println("  -", w.Name)
		}
	}

	// 後片付け: 不要なプレイリストテキストを消去
	deleted, err := m.repository.CleanUpPlaylistContent()
	for _, d := range deleted {
		log.Println(d, "was deleted.")
	}
	if err != nil {
		return err
	}
	return nil
}

func (m *model) ComparePlaylists() error {
	diff, err := m.calcDiffPlaylist()
	if err != nil {
		return err
	}

	for _, v := range diff.localOnly {
		fmt.Println("+", v.DirName)
		tracks, err := m.repository.FetchLocalPlaylistTrack(v.DirName)
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  +", w.FileName)
		}
	}

	for _, v := range diff.remoteOnly {
		fmt.Println("-", v.Name)
	}

	for _, v := range diff.both {
		diff, err := m.calcDiffTrack(v)
		if err != nil {
			return err
		}
		// 差分が0の場合は何も出力しない
		if len(diff.localOnly) == 0 && len(diff.remoteOnly) == 0 {
			continue
		}

		fmt.Println(" ", v.Name)

		for _, w := range diff.localOnly {
			fmt.Println("  +", w.Name)
		}
		for _, w := range diff.remoteOnly {
			fmt.Println("  -", w.Name)
		}
	}

	return nil
}

type diffPlaylist struct {
	localOnly  []models.PlaylistContent
	remoteOnly []models.PlaylistContent
	both       []models.PlaylistContent
}

func (m *model) calcDiffPlaylist() (diffPlaylist, error) {
	// プレイリスト情報のテキストファイル読み込み
	playlistContents, err := m.repository.FetchLocalPlaylistContent()
	if err != nil {
		return diffPlaylist{}, err
	}
	// 検索しやすいようにmapにする
	dirNameToPL := map[string]models.PlaylistContent{}
	for _, v := range playlistContents {
		dirNameToPL[v.DirName] = v
	}

	// ディレクトリの一覧を取得
	dirs, err := m.repository.FetchLocalPlaylistDir()
	if err != nil {
		return diffPlaylist{}, err
	}
	// ディレクトリを "プレイリスト情報のテキストファイル" の情報と関連付けて配列で保存
	localPLs := []models.PlaylistContent{}
	for _, v := range dirs {
		if w, isExist := dirNameToPL[v]; isExist {
			// プレイリスト情報txtが存在する場合
			localPLs = append(localPLs, w)
		} else {
			localPLs = append(localPLs, models.PlaylistContent{DirName: v})
		}
	}

	// リモートのプレイリストを配列で取得
	remotePLs, err := m.repository.FetchRemotePlaylistContent()
	if err != nil {
		return diffPlaylist{}, err
	}

	// 検索しやすいようにmapを作成
	idToPlaylist := map[string]models.PlaylistContent{}
	for _, v := range localPLs {
		idToPlaylist[v.Id] = v
	}
	for _, v := range remotePLs {
		// 上書きしないように
		old := idToPlaylist[v.Id]
		idToPlaylist[v.Id] = models.PlaylistContent{Name: v.Name, DirName: old.DirName, Id: v.Id}
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

	remoteOnly := []models.PlaylistContent{}
	both := []models.PlaylistContent{}
	for id, bit := range playlistState {
		if bit == 1 {
			localOnly = append(localOnly, idToPlaylist[id])
		}
		if bit == 2 {
			remoteOnly = append(remoteOnly, idToPlaylist[id])
		}
		if bit == 3 {
			both = append(both, idToPlaylist[id])
		}
	}

	return diffPlaylist{localOnly: localOnly, remoteOnly: remoteOnly, both: both}, nil

}

type diffTrack struct {
	localOnly  []models.TrackContent
	remoteOnly []models.TrackContent
	both       []models.TrackContent
}

func (m *model) calcDiffTrack(playlist models.PlaylistContent) (diffTrack, error) {
	if playlist.DirName == "" {
		return diffTrack{}, fmt.Errorf("property DirName is empty")
	}
	if playlist.Id == "" {
		return diffTrack{}, fmt.Errorf("property Id is empty")
	}

	localTracks, err := m.repository.FetchLocalPlaylistTrack(playlist.DirName)
	if err != nil {
		return diffTrack{}, err
	}
	remoteTracks, err := m.repository.FetchRemotePlaylistTrack(playlist.Id)
	if err != nil {
		return diffTrack{}, err
	}

	idToTrack := map[string]models.TrackContent{}
	for _, v := range localTracks {
		idToTrack[v.Id] = v
	}
	for _, v := range remoteTracks {
		idToTrack[v.Id] = v
	}

	// localにあれば +1, remoteにあれば +2する
	// 1ならlocalのみ, 2ならremoteのみ, 3なら両方に存在することになる
	trackStatus := map[string]int{}
	localOnly := []models.TrackContent{}
	for _, v := range localTracks {
		if v.Id == "" {
			localOnly = append(localOnly, v)
		} else {
			trackStatus[v.Id] += 1
		}
	}
	for _, v := range remoteTracks {
		trackStatus[v.Id] += 2
	}

	remoteOnly := []models.TrackContent{}
	both := []models.TrackContent{}
	for id, bit := range trackStatus {
		if bit == 1 {
			localOnly = append(localOnly, idToTrack[id])
		}
		if bit == 2 {
			remoteOnly = append(remoteOnly, idToTrack[id])
		}
		if bit == 3 {
			both = append(both, idToTrack[id])
		}
	}

	return diffTrack{localOnly: localOnly, remoteOnly: remoteOnly, both: both}, nil
}

func (m *model) CreatePlaylistDirectory(playlist models.PlaylistContent) error {
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

func (m *model) PullPlaylists() error {
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
