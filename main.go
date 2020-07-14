package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"playmusic/player"
)

var (
	help    bool
	version bool
	input   string
)

func init() {
	flag.BoolVar(&help, "h", false, "help info")
	flag.BoolVar(&version, "v", false, "version info")
	flag.StringVar(&input, "i", "", "input music folder/file path")
	flag.Usage = usage
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `playmusic tool in golang to play music from cmd
Version: 0.0.1
Usage: playmusic [-hvi] [-h help] [-v version] [-i input music folder/file path]
Options
`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
	} else if version {
		fmt.Println("version: 0.0.1")
	} else if input == "" {
		fmt.Println("too less arguments, use '-h' to see help info")
		os.Exit(1)
	} else {
		songDir, err := filepath.Abs(input)
		if err != nil {
			log.Fatal("Can't open your music directory/file")
		}
		songs, err := player.GetSongList(songDir)
		if err != nil {
			log.Fatal("Can't get song list")
		}

		if len(songs) == 0 {
			log.Fatal("Could find any songs to play")
		}
		err = player.NewUi(songs, len(songDir))
		if err != nil {
			log.Fatal(err)
		}
	}
}
