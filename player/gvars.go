package player

import (
	"errors"
	"github.com/faiface/beep/effects"
)

var SupportedFormat = []string{".wav", ".mp3", ".ogg", ".weba", ".webm", ".flac"}
var ErrUnsupportedFormat = errors.New("song format not supported")
var playerSgt *Player

type uiState int

const (
	Stopped uiState = iota
	Playing
	Paused
)

func init() {
	if playerSgt == nil {
		playerSgt = &Player{
			Volume: &effects.Volume{Base: 2},
		}
	}
}
