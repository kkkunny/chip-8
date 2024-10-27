package main

import (
	"image"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	stlerr "github.com/kkkunny/stl/error"

	hook "github.com/robotn/gohook"

	"github.com/kkkunny/chip-8/emulator"
)

type EmulatorApp struct {
	app      fyne.App
	window   fyne.Window
	screen   *canvas.Image
	loadChan chan string
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
	eapp := &EmulatorApp{
		app:      app,
		window:   window,
		screen:   sc,
		loadChan: make(chan string, 1),
		emulator: e,
	}

	window.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("菜单", fyne.NewMenuItem("载入游戏", func() {
		selectFileWindow := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				panic(err)
			}
			eapp.loadChan <- reader.URI().Path()
		}, window)
		selectFileWindow.Show()
	}))))
	return eapp
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
		var gaming bool
		for {
			select {
			case path := <-app.loadChan:
				app.emulator.Reset()
				stlerr.Must(app.emulator.LoadGame(path))
				gaming = true
			default:
				if !gaming {
					break
				}
				app.emulator.Run()
				app.emulator.Draw()
				app.screen.Refresh()
			}
		}
	}()

	app.window.ShowAndRun()
}

func main() {
	app := NewEmulatorApp()
	app.Run()
}
