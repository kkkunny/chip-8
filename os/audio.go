package os

import (
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/chip-8/config"
)

var beepAudio beep.StreamSeekCloser

func init() {
	beepFile := stlerr.MustWith(os.Open("beep.mp3"))
	streamer, format := stlerr.MustWith2(mp3.Decode(beepFile))
	stlerr.Must(speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)))
	beepAudio = streamer
}

type audio struct {
	lock sync.Locker
}

func newAudio() *audio {
	return &audio{
		lock: &sync.Mutex{},
	}
}

func (a *audio) Play() {
	go func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		config.Logger.Debug("play audio")

		done := make(chan struct{})
		speaker.Play(beep.Seq(beepAudio, beep.Callback(func() {
			done <- struct{}{}
		})))
		<-done
		stlerr.Must(beepAudio.Seek(0))
	}()
}
