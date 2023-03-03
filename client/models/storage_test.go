package models

import "testing"

func Test_TrackContent_Marshal(t *testing.T) {
	tc := TrackContent{
		Id:       "123",
		Name:     "test track",
		Artist:   "test artist",
		Album:    "test album",
		Seconds:  "123",
		Isrc:     "ABCDEFG",
		FileName: "test track.txt",
	}
	actual := tc.Marshal()
	expected :=
		`id 123
name test track
artist test artist
album test album
seconds 123
isrc ABCDEFG
file_name test track.txt
`
	if expected != actual {
		t.Errorf("\nactual:\n%s \nexpected:\n%s", actual, expected)
	}
}

func Test_PlaylistContent_Marshal(t *testing.T) {
	tc := PlaylistContent{
		Id:      "123",
		Name:    "test playlist name",
		DirName: "test playlist name",
	}
	actual := tc.Marshal()
	expected :=
		`NOTE: Do not delete or edit this file.

id 123
name test playlist name
dir_name test playlist name
`
	if expected != actual {
		t.Errorf("\nactual:\n%s \nexpected:\n%s", actual, expected)
	}
}

func Test_UnmarshalPlaylistContent(t *testing.T) {
	text :=
		`id 123
name test playlist name
dir_name test playlist name
`

	actual := UnmarshalPlaylistContent(text)
	expected := PlaylistContent{
		Id:      "123",
		Name:    "test playlist name",
		DirName: "test playlist name",
	}
	if expected != actual {
		t.Errorf("\nactual:\n%s \nexpected:\n%s", actual, expected)
	}
}

func Test_UnmarshalTrackContent(t *testing.T) {
	text :=
		`id 123
name test track
artist test artist
album test album
seconds 123
isrc ABCDEFG
file_name test track.txt
`

	actual := UnmarshalTrackContent(text)
	expected := TrackContent{
		Id:       "123",
		Name:     "test track",
		Artist:   "test artist",
		Album:    "test album",
		Seconds:  "123",
		Isrc:     "ABCDEFG",
		FileName: "test track.txt",
	}
	if expected != actual {
		t.Errorf("\nactual:\n%s \nexpected:\n%s", actual, expected)
	}
}
