package os

import (
	"fmt"
	"image/color"
	rand2 "math/rand/v2"
	"os"
	"time"

	"github.com/kkkunny/stl/container/stack"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"
	stlval "github.com/kkkunny/stl/value"
	"golang.org/x/exp/rand"

	"github.com/kkkunny/chip-8/config"
)

type CPU struct {
	v          [16]uint8 // 16个寄存器
	i          uint16    // 索引寄存器
	pc         uint16    // 程序计数器
	delayTimer uint8
	soundTimer uint8

	memory   *memory
	stack    stack.Stack[uint16]
	screen   *screen
	keyboard *keyboard
	audio    *audio
}

func NewCPU(keyboardInput <-chan tuple.Tuple2[KeyEvent, uint8]) *CPU {
	return &CPU{
		memory:   newMemory(),
		stack:    stack.New[uint16](),
		screen:   newScreen(),
		keyboard: newKeyboard(keyboardInput),
		audio:    newAudio(),
	}
}

func (e *CPU) SetOnScreenUpdate(onScreenUpdate func(uint16, uint16, color.Color)) {
	e.screen.SetOnScreenUpdate(onScreenUpdate)
}

func (e *CPU) SetOnReset(onReset func()) {
	e.screen.SetOnReset(onReset)
}

func (e *CPU) Reset() {
	config.Logger.Debug("cpu reset")
	e.v = [16]uint8{}
	e.i = 0
	e.pc = 0x200
	e.delayTimer = 0
	e.soundTimer = 0
	e.memory.Reset()
	e.stack.Clear()
	e.screen.Reset()
	e.keyboard.Reset()
}

func (e *CPU) getVX(opcode opcode) uint8    { return e.v[opcode.X()] }
func (e *CPU) getVY(opcode opcode) uint8    { return e.v[opcode.Y()] }
func (e *CPU) setVX(opcode opcode, x uint8) { e.v[opcode.X()] = x }
func (e *CPU) setCarry(carry bool)          { e.v[0xF] = stlval.Ternary[uint8](carry, 1, 0) }

func (e *CPU) Load(path string) error {
	config.Logger.Debugf("load %s", path)
	info, err := stlerr.ErrorWith(os.Stat(path))
	if err != nil {
		return err
	}

	file, err := stlerr.ErrorWith(os.Open(path))
	if err != nil {
		return err
	}
	defer file.Close()

	err = e.memory.Load(file, uint(info.Size()))
	return err
}

func (e *CPU) executeOpcode(opcode opcode) {
	// config.Logger.Debugf("execute opcode: 0x%04X", opcode)
	switch opcode.Op() {
	case 0x0:
		switch opcode.N() {
		case 0x0: // 0x00E0: clear the screen
			e.screen.Reset()
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
		case 0x0: // 9xy0: skip instruction if Vx != Vy
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
		collision := e.screen.UpdatePixelsFromMemory(e.getVX(opcode), e.getVY(opcode), e.memory.IndexByN(uint(e.i), uint(opcode.N())))
		e.setCarry(collision)
		e.pc += 2
	case 0xE: // key-pressed events
		switch opcode.NN() {
		case 0x9E: // skip next instr if key[Vx] is pressed
			e.pc += stlval.Ternary[uint16](e.keyboard.Pressed(e.getVX(opcode)), 4, 2)
		case 0xA1: // skip next instr if key[Vx] is not pressed
			e.pc += stlval.Ternary[uint16](!e.keyboard.Pressed(e.getVX(opcode)), 4, 2)
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	case 0xF: // misc
		switch opcode.NN() {
		case 0x07:
			e.setVX(opcode, e.delayTimer)
			e.pc += 2
		case 0x0A:
			firstPressedKey, ok := e.keyboard.FirstPressed()
			if ok {
				e.setVX(opcode, firstPressedKey)
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
			e.memory.Set(uint(e.i), e.getVX(opcode)/100)
			e.memory.Set(uint(e.i)+1, (e.getVX(opcode)%100)/10)
			e.memory.Set(uint(e.i)+2, e.getVX(opcode)%10)
			e.pc += 2
		case 0x55:
			for i := range opcode.X() + 1 {
				e.memory.Set(uint(e.i)+uint(i), e.v[i])
			}
			e.pc += 2
		case 0x65:
			for i := range opcode.X() + 1 {
				e.v[i] = e.memory.Get(uint(e.i) + uint(i))
			}
			e.pc += 2
		default:
			panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
		}
	default:
		panic(fmt.Sprintf("Unknown opcode: 0x%04X", opcode))
	}
}

func (e *CPU) Next() {
	op := newOpcode(e.memory.Get(uint(e.pc)), e.memory.Get(uint(e.pc)+1))
	e.executeOpcode(op)
}

func (e *CPU) Ticker() {
	if e.delayTimer > 0 {
		e.delayTimer -= 1
	}
	if e.soundTimer == 1 {
		e.audio.Play()
	}
	if e.soundTimer > 0 {
		e.soundTimer -= 1
	}
}
