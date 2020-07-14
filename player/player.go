package player

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"os"
	"path/filepath"
	"time"
)

type Player struct {
	Ctrl       *beep.Ctrl
	SSC        beep.StreamSeekCloser
	BFormat    beep.Format
	Volume     *effects.Volume
	SampleRate beep.SampleRate
}

func (p *Player) Play(s *Song) (int, error) {
	speaker.Clear()
	f, err := os.Open(s.Path)
	if err != nil {
		return 0, err
	}

	switch filepath.Ext(s.Path) {
	case ".wav":
		p.SSC, p.BFormat, err = wav.Decode(f)
	case ".mp3":
		p.SSC, p.BFormat, err = mp3.Decode(f)
	case ".ogg", ".weba", ".webm":
		p.SSC, p.BFormat, err = vorbis.Decode(f)
	case ".flac":
		p.SSC, p.BFormat, err = flac.Decode(f)
	default:
		return 0, ErrUnsupportedFormat
	}
	if err != nil {
		return 0, err
	}
	if p.SampleRate == 0 {
		p.SampleRate = p.BFormat.SampleRate
		err = speaker.Init(p.SampleRate, p.SampleRate.N(time.Second/10))
		if err != nil {
			return 0, err
		}
	}
	// TODO: Resample is only a Streamer, not a StreamSeekCloser
	// resampled := beep.Resample(4, p.BFormat.SampleRate, p.SampleRate, p.SSC)
	p.SampleRate = p.BFormat.SampleRate
	p.Volume.Streamer = p.SSC
	p.Ctrl = &beep.Ctrl{Streamer: p.Volume}
	speaker.Play(p.Ctrl)
	return int(float32(p.SSC.Len()) / float32(p.BFormat.SampleRate)), nil
}

func (p *Player) Pause(state bool) {
	speaker.Lock()
	p.Ctrl.Paused = state
	speaker.Unlock()
}

func (p *Player) Seek(pos int) error {
	speaker.Lock()
	err := p.SSC.Seek(pos * int(p.BFormat.SampleRate))
	speaker.Unlock()
	return err
}

func (p *Player) SetVolume(percent int) {
	if percent == 0 {
		p.Volume.Silent = true
	} else {
		p.Volume.Silent = false
		p.Volume.Volume = -float64(100-percent) / 100.0 * 5
	}
}
