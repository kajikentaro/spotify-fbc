package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
	"github.com/kajikentaro/spotify-file-based-client/client/repositories"
	"github.com/zmb3/spotify/v2"
)

type model struct {
	client   *spotify.Client
	ctx      context.Context
	rootPath string
}

func NewModel(client *spotify.Client, ctx context.Context, rootPath string) model {
	return model{client: client, ctx: ctx, rootPath: rootPath}
}

func (m *model) ComparePlaylists() error {
	diff, err := m.calcDiffPlaylist()
	if err != nil {
		return err
	}

	for _, v := range diff.localOnly {
		fmt.Println("+", v.DirName)
		tracks, err := repositories.FetchLocalPlaylistTrack(filepath.Join(m.rootPath, v.DirName))
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  +", w.FileName)
		}
	}

	for _, v := range diff.remoteOnly {
		fmt.Println("-", v.Name)
		tracks, err := repositories.FetchRemotePlaylistTrack(m.client, m.ctx, v.Id)
		if err != nil {
			return err
		}
		for _, w := range tracks {
			fmt.Println("  -", w.Name)
		}
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
	playlistContents, err := repositories.FetchLocalPlaylistContent(m.rootPath)
	if err != nil {
		return diffPlaylist{}, err
	}
	// 検索しやすいようにmapにする
	dirNameToPL := map[string]models.PlaylistContent{}
	for _, v := range playlistContents {
		dirNameToPL[v.DirName] = v
	}

	// ディレクトリの一覧を取得
	dirs, err := repositories.FetchLocalPlaylistDir(m.rootPath)
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
	remotePLs, err := repositories.FetchRemotePlaylistContent(m.client, m.ctx)
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

	localTracks, err := repositories.FetchLocalPlaylistTrack(filepath.Join(m.rootPath, playlist.DirName))
	if err != nil {
		return diffTrack{}, err
	}
	remoteTracks, err := repositories.FetchRemotePlaylistTrack(m.client, m.ctx, playlist.Id)
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
	err := repositories.CreatePlaylistContent(m.rootPath, playlist)
	if err != nil {
		return err
	}

	// generate a playlist directory
	err = repositories.CreatePlaylistDirectory(m.rootPath, playlist)
	if err != nil {
		return err
	}

	playlistTrack, err := repositories.FetchRemotePlaylistTrack(m.client, m.ctx, playlist.Id)
	if err != nil {
		return err
	}

	// generate a track file in the directory
	usedTrackNames := map[string]struct{}{}
	for _, track := range playlistTrack {
		fileStem := replaceBannedCharacter(track.Name)
		track.FileName = unique(&usedTrackNames, fileStem) + ".txt"
		repositories.CreateTrackContent(filepath.Join(m.rootPath, playlist.DirName), track)
	}
	return nil
}

func replaceBannedCharacter(path string) string {
	reg := regexp.MustCompile("[\\\\/:*?\"<>|]")
	return reg.ReplaceAllString(path, " ")
}

func (m *model) PullPlaylists() error {
	playlists, err := m.client.CurrentUsersPlaylists(m.ctx)
	if err != nil {
		return err
	}
	os.Mkdir(m.rootPath, os.ModePerm)

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
