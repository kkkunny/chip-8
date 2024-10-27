package emulator

import (
	"fmt"
	"image"
	"image/color"
	rand2 "math/rand/v2"
	"os"
	"time"

	"github.com/kkkunny/stl/container/stack"
	stlval "github.com/kkkunny/stl/value"
	"golang.org/x/exp/rand"
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

var KeyMap = map[uint16]uint8{
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

type Emulator struct {
	memory     [4096]uint8 // 内存 4K
	v          [16]uint8   // 16个寄存器
	i          uint16      // 索引寄存器
	pc         uint16      // 程序计数器
	delayTimer uint8
	soundTimer uint8
	stack      stack.Stack[uint16] // 栈
	gfx        [64][32]bool        // 屏幕
	keyboards  [16]bool            // 键盘

	img image.Image
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
	e.keyboards = [16]bool{}
	draw.Draw(e.img.(draw.Image), e.img.Bounds(), &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 0, A: 255}}, image.ZP, draw.Src)
}

func (e *Emulator) getVX(opcode Opcode) uint8    { return e.v[opcode.X()] }
func (e *Emulator) getVY(opcode Opcode) uint8    { return e.v[opcode.Y()] }
func (e *Emulator) setVX(opcode Opcode, x uint8) { e.v[opcode.X()] = x }
func (e *Emulator) setCarry(carry bool)          { e.v[0xF] = stlval.Ternary[uint8](carry, 1, 0) }
func (e *Emulator) KeyDown(key uint16) {
	index, ok := KeyMap[key]
	if !ok {
		return
	}
	e.keyboards[index] = true
}
func (e *Emulator) KeyUp(key uint16) {
	index, ok := KeyMap[key]
	if !ok {
		return
	}
	e.keyboards[index] = false
}

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
	fmt.Printf("0x%04X\n", opcode)
	switch opcode.Op() {
	case 0x0:
		switch opcode.N() {
		case 0x0: // 0x00E0: clear the screen
			e.gfx = [64][32]bool{}
			e.pc += 2
		case 0xE: // 0x00EE: ret
			e.pc = e.stack.Pop()
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0x1: // 1nnn: jump to address nnn
		e.pc = opcode.NNN()
	case 0x2: // 2nnn: call address nnn
		e.stack.Push(e.pc + 2)
		e.pc = opcode.NNN()
	case 0x3: // 3xkk: skip next instr if V[x] = kk
		e.pc += stlval.Ternary[uint16](e.getVX(opcode) == opcode.NN(), 4, 2)
	case 0x4: // 4xkk: skip next instr if V[x] != kk
		e.pc += stlval.Ternary[uint16](e.getVX(opcode) != opcode.NN(), 4, 2)
	case 0x5: // 5xy0: skip next instr if V[x] == V[y]
		e.pc += stlval.Ternary[uint16](e.getVX(opcode) == e.getVY(opcode), 4, 2)
	case 0x6: // 6xkk: set V[x] = kk
		e.setVX(opcode, opcode.NN())
		e.pc += 2
	case 0x7: // 7xkk: set V[x] = V[x] + kk
		e.setVX(opcode, e.getVX(opcode)+opcode.NN())
		e.pc += 2
	case 0x8: // 8xyn: Arithmetic stuff
		switch opcode.N() {
		case 0x0: // Set vy
			e.setVX(opcode, e.getVY(opcode))
		case 0x1: // OR
			e.setVX(opcode, e.getVX(opcode)|e.getVY(opcode))
		case 0x2: // AND
			e.setVX(opcode, e.getVX(opcode)&e.getVY(opcode))
		case 0x3: // XOR
			e.setVX(opcode, e.getVX(opcode)^e.getVY(opcode))
		case 0x4: // Add vy
			e.setCarry((uint16(e.getVX(opcode)) + uint16(e.getVY(opcode))) > 255)
			e.setVX(opcode, e.getVX(opcode)+e.getVY(opcode))
		case 0x5: // Sub vy
			e.setCarry(e.getVX(opcode) > e.getVY(opcode))
			e.setVX(opcode, e.getVX(opcode)-e.getVY(opcode))
		case 0x6: // Shift right
			e.setCarry(e.getVX(opcode)&0x01 != 0)
			e.setVX(opcode, e.getVX(opcode)>>1)
		case 0x7: // Sub from vy
			e.setCarry(e.getVX(opcode) < e.getVY(opcode))
			e.setVX(opcode, e.getVY(opcode)-e.getVX(opcode))
		case 0xE: // Shift left
			e.setCarry((e.getVX(opcode)>>7)&0x01 != 0)
			e.setVX(opcode, e.getVX(opcode)<<1)
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
		e.pc += 2
	case 0x9:
		switch opcode.N() {
		case 0xE: // 9xy0: skip instruction if Vx != Vy
			e.pc += stlval.Ternary[uint16](e.getVX(opcode) != e.getVY(opcode), 4, 2)
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0xA: // Annn: set I to address nnn
		e.i = opcode.NNN()
		e.pc += 2
	case 0xB: // Bnnn: jump to location nnn + V[0]
		e.pc = opcode.NNN() + uint16(e.v[0])
	case 0xC: // Cxkk: V[x] = random byte AND kk
		val := uint8(rand2.New(rand.NewSource(uint64(time.Now().Unix()))).UintN(256))
		e.setVX(opcode, val&opcode.NN())
		e.pc += 2
	case 0xD:
		// Dxyn: Display an n-byte sprite starting at memory
		// location I at (Vx, Vy) on the screen, VF = collision
		e.setCarry(false)
		var collision bool

		for j := range uint16(opcode.N()) {
			row := uint16(e.memory[e.i+j])
			for i := range uint16(8) {
				newPixel := (row >> (7 - i) & 0x01) != 0
				if newPixel {
					xi, yj := (uint16(e.getVX(opcode))+i)%64, (uint16(e.getVY(opcode))+j)%32
					oldPixel := &e.gfx[xi][yj]
					if *oldPixel {
						collision = true
					}
					*oldPixel = newPixel != *oldPixel
				}
			}
		}

		e.setCarry(collision)
		e.pc += 2
	case 0xE: // key-pressed events
		switch opcode.NN() {
		case 0x9E: // skip next instr if key[Vx] is pressed
			e.pc += stlval.Ternary[uint16](e.keyboards[e.getVX(opcode)], 4, 2)
		case 0xA1: // skip next instr if key[Vx] is not pressed
			e.pc += stlval.Ternary[uint16](!e.keyboards[e.getVX(opcode)], 4, 2)
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0xF: // misc
		switch opcode.NN() {
		case 0x07:
			e.setVX(opcode, e.delayTimer)
			e.pc += 2
		case 0x0A:
			var isPressed bool

			for i, pressed := range e.keyboards {
				if pressed {
					isPressed = true
					e.setVX(opcode, uint8(i))
					break
				}
			}

			if isPressed {
				e.pc += 2
			}
		case 0x15:
			e.delayTimer = e.getVX(opcode)
			e.pc += 2
		case 0x18:
			e.soundTimer = e.getVX(opcode)
			e.pc += 2
		case 0x1E:
			e.setCarry(e.i+uint16(e.getVX(opcode)) > 0x0FFF)
			e.i += uint16(e.getVX(opcode))
			e.pc += 2
		case 0x29:
			e.i = 5 * uint16(e.getVX(opcode))
			e.pc += 2
		case 0x33:
			e.memory[e.i] = e.getVX(opcode) / 100
			e.memory[e.i+1] = (e.getVX(opcode) % 100) / 10
			e.memory[e.i+2] = e.getVX(opcode) % 10
			e.pc += 2
		case 0x55:
			for i := range opcode.X() + 1 {
				e.memory[e.i+uint16(i)] = e.v[i]
			}
			e.i += uint16(opcode.X()) + 1
			e.pc += 2
		case 0x65:
			for i := range opcode.X() + 1 {
				e.v[i] = e.memory[e.i+uint16(i)]
			}
			e.i += uint16(opcode.X()) + 1
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
	// fmt.Printf("Opcode: 0x%04X\n", opcode)
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
