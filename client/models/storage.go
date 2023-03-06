package models

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/zmb3/spotify/v2"
)

type TrackContent struct {
	Id       string `title:"id" query:"-"`
	Name     string `title:"name" query:""`
	Artist   string `title:"artist" query:"artist"`
	Album    string `title:"album" query:"album"`
	Seconds  string `title:"seconds" query:"-"`
	Isrc     string `title:"isrc" query:"isrc"`
	FileName string `title:"file_name" query:"-"`
}

type PlaylistContent struct {
	Id      string `title:"id"`
	Name    string `title:"name"`
	DirName string `title:"dir_name"`
}

func UnmarshalTrackContent(text string) TrackContent {
	result := TrackContent{}

	entries := strings.Split(text, "\n")
	ts := reflect.TypeOf(result)
	vs := reflect.ValueOf(&result)
	for _, e := range entries {
		substring := strings.SplitN(e, " ", 2)
		if len(substring) < 2 {
			continue
		}
		key := substring[0]
		value := substring[1]

		for i := 0; i < ts.NumField(); i++ {
			t := ts.Field(i)
			if t.Tag.Get("title") != key {
				continue
			}
			vs.Elem().Field(i).SetString(value)
		}
	}
	return result
}
func UnmarshalPlaylistContent(text string) PlaylistContent {
	result := PlaylistContent{}

	entries := strings.Split(text, "\n")
	ts := reflect.TypeOf(result)
	vs := reflect.ValueOf(&result)
	for _, e := range entries {
		substring := strings.SplitN(e, " ", 2)
		if len(substring) < 2 {
			continue
		}
		key := substring[0]
		value := substring[1]

		for i := 0; i < ts.NumField(); i++ {
			t := ts.Field(i)
			if t.Tag.Get("title") != key {
				continue
			}
			vs.Elem().Field(i).SetString(value)
		}
	}
	return result
}

func (p PlaylistContent) Marshal() string {
	ts := reflect.TypeOf(p)
	vs := reflect.ValueOf(p)

	result := "NOTE: Do not delete or edit this file.\n\n"
	for i := 0; i < ts.NumField(); i++ {
		titleValue := ts.Field(i).Tag.Get("title")
		fieldValue := vs.Field(i).String()
		result += titleValue + " " + fieldValue + "\n"
	}
	return result
}

func (p TrackContent) Marshal() string {
	ts := reflect.TypeOf(p)
	vs := reflect.ValueOf(p)

	result := ""
	for i := 0; i < ts.NumField(); i++ {
		titleValue := ts.Field(i).Tag.Get("title")
		fieldValue := vs.Field(i).String()
		result += titleValue + " " + fieldValue + "\n"
	}
	return result
}

func (p TrackContent) SearchQuery() string {
	ts := reflect.TypeOf(p)
	vs := reflect.ValueOf(p)

	result := ""
	for i := 0; i < ts.NumField(); i++ {
		titleValue := ts.Field(i).Tag.Get("query")
		// "-"のときは無視
		if titleValue == "-" {
			continue
		}
		fieldValue := vs.Field(i).String()
		if fieldValue == "" {
			continue
		}
		// 曲名のときはタグを付けない
		if titleValue == "" {
			result += fieldValue + " "
			continue
		}

		result += titleValue + ":" + fieldValue + " "
	}
	return result
}

func FullTrackToContent(track *spotify.FullTrack) TrackContent {
	return TrackContent{
		Id:      track.ID.String(),
		Name:    track.Name,
		Artist:  joinArtistText(track.Artists),
		Album:   track.Album.Name,
		Seconds: strconv.Itoa(track.Duration),
		Isrc:    track.ExternalIDs["isrc"],
	}
}

func SimplePlaylistToContent(playlist spotify.SimplePlaylist) PlaylistContent {
	return PlaylistContent{Id: playlist.ID.String(), Name: playlist.Name}
}

func joinArtistText(artists []spotify.SimpleArtist) string {
	text := []string{}
	for _, a := range artists {
		text = append(text, a.Name)
	}
	return strings.Join(text, ", ")
}
