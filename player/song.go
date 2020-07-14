package player

import (
	"github.com/dhowden/tag"
	"os"
	"path/filepath"
)

type Song struct {
	Path string
	Meta tag.Metadata
}

func GetSongList(path string) ([]*Song, error) {
	sl := make([]*Song, 0)
	appendSong := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, sf := range SupportedFormat {
				if filepath.Ext(path) == sf {
					// skip error file
					if f, err := os.Open(path); err == nil {
						metadata, _ := tag.ReadFrom(f)
						sl = append(sl, &Song{
							Path: path,
							Meta: metadata,
						})
						_ = f.Close()
					}
				}
			}
		}
		return nil
	}
	err := filepath.Walk(path, appendSong)
	return sl, err
}
