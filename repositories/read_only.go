package repositories

import (
	"context"
	"fmt"

	"github.com/kajikentaro/spotify-fbc/models"
	"github.com/kajikentaro/spotify-fbc/services/interfaces"
	"github.com/zmb3/spotify/v2"
)

type ReadOnlyRepository struct {
	rootPath       string
	realRepository *Repository
	showLog        bool
}

func NewReadOnlyRepository(client *spotify.Client, ctx context.Context, rootPath string, showLog bool) interfaces.Repository {
	real := NewRepository(client, ctx, rootPath)
	return &ReadOnlyRepository{rootPath: rootPath, realRepository: real, showLog: showLog}
}

func (r *ReadOnlyRepository) AddRemoteTrack(playlistId string, tracks []models.TrackContent, c chan []models.TrackContent) error {
	c <- tracks
	if r.showLog {
		fmt.Printf("===DRY RUN=== AddRemoteTrack: playlistId=%s, tracks=%v\n", playlistId, tracks)
	}
	return nil
}

func (r *ReadOnlyRepository) CleanUpPlaylistContent() ([]string, error) {
	if r.showLog {
		fmt.Println("===DRY RUN=== CleanUpPlaylistContent")
	}
	return nil, nil
}

func (r *ReadOnlyRepository) CreatePlaylistContent(playlist models.PlaylistContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== CreatePlaylistContent: playlist=%v\n", playlist)
	}
	return nil
}

func (r *ReadOnlyRepository) CreatePlaylistDirectory(playlist models.PlaylistContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== CreatePlaylistDirectory: playlist=%v\n", playlist)
	}
	return nil
}

func (r *ReadOnlyRepository) CreateRemotePlaylist(name string) (models.PlaylistContent, error) {
	if r.showLog {
		fmt.Printf("===DRY RUN=== CreateRemotePlaylist: name=%s\n", name)
	}
	return models.PlaylistContent{Name: name, DirName: name}, nil
}

func (r *ReadOnlyRepository) CreateRootDir() error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== CreateRootDir: rootPath=%s\n", r.rootPath)
	}
	return nil
}

func (r *ReadOnlyRepository) CreateTrackContent(dirName string, track models.TrackContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== CreateTrackContent: dirName=%s, track=%v\n", dirName, track)
	}
	return nil
}

func (r *ReadOnlyRepository) FetchLocalPlaylistContent() ([]models.PlaylistContent, error) {
	return r.realRepository.FetchLocalPlaylistContent()
}

func (r *ReadOnlyRepository) FetchLocalPlaylistTrack(dirName string) ([]models.TrackContent, error) {
	return r.realRepository.FetchLocalPlaylistTrack(dirName)
}

func (r *ReadOnlyRepository) FetchRemotePlaylistContent() ([]models.PlaylistContent, error) {
	return r.realRepository.FetchRemotePlaylistContent()
}

func (r *ReadOnlyRepository) FetchRemotePlaylistTrack(id string) ([]models.TrackContent, error) {
	return r.realRepository.FetchRemotePlaylistTrack(id)
}

func (r *ReadOnlyRepository) RemoveRemotePlaylist(playlist models.PlaylistContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== RemoveRemotePlaylist: playlist=%v\n", playlist)
	}
	return nil
}

func (r *ReadOnlyRepository) RemoveRemoteTrack(playlist models.PlaylistContent, tracks []models.TrackContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== RemoveRemoteTrack: playlist=%v, tracks=%v\n", playlist, tracks)
	}
	return nil
}

func (r *ReadOnlyRepository) RemoveTrackContent(dirName string, track models.TrackContent) error {
	if r.showLog {
		fmt.Printf("===DRY RUN=== RemoveTrackContent: dirName=%s, track=%v\n", dirName, track)
	}
	return nil
}
