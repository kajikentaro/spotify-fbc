package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
	"github.com/zmb3/spotify/v2"
)

func FetchRemotePlaylistContent(client *spotify.Client, ctx context.Context) ([]models.PlaylistContent, error) {
	result := []models.PlaylistContent{}
	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range playlists.Playlists {
		result = append(result, models.PlaylistContent{Id: v.ID.String(), Name: v.Name})
	}

	return result, nil
}

func FetchRemotePlaylistTrack(client *spotify.Client, ctx context.Context, id string) ([]models.TrackContent, error) {
	LIMIT := 100
	result := []models.TrackContent{}
	for offset := 0; true; offset += 100 {
		playlistItemPage, err := client.GetPlaylistItems(ctx, spotify.ID(id), spotify.Limit(LIMIT), spotify.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist %s: %s", id, err)
		}
		for _, playlistItem := range playlistItemPage.Items {
			track := playlistItem.Track.Track
			trackContent := models.TrackContent{
				Id:      track.ID.String(),
				Name:    track.Name,
				Artist:  joinArtistText(track.Artists),
				Album:   track.Album.Name,
				Seconds: strconv.Itoa(track.Duration),
				Isrc:    track.ExternalIDs["isrc"],
			}
			result = append(result, trackContent)
		}
		if len(playlistItemPage.Items) != LIMIT {
			break
		}
	}
	return result, nil
}

func joinArtistText(artists []spotify.SimpleArtist) string {
	text := []string{}
	for _, a := range artists {
		text = append(text, a.Name)
	}
	return strings.Join(text, ", ")
}
