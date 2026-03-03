package interfaces

import "github.com/kajikentaro/spotify-fbc/models"

type Repository interface {
	AddRemoteTrack(playlistId string, tracks []models.TrackContent, c chan []models.TrackContent) error
	CleanUpPlaylistContent() ([]string, error)
	CreatePlaylistContent(playlist models.PlaylistContent) error
	CreatePlaylistDirectory(playlist models.PlaylistContent) error
	CreateRemotePlaylist(name string) (models.PlaylistContent, error)
	CreateRootDir() error
	CreateTrackContent(dirName string, track models.TrackContent) error
	FetchLocalPlaylistContent() ([]models.PlaylistContent, error)
	FetchLocalPlaylistTrack(dirName string) ([]models.TrackContent, error)
	FetchRemotePlaylistContent() ([]models.PlaylistContent, error)
	FetchRemotePlaylistTrack(id string) ([]models.TrackContent, error)
	RemoveRemotePlaylist(playlist models.PlaylistContent) error
	RemoveRemoteTrack(playlist models.PlaylistContent, tracks []models.TrackContent) error
	RemoveTrackContent(dirName string, track models.TrackContent) error
}
