package player

import (
	"fmt"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"math/rand"
	"time"
)

type selectCallback func(*Song) (int, error)
type pauseCallback func(bool)
type seekCallback func(int) error
type volumeCallback func(int)

type Ui struct {
	infoList      *widgets.List
	playList      *widgets.List
	scrollerGauge *widgets.Gauge
	volumeGauge   *widgets.Gauge
	controlsPar   *widgets.Paragraph

	songs     []*Song
	songNames []string

	volume  int
	songNum int
	songPos int
	songLen int

	// next song
	nextFlag uint8 // 0: sequential, 1: single, 2: random

	OnSelect selectCallback
	OnPause  pauseCallback
	OnSeek   seekCallback
	OnVolume volumeCallback

	state uiState
}

func NewUi(songList []*Song, pathPrefix int) error {
	err := termui.Init()
	if err != nil {
		return err
	}
	defer termui.Close()

	rand.Seed(time.Now().Unix())

	ui := new(Ui)
	ui.OnSelect = playerSgt.Play
	ui.OnPause = playerSgt.Pause
	ui.OnSeek = playerSgt.Seek
	ui.OnVolume = playerSgt.SetVolume

	ui.volume = 100

	ui.songs = songList
	ui.songNum = -1
	ui.infoList = widgets.NewList()
	ui.infoList.Title = "Song info"
	ui.infoList.TitleStyle = termui.NewStyle(termui.ColorGreen)
	ui.infoList.SelectedRowStyle.Fg = termui.ColorBlack
	ui.infoList.SelectedRowStyle.Bg = termui.ColorGreen
	ui.infoList.Border = true
	ui.infoList.BorderStyle = termui.NewStyle(termui.ColorGreen)

	ui.playList = widgets.NewList()
	ui.playList.Title = "Playlist"
	ui.playList.TitleStyle = termui.NewStyle(termui.ColorGreen)
	ui.playList.SelectedRowStyle.Fg = termui.ColorBlack
	ui.playList.SelectedRowStyle.Bg = termui.ColorGreen
	ui.playList.Border = true
	ui.playList.BorderStyle = termui.NewStyle(termui.ColorGreen)

	ui.scrollerGauge = widgets.NewGauge()
	ui.scrollerGauge.Title = "Stopped"
	ui.scrollerGauge.TitleStyle.Fg = termui.ColorBlack
	ui.scrollerGauge.TitleStyle.Bg = termui.ColorRed
	ui.scrollerGauge.BarColor = termui.ColorWhite

	ui.volumeGauge = widgets.NewGauge()
	ui.volumeGauge.Title = "Volume"
	ui.volumeGauge.TitleStyle.Fg = termui.ColorGreen
	ui.volumeGauge.BarColor = termui.ColorWhite
	ui.volumeGauge.Percent = ui.volume

	ui.controlsPar = widgets.NewParagraph()
	ui.controlsPar.Text = "[ Enter ](fg:black,bg:white)[Select](fg:black,bg:green) " +
		"[ p ](fg:black,bg:white)[Play/Pause](fg:black,bg:green) " +
		"[Esc](fg:black,bg:white)[Stop](fg:black,bg:green) " +
		"[ r ](fg:black,bg:white)[Random](fg:black,bg:green) " +
		"[ s ](fg:black,bg:white)[Single](fg:black,bg:green) " +
		"[Right](fg:black,bg:white)[+10s](fg:black,bg:green) " +
		"[Left](fg:black,bg:white)[-10s](fg:black,bg:green) " +
		"[ + ](fg:black,bg:white)[+Volume](fg:black,bg:green) " +
		"[ - ](fg:black,bg:white)[-Volume](fg:black,bg:green) " +
		"[ q ](fg:black,bg:white)[Exit](fg:black,bg:green) "
	ui.controlsPar.Border = false

	uiEvents := termui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	ui.songNames = make([]string, len(ui.songs))
	for i, v := range ui.songs {
		if v.Meta != nil {
			ui.songNames[i] = fmt.Sprintf("[%d] %s - %s", i+1, v.Meta.Artist(), v.Meta.Title())
		} else {
			ui.songNames[i] = fmt.Sprintf("[%d] %s", i+1, v.Path[pathPrefix:])
		}
	}
	ui.playList.Rows = ui.songNames
	ui.render()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "r":
				if ui.nextFlag == 2 {
					ui.nextFlag = 0
				} else {
					ui.nextFlag = 2
				}
				ui.updateSongInfo()
			case "s":
				if ui.nextFlag == 1 {
					ui.nextFlag = 0
				} else {
					ui.nextFlag = 1
				}
				ui.updateSongInfo()
			case "<Up>":
				ui.playList.ScrollUp()
			case "<Down>":
				ui.playList.ScrollDown()
			case "<Enter>":
				ui.songNum = ui.playList.SelectedRow
				ui.playSong(ui.songNum)
			case "p":
				if ui.songNum != -1 {
					if ui.state == Playing {
						ui.OnPause(true)
						ui.state = Paused
					} else {
						ui.OnPause(false)
						ui.state = Playing
					}
					ui.updateStatus()
				}
			case "<Escape>":
				ui.OnPause(true)
				ui.state = Stopped
				ui.scrollerGauge.Title = "Stopped"
				ui.scrollerGauge.TitleStyle.Bg = termui.ColorRed
				ui.scrollerGauge.Percent = 0
				ui.scrollerGauge.Label = "0:00 / 0:00"
			case "<Left>":
				if ui.songNum != -1 {
					ui.songPos -= 10
					if ui.songPos < 0 {
						ui.songPos = 0
					}
					if err := ui.OnSeek(ui.songPos); err != nil {
						ui.infoList.Rows = append(ui.infoList.Rows,
							fmt.Sprintf("[Error:](fg:red)   %s", err.Error()))
					}
				}
			case "<Right>":
				if ui.songNum != -1 {
					ui.songPos += 10
					if err := ui.OnSeek(ui.songPos); err != nil {
						ui.infoList.Rows = append(ui.infoList.Rows,
							fmt.Sprintf("[Error:](fg:red)   %s", err.Error()))
					}
				}
			case "=", "+":
				ui.volumeUp()
			case "-", "_":
				ui.volumeDown()
			case "<Resize>":
				// render
			}
		case <-ticker:
			if ui.state == Playing {
				ui.songPos++
				ui.updateGauge()
			} else if ui.state == Stopped {
				ui.songPos = 0
			}
		}
		ui.render()
	}
}

func (ui *Ui) playSong(number int) {
	ui.songPos = 0
	var err error
	ui.songLen, err = ui.OnSelect(ui.songs[number])
	if err == nil {
		ui.state = Playing
		ui.updateSongInfo()
		ui.updateStatus()
	} else {
		ui.infoList.Rows = []string{err.Error()}
	}
}

// Rendering
func (ui *Ui) render() {
	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		termui.NewRow(0.9,
			termui.NewCol(0.5,
				termui.NewRow(0.8, ui.infoList),
				termui.NewRow(0.1, ui.scrollerGauge),
				termui.NewRow(0.1, ui.volumeGauge)),
			termui.NewCol(1.0/2, ui.playList)),
		termui.NewRow(0.1, ui.controlsPar))
	termui.Clear()
	termui.Render(grid)
}

func (ui *Ui) updateSongInfo() {
	if ui.songNum != -1 {
		var nextStr string
		switch ui.nextFlag {
		case 0:
			nextStr = "Sequential"
		case 1:
			nextStr = "Single"
		case 2:
			nextStr = "Random"
		}
		if ui.songs[ui.songNum].Meta != nil {
			lyrics := ui.songs[ui.songNum].Meta.Lyrics()
			trackNum, _ := ui.songs[ui.songNum].Meta.Track()
			ui.infoList.Rows = []string{
				"[Artist:](fg:green) " + ui.songs[ui.songNum].Meta.Artist(),
				"[Title:](fg:green)  " + ui.songs[ui.songNum].Meta.Title(),
				"[Album:](fg:green)  " + ui.songs[ui.songNum].Meta.Album(),
				fmt.Sprintf("[Track:](fg:green)  %d", trackNum),
				"[Genre:](fg:green)  " + ui.songs[ui.songNum].Meta.Genre(),
				fmt.Sprintf("[Year:](fg:green)   %d", ui.songs[ui.songNum].Meta.Year()),
				fmt.Sprintf("[SampleRate:](fg:green)   %d", playerSgt.SampleRate),
				fmt.Sprintf("[Next Order:](fg:green)   %s", nextStr),
			}
			if lyrics != "" {
				ui.infoList.Rows = append(ui.infoList.Rows, "Lyrics:  "+lyrics)
			}
		} else {
			ui.infoList.Rows = []string{"No song info",
				fmt.Sprintf("[SampleRate:](fg:green)   %d", playerSgt.SampleRate),
				fmt.Sprintf("[Next Order:](fg:green)   %s", nextStr),
			}
		}
	}
}

func (ui *Ui) updateStatus() {
	switch ui.state {
	case Playing:
		ui.scrollerGauge.Title = "Playing"
		ui.scrollerGauge.TitleStyle.Bg = termui.ColorGreen
	case Paused:
		ui.scrollerGauge.Title = "Paused"
		ui.scrollerGauge.TitleStyle.Bg = termui.ColorYellow
	case Stopped:
		ui.scrollerGauge.Title = "Stopped"
		ui.scrollerGauge.TitleStyle.Bg = termui.ColorRed
	}
}

func (ui *Ui) volumeUp() {
	if ui.volume < 100 {
		ui.volume += 5
	}
	ui.volumeGauge.Percent = ui.volume
	ui.OnVolume(ui.volume)
}

func (ui *Ui) volumeDown() {
	if ui.volume > 0 {
		ui.volume -= 5
	}
	ui.volumeGauge.Percent = ui.volume
	ui.OnVolume(ui.volume)
}

func (ui *Ui) updateGauge() {
	if ui.songLen != 0 {
		ui.scrollerGauge.Percent = int(float32(ui.songPos) / float32(ui.songLen) * 100)
		ui.scrollerGauge.Label = fmt.Sprintf("%d:%.2d / %d:%.2d", ui.songPos/60, ui.songPos%60, ui.songLen/60, ui.songLen%60)
		if ui.scrollerGauge.Percent >= 100 {
			lastNum := ui.songNum
			ui.songNum = ui.nextSongNum(lastNum)
			ui.playList.ScrollAmount(ui.songNum - lastNum)
			ui.playSong(ui.songNum)
		}
	}
}

func (ui *Ui) nextSongNum(current int) int {
	switch ui.nextFlag {
	// sequential
	case 0:
		current++
		if current >= len(ui.songs) {
			current = 0
		}
		return current
	// single
	case 1:
		return current
	// random
	case 2:
		return rand.Intn(len(ui.songNames))
	default:
		return 0
	}
}
