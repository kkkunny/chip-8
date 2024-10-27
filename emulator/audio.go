package emulator

import (
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	stlerr "github.com/kkkunny/stl/error"
)

var beepAudio beep.StreamSeekCloser

func init() {
	beepFile := stlerr.MustWith(os.Open("beep.mp3"))
	streamer, format := stlerr.MustWith2(mp3.Decode(beepFile))
	stlerr.Must(speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)))
	beepAudio = streamer
}

type Audio struct {
	lock sync.Locker
}

func newAudio() *Audio {
	return &Audio{
		lock: &sync.Mutex{},
	}
}

func (a *Audio) Play() {
	go func() {
		a.lock.Lock()
		defer a.lock.Unlock()

		done := make(chan struct{})
		speaker.Play(beep.Seq(beepAudio, beep.Callback(func() {
			done <- struct{}{}
		})))
		<-done
		stlerr.Must(beepAudio.Seek(0))
	}()
}
