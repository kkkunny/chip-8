package main

import (
	"image"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	stlerr "github.com/kkkunny/stl/error"

	hook "github.com/robotn/gohook"

	"github.com/kkkunny/chip-8/emulator"
)

type EmulatorApp struct {
	app      fyne.App
	window   fyne.Window
	screen   *canvas.Image
	emulator *emulator.Emulator
}

func NewEmulatorApp() *EmulatorApp {
	app := fyneApp.New()
	window := app.NewWindow("Chip-8 Emulator")
	window.Resize(fyne.NewSize(640, 320))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	img := image.NewRGBA(image.Rect(0, 0, 640, 320))
	e := emulator.NewEmulator(img)

	sc := canvas.NewImageFromImage(img)
	sc.FillMode = canvas.ImageFillOriginal
	window.SetContent(sc)
	return &EmulatorApp{
		app:      app,
		window:   window,
		screen:   sc,
		emulator: e,
	}
}

func (app *EmulatorApp) LoadGame(path string) error {
	return app.emulator.LoadGame(path)
}

func (app *EmulatorApp) Run() {
	go func() {
		for key := range hook.Start() {
			switch key.Rawcode {
			case 27:
				app.app.Quit()
				return
			default:
				switch key.Kind {
				case hook.KeyDown:
					app.emulator.KeyDown(key.Rawcode)
				case hook.KeyUp:
					app.emulator.KeyUp(key.Rawcode)
				}
			}
		}
	}()

	go func() {
		for {
			app.emulator.Run()
			app.emulator.Draw()
			app.screen.Refresh()
		}
	}()

	app.window.ShowAndRun()
}

func main() {
	app := NewEmulatorApp()
	stlerr.Must(app.LoadGame("roms/BRIX.ch8"))
	app.Run()
}
