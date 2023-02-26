package models

import (
	"reflect"
	"strings"
)

type TruckContent struct {
	Id      string `title:"id"`
	Name    string `title:"name"`
	Artist  string `title:"artist"`
	Album   string `title:"album"`
	Seconds string `title:"seconds"`
	Isrc    string `title:"isrc"`
}

type PlaylistContent struct {
	Id      string `title:"id"`
	Name    string `title:"name"`
	DirName string `title:"dir_name"`
}

func unmarshalTruckContent(text string) TruckContent {
	result := TruckContent{}

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
func unmarshalPlaylistContent(text string) PlaylistContent {
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

func (p PlaylistContent) marshal() (string, error) {
	ts := reflect.TypeOf(p)
	vs := reflect.ValueOf(p)

	result := ""
	for i := 0; i < ts.NumField(); i++ {
		fieldValue := vs.Field(i).String()
		result += ts.Field(i).Tag.Get("title") + " " + fieldValue + "\n"
	}
	return result, nil
}

func (p TruckContent) marshal() (string, error) {
	ts := reflect.TypeOf(p)
	vs := reflect.ValueOf(p)

	result := ""
	for i := 0; i < ts.NumField(); i++ {
		fieldValue := vs.Field(i).String()
		result += ts.Field(i).Tag.Get("title") + " " + fieldValue + "\n"
	}
	return result, nil
}
