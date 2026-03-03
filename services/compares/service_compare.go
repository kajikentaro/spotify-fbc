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

type PlaylistTrackDiff struct {
	Playlist WithDiffState[models.PlaylistContent]
	Tracks   []WithDiffState[models.TrackContent]
}

func (m *compare) CompareAll() ([]PlaylistTrackDiff, error) {
	playlistDiff, err := m.calcDiffPlaylist()
	if err != nil {
		return nil, err
	}

	res := []PlaylistTrackDiff{}
	for _, v := range playlistDiff {
		if v.DiffState == LocalOnly {
			tracks, err := m.repository.FetchLocalPlaylistTrack(v.V.DirName)
			if err != nil {
				return nil, err
			}
			trackWithDiffStates := make([]WithDiffState[models.TrackContent], len(tracks))
			for i, track := range tracks {
				trackWithDiffStates[i] = WithDiffState[models.TrackContent]{V: track, DiffState: LocalOnly}
			}
			r := PlaylistTrackDiff{Playlist: v, Tracks: trackWithDiffStates}
			res = append(res, r)
		}

		if v.DiffState == RemoteOnly {
			// パフォーマンスのため、Remoteにのみ存在するTrackはfetchしない (必要があれば後で修正する)
			res = append(res, PlaylistTrackDiff{Playlist: v})
		}

		if v.DiffState == Both {
			diff, err := m.calcDiffTrack(v.V)
			if err != nil {
				return nil, err
			}
			r := PlaylistTrackDiff{Playlist: v, Tracks: diff}
			res = append(res, r)
		}
	}

	return res, nil
}

func (m *compare) calcDiffPlaylist() ([]WithDiffState[models.PlaylistContent], error) {
	localPLs, err := m.repository.FetchLocalPlaylistContent()
	if err != nil {
		return nil, err
	}

	remotePLs, err := m.repository.FetchRemotePlaylistContent()
	if err != nil {
		return nil, err
	}

	getId := func(p models.PlaylistContent) string {
		return p.Id
	}
	merge := func(local models.PlaylistContent, remote models.PlaylistContent) models.PlaylistContent {
		return models.PlaylistContent{Name: remote.Name, DirName: local.DirName, Id: remote.Id}
	}
	diff := calcDiff(localPLs, remotePLs, getId, merge)

	return diff, nil
}

func (m *compare) calcDiffTrack(playlist models.PlaylistContent) ([]WithDiffState[models.TrackContent], error) {
	if playlist.DirName == "" {
		return nil, fmt.Errorf("property DirName is empty")
	}
	if playlist.Id == "" {
		return nil, fmt.Errorf("property Id is empty")
	}

	localTracks, err := m.repository.FetchLocalPlaylistTrack(playlist.DirName)
	if err != nil {
		return nil, err
	}
	remoteTracks, err := m.repository.FetchRemotePlaylistTrack(playlist.Id)
	if err != nil {
		return nil, err
	}
	getId := func(track models.TrackContent) string {
		return track.Id
	}
	merge := func(local models.TrackContent, remote models.TrackContent) models.TrackContent {
		return models.TrackContent{Id: remote.Id, Name: remote.Name, FileName: local.FileName}
	}

	res := calcDiff(localTracks, remoteTracks, getId, merge)
	return res, nil
}
