package app

import (
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"
	hook "github.com/robotn/gohook"
	"golang.org/x/image/draw"

	"github.com/kkkunny/chip-8/os"
)

var keyMap = map[uint16]uint8{
	'1': 0,
	'2': 1,
	'3': 2,
	'4': 3,
	81:  4,  // q
	87:  5,  // w
	69:  6,  // e
	82:  7,  // r
	65:  8,  // a
	83:  9,  // s
	68:  10, // d
	70:  11, // f
	90:  12, // z
	88:  13, // x
	67:  14, // c
	86:  15, // v
}

type App struct {
	app    fyne.App
	window fyne.Window
	screen *canvas.Image

	cpu        *os.CPU
	loadChan   chan string
	screenImg  image.Image
	onKeyEvent chan<- tuple.Tuple2[os.KeyEvent, uint8]
}

func NewApp() *App {
	app := &App{app: fyneApp.New()}
	app.window = app.app.NewWindow("Chip-8 CPU")
	keyChan := make(chan tuple.Tuple2[os.KeyEvent, uint8])
	app.cpu = os.NewCPU(keyChan)
	app.onKeyEvent = keyChan
	app.loadChan = make(chan string, 1)

	app.initWindow()
	app.initScreen()

	return app
}

func (app *App) initWindow() {
	app.window.Resize(fyne.NewSize(640, 320))
	app.window.SetFixedSize(true)
	app.window.CenterOnScreen()
	app.window.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("菜单", fyne.NewMenuItem("载入游戏", func() {
		selectFileWindow := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				panic(err)
			}
			app.loadChan <- reader.URI().Path()
		}, app.window)
		selectFileWindow.Show()
	}))))
}

func (app *App) initScreen() {
	app.screenImg = image.NewRGBA(image.Rect(0, 0, 640, 320))
	app.screen = canvas.NewImageFromImage(app.screenImg)
	app.screen.FillMode = canvas.ImageFillOriginal
	app.window.SetContent(app.screen)

	img := app.screenImg.(draw.Image)
	app.cpu.SetOnReset(func() {
		draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)
	})
	app.cpu.SetOnScreenUpdate(func(x uint16, y uint16, clr color.Color) {
		for i := range 10 {
			for j := range 10 {
				img.Set(int(x)*10+i, int(y)*10+j, clr)
			}
		}
	})
}

func (app *App) listenKeyboard() {
	go func() {
		for key := range hook.Start() {
			switch key.Rawcode {
			case 27:
				app.app.Quit()
				return
			default:
				chip8Key, ok := keyMap[key.Rawcode]
				if !ok {
					break
				}
				switch key.Kind {
				case hook.KeyDown:
					app.onKeyEvent <- tuple.Pack2(os.KeyEventDown, chip8Key)
				case hook.KeyUp:
					app.onKeyEvent <- tuple.Pack2(os.KeyEventUp, chip8Key)
				}
			}
		}
	}()
}

func (app *App) mainLoop() {
	var gaming bool
	for {
		select {
		case path := <-app.loadChan:
			app.cpu.Reset()
			stlerr.Must(app.cpu.Load(path))
			gaming = true
		default:
			if !gaming {
				break
			}
			time.Sleep(time.Second / 200)
			app.cpu.Next()
			app.screen.Refresh()
		}
	}
}

func (app *App) Run() {
	app.listenKeyboard()
	go app.mainLoop()
	app.window.ShowAndRun()
}
