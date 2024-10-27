package main

import (
	"image"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	hook "github.com/robotn/gohook"

	"github.com/kkkunny/chip-8/emulator"
)

func main() {
	app := fyneApp.New()
	window := app.NewWindow("Chip-8 Emulator")
	window.Resize(fyne.NewSize(640, 320))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	img := image.NewRGBA(image.Rect(0, 0, 640, 320))
	chip8 := emulator.NewEmulator(img)
	err := chip8.LoadGame("roms/BRIX")
	if err != nil {
		panic(err)
	}

	sc := canvas.NewImageFromImage(img)
	sc.FillMode = canvas.ImageFillOriginal
	window.SetContent(sc)

	go func() {
		for key := range hook.Start() {
			switch key.Rawcode {
			case 27:
				app.Quit()
				return
			default:
				switch key.Kind {
				case hook.KeyDown:
					chip8.KeyDown(key.Rawcode)
				case hook.KeyUp:
					chip8.KeyUp(key.Rawcode)
				}
			}
		}
	}()

	go func() {
		for {
			chip8.Run()
			chip8.Draw()
			sc.Refresh()
		}
	}()

	window.ShowAndRun()
}
