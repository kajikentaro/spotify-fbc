package service_compares

import (
	"fmt"

	"github.com/kajikentaro/spotify-fbc/models"
	"github.com/kajikentaro/spotify-fbc/repositories"
)

type compare struct {
	repository *repositories.Repository
}

func NewCompare(repository *repositories.Repository) compare {
	return compare{repository: repository}
}

type diffAll struct {
	LocalOnly []struct {
		Playlist models.PlaylistContent
		Tracks   []models.TrackContent
	}
	RemoteOnly []models.PlaylistContent
	Both       []struct {
		Playlist         models.PlaylistContent
		TracksRemoteOnly []models.TrackContent
		TracksLocalOnly  []models.TrackContent
	}
}

func (m *compare) CompareAll() (diffAll, error) {
	playlistDiff, err := m.calcDiffPlaylist()
	if err != nil {
		return diffAll{}, err
	}

	answer := diffAll{}

	for _, v := range playlistDiff.localOnly {
		tracks, err := m.repository.FetchLocalPlaylistTrack(v.DirName)
		if err != nil {
			return diffAll{}, err
		}

		r := struct {
			Playlist models.PlaylistContent
			Tracks   []models.TrackContent
		}{}
		r.Playlist = v
		r.Tracks = tracks
		answer.LocalOnly = append(answer.LocalOnly, r)
	}

	answer.RemoteOnly = playlistDiff.remoteOnly

	for _, v := range playlistDiff.both {
		diff, err := m.calcDiffTrack(v)
		if err != nil {
			return diffAll{}, err
		}

		if len(diff.localOnly)+len(diff.remoteOnly) == 0 {
			continue
		}

		r := struct {
			Playlist         models.PlaylistContent
			TracksRemoteOnly []models.TrackContent
			TracksLocalOnly  []models.TrackContent
		}{}
		r.Playlist = v
		r.TracksRemoteOnly = diff.remoteOnly
		r.TracksLocalOnly = diff.localOnly
		answer.Both = append(answer.Both, r)
	}

	return answer, nil
}

type diffPlaylist struct {
	localOnly  []models.PlaylistContent
	remoteOnly []models.PlaylistContent
	both       []models.PlaylistContent
}

func (m *compare) calcDiffPlaylist() (diffPlaylist, error) {
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

func (m *compare) calcDiffTrack(playlist models.PlaylistContent) (diffTrack, error) {
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
