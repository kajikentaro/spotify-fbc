package repositories

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kajikentaro/spotify-file-based-client/client/models"
)

// ローカルのプレイリスト情報txtファイルを読み込み
func FetchLocalPlaylistContent(rootPath string) ([]models.PlaylistContent, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	result := []models.PlaylistContent{}
	for _, v := range entries {
		reText := regexp.MustCompile(".txt$")
		if !reText.MatchString(v.Name()) || v.IsDir() {
			// .txtで終わらないファイル, ディレクトリの場合
			continue
		}

		// .txtで終わる名前のファイルの場合
		b, err := os.ReadFile(filepath.Join(rootPath, v.Name()))
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", v.Name(), err)
		}
		p := models.UnmarshalPlaylistContent(string(b))
		result = append(result, p)
	}
	return result, nil
}

// ローカルのディレクトリ一覧を読み込み
func FetchLocalPlaylistDir(rootPath string) ([]string, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		result = append(result, e.Name())
	}
	return result, nil
}

func FetchLocalPlaylistTrack(dirPath string) ([]models.TrackContent, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", dirPath, err)
	}

	// トラック用txtファイルを読み込み
	result := []models.TrackContent{}
	for _, e := range entries {
		reText := regexp.MustCompile(".txt$")
		if !reText.MatchString(e.Name()) || e.IsDir() {
			// .txtで終わらないファイル, ディレクトリの場合
			continue
		}

		content, err := os.ReadFile(filepath.Join(dirPath, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %w", filepath.Join(dirPath, e.Name()), err)
		}
		t := models.UnmarshalTrackContent(string(content))
		if t.FileName == "" {
			// ユーザーが新規作成したTrackのtxtにはおそらくfile_nameプロパティが無い
			t.FileName = e.Name()
		}
		if t.FileName != e.Name() {
			log.Printf("Warning: a file_name property was incorrect. The property in the file was '%s', but path was '%s'.", t.FileName, e.Name())
			t.FileName = e.Name()
		}
		result = append(result, t)
	}
	return result, nil
}

// ローカルのプレイリスト情報txtファイルを生成
func CreatePlaylistContent(rootPath string, playlist models.PlaylistContent) error {
	textContent := playlist.Marshal()
	filePath := filepath.Join(rootPath, playlist.DirName+".txt")
	err := os.WriteFile(filePath, []byte(textContent), 0666)
	if err != nil {
		return fmt.Errorf("failed to create %s", filePath)
	}
	return nil
}

// ローカルのプレイリスト用ディレクトリを作成
func CreatePlaylistDirectory(rootPath string, playlist models.PlaylistContent) error {
	dirPath := filepath.Join(rootPath, playlist.DirName)
	err := os.Mkdir(dirPath, os.ModePerm)
	if os.IsExist(err) {
		log.Println(playlist.Name, "is already created")
	} else if err != nil {
		return fmt.Errorf("failed to create %s", dirPath)
	}
	return nil
}

func CreateTrackContent(dirPath string, track models.TrackContent) error {
	textContent := track.Marshal()
	filePath := filepath.Join(dirPath, track.FileName)
	err := os.WriteFile(filePath, []byte(textContent), 0666)
	if err != nil {
		return fmt.Errorf("failed to create %s", filePath)
	}
	return nil
}
