package emulator

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/kkkunny/stl/container/stack"
	stlval "github.com/kkkunny/stl/value"
	"golang.org/x/image/draw"
)

// 字符集
var fontset = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Emulator struct {
	memory     [4096]uint8 // 内存 4K
	v          [16]uint8   // 16个寄存器
	i          uint16      // 索引寄存器
	pc         uint16      // 程序计数器
	delayTimer uint8
	soundTimer uint8
	stack      stack.Stack[uint16] // 栈
	gfx        [64][32]bool        // 屏幕
	img        image.Image
}

func NewEmulator(img image.Image) *Emulator {
	emulator := &Emulator{img: img}
	emulator.Reset()
	return emulator
}

func (e *Emulator) Reset() {
	e.memory = [4096]uint8{}
	copy(e.memory[:], fontset[:])
	e.v = [16]uint8{}
	e.i = 0
	e.pc = 0x200
	e.delayTimer = 0
	e.soundTimer = 0
	e.stack = stack.New[uint16]()
	e.gfx = [64][32]bool{}
	draw.Draw(e.img.(draw.Image), e.img.Bounds(), &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)
}

func (e *Emulator) getVX(opcode Opcode) uint8    { return e.v[opcode.X()] }
func (e *Emulator) getVY(opcode Opcode) uint8    { return e.v[opcode.Y()] }
func (e *Emulator) setVX(opcode Opcode, x uint8) { e.v[opcode.X()] = x }

func (e *Emulator) LoadGame(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Read(e.memory[0x200 : 0x200+info.Size()])
	return err
}

func (e *Emulator) executeOpcode(opcode Opcode) {
	switch opcode.Op() {
	case 0x0:
		switch opcode.N() {
		case 0x0:
			e.gfx = [64][32]bool{}
			e.pc += 2
		case 0xE:
			e.pc = e.stack.Pop()
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0x1:
		e.pc = opcode.NNN()
	case 0x2:
		e.stack.Push(e.pc)
		e.pc = opcode.NNN()
	case 0x6:
		e.setVX(opcode, opcode.NN())
		e.pc += 2
	case 0x7:
		e.setVX(opcode, e.getVX(opcode)+opcode.NN())
		e.pc += 2
	case 0x8:
		switch opcode.N() {
		case 0x4:
			if e.getVX(opcode) < e.getVY(opcode) {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.setVX(opcode, e.getVX(opcode)+e.getVY(opcode))
			e.pc += 2
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0xA:
		e.i = opcode.NNN()
		e.pc += 2
	case 0xD:
		startX, startY := uint16(e.getVX(opcode))%64, uint16(e.getVY(opcode))%32
		z := uint16(opcode.N())

		e.v[0xF] = 0

		for row := range z {
			y := startY + row
			sprite := uint16(e.memory[e.i+row])
			for col := range uint16(8) {
				x := startX + col
				oldPixel := &e.gfx[x][y]
				newPixel := (sprite & (1 << (7 - col))) != 0
				*oldPixel = *oldPixel != newPixel
			}
		}

		e.pc += 2
	case 0xF:
		switch opcode.NN() {
		case 0x33:
			e.memory[e.i] = e.getVX(opcode) / 100
			e.memory[e.i+1] = (e.getVX(opcode) % 100) / 10
			e.memory[e.i+2] = e.getVX(opcode) % 10
			e.pc += 2
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	default:
		panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
	}
}

func (e *Emulator) Run() {
	opcode := NewOpcode(e.memory[e.pc], e.memory[e.pc+1])
	fmt.Printf("Opcode: 0x%04X\n", opcode)
	e.executeOpcode(opcode)

	if e.delayTimer > 0 {
		e.delayTimer -= 1
	}
	if e.soundTimer > 0 {
		if e.soundTimer == 1 {
			fmt.Println("BEEP!")
		}
		e.soundTimer -= 1
	}
}

func (e *Emulator) Draw() {
	img := e.img.(draw.Image)
	for y := range 32 {
		for x := range 64 {
			pixel := stlval.Ternary(e.gfx[x][y], color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			for i := range 10 {
				for j := range 10 {
					img.Set(x*10+i, y*10+j, pixel)
				}
			}
		}
	}
}
