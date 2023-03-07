package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
	"github.com/zmb3/spotify/v2"
)

func (r *Repository) FetchRemotePlaylistContent() ([]models.PlaylistContent, error) {
	result := []models.PlaylistContent{}
	playlists, err := r.client.CurrentUsersPlaylists(r.ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range playlists.Playlists {
		content := models.SimplePlaylistToContent(v)
		result = append(result, content)
	}

	return result, nil
}

func (r *Repository) FetchRemotePlaylistTrack(id string) ([]models.TrackContent, error) {
	LIMIT := 100
	result := []models.TrackContent{}
	for offset := 0; true; offset += LIMIT {
		playlistItemPage, err := r.client.GetPlaylistItems(r.ctx, spotify.ID(id), spotify.Limit(LIMIT), spotify.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist %s: %s", id, err)
		}
		for _, playlistItem := range playlistItemPage.Items {
			track := playlistItem.Track.Track
			trackContent := models.FullTrackToContent(track)
			result = append(result, trackContent)
		}
		if len(playlistItemPage.Items) != LIMIT {
			break
		}
	}
	return result, nil
}

func (r *Repository) CreateRemotePlaylist(name string) (models.PlaylistContent, error) {
	user, err := r.client.CurrentUser(r.ctx)
	if err != nil {
		return models.PlaylistContent{}, fmt.Errorf("failed to get a current user info: %w", err)
	}
	new, err := r.client.CreatePlaylistForUser(r.ctx, user.ID, name, "", false, false)
	if err != nil {
		return models.PlaylistContent{}, fmt.Errorf("failed to create playlist %s: %w", name, err)
	}
	return models.SimplePlaylistToContent(new.SimplePlaylist), nil
}

func (r *Repository) addRemoteTrack(playlistId string, tracks []models.TrackContent) ([]models.TrackContent, error) {
	if playlistId == "" {
		return nil, fmt.Errorf("playlistId is empty")
	}

	confirmedIds := []spotify.ID{}
	confirmedTracks := []models.TrackContent{}

	// IDが存在するものから先に処理
	inputIds := []spotify.ID{}
	inputTracks := []models.TrackContent{} // ログ用
	for _, v := range tracks {
		if v.Id == "" {
			continue
		}
		inputIds = append(inputIds, spotify.ID(v.Id))
		inputTracks = append(inputTracks, v)
	}
	if len(inputIds) > 0 {
		result, err := r.client.GetTracks(r.ctx, inputIds)
		if err != nil {
			// 失敗したときはIDが存在するすべてのトラックをエラーにする
			for _, w := range inputTracks {
				fmt.Println(w.FileName, "failed to search track: ", err.Error())
			}
		} else {
			for idx, w := range result {
				if w == nil {
					fmt.Println(inputTracks[idx].FileName, "no search result found", err.Error())
					continue
				}
				t := models.FullTrackToContent(w)
				t.FileName = inputTracks[idx].FileName
				confirmedIds = append(confirmedIds, spotify.ID(t.Id))
				confirmedTracks = append(confirmedTracks, t)
				fmt.Println(t.Name, "was found")
			}
		}
	}

	for _, v := range tracks {
		if v.Id != "" {
			continue
		}
		// IDがないときは検索する
		res, err := r.client.Search(r.ctx, v.SearchQuery(), spotify.SearchTypeTrack, spotify.Limit(1))
		// 30秒ごとのaccess limitがあるので1秒待機する
		time.Sleep(time.Second * 1)
		if err != nil {
			fmt.Println(v.FileName, "failed to search track: ", err.Error())
			continue
		}
		if len(res.Tracks.Tracks) == 0 {
			fmt.Println(v.FileName, "no search result found")
			continue
		}
		t := models.FullTrackToContent(&res.Tracks.Tracks[0])
		t.FileName = v.FileName
		confirmedIds = append(confirmedIds, spotify.ID(t.Id))
		confirmedTracks = append(confirmedTracks, t)
		fmt.Println(t.Name, "was found")
	}
	if len(confirmedIds) == 0 {
		// 検索結果が何も見つからなかった場合
		return nil, nil
	}

	_, err := r.client.AddTracksToPlaylist(r.ctx, spotify.ID(playlistId), confirmedIds...)
	if err != nil {
		return nil, err
	}
	return confirmedTracks, nil
}

func (r *Repository) AddRemoteTrack(playlistId string, tracks []models.TrackContent, c chan []models.TrackContent) error {
	// 50個ずつに分割して実行
	err := splitProcess(50, tracks, func(chunk []models.TrackContent) error {
		doneTracks, err := r.addRemoteTrack(playlistId, chunk)
		if err != nil {
			return err
		}
		// 実行の途中結果をすぐに返す
		c <- doneTracks
		return nil
	})
	return err
}

func splitProcess[T any](limit int, massiveArray []T, f func([]T) error) error {
	// execute with splited array
	for offset := 0; true; offset += limit {
		if len(massiveArray)-1 < offset {
			break
		}
		var chunk []T
		if offset+limit < len(massiveArray) {
			chunk = massiveArray[offset : offset+limit]
		} else {
			chunk = massiveArray[offset:]
		}
		err := f(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RemoveRemoteTrack(playlist models.PlaylistContent, tracks []models.TrackContent) error {
	if playlist.Id == "" {
		return errors.New("playlist id is empty")
	}

	trackIds := []spotify.ID{}
	for _, v := range tracks {
		if v.Id == "" {
			return fmt.Errorf("track %v is not have track id", v)
		}
		trackIds = append(trackIds, spotify.ID(v.Id))
	}

	err := splitProcess(50, trackIds, func(chunk []spotify.ID) error {
		_, err := r.client.RemoveTracksFromPlaylist(r.ctx, spotify.ID(playlist.Id), chunk...)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) RemoveRemotePlaylist(playlist models.PlaylistContent) error {
	if playlist.Id == "" {
		return errors.New("playlist id is empty")
	}
	err := r.client.UnfollowPlaylist(r.ctx, spotify.ID(playlist.Id))
	if err != nil {
		return err
	}
	return nil
}
