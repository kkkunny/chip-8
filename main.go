package main

import (
	"image"
	"time"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/kkkunny/chip-8/emulator"
)

func main() {
	app := fyneApp.New()
	window := app.NewWindow("Chip-8 Emulator")
	window.Resize(fyne.NewSize(640, 320))
	window.SetFixedSize(true)

	img := image.NewRGBA(image.Rect(0, 0, 640, 320))
	chip8 := emulator.NewEmulator(img)
	err := chip8.LoadGame("roms/ibm_logo.ch8")
	if err != nil {
		panic(err)
	}

	sc := canvas.NewImageFromImage(img)
	sc.FillMode = canvas.ImageFillOriginal
	window.SetContent(sc)

	frameDuration := time.Second / 60
	go func() {
		for {
			start := time.Now()

			chip8.Run()
			chip8.Draw()
			sc.Refresh()

			elapsed := time.Since(start)
			sleepDuration := frameDuration - elapsed
			if sleepDuration > 0 {
				time.Sleep(sleepDuration)
			}
		}
	}()

	window.ShowAndRun()
}
