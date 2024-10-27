package os

import (
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/chip-8/config"
)

type KeyEvent uint8

const (
	KeyEventDown KeyEvent = iota
	KeyEventUp
)

type keyboard struct {
	input <-chan tuple.Tuple2[KeyEvent, uint8]
	keys  [16]bool
}

func newKeyboard(input <-chan tuple.Tuple2[KeyEvent, uint8]) *keyboard {
	k := &keyboard{input: input}
	go func() {
		for key := range k.input {
			switch key.E1() {
			case KeyEventDown:
				config.Logger.Debugf("keyboard event: down %d", key.E2())
				k.KeyDown(key.E2())
			case KeyEventUp:
				config.Logger.Debugf("keyboard event: up %d", key.E2())
				k.KeyUp(key.E2())
			}
		}
	}()
	return k
}

func (k *keyboard) Reset() {
	k.keys = [16]bool{}
}

func (k *keyboard) KeyDown(key uint8) {
	k.keys[key] = true
}

func (k *keyboard) KeyUp(key uint8) {
	k.keys[key] = false
}

func (k *keyboard) Pressed(key uint8) bool {
	return k.keys[key]
}

func (k *keyboard) FirstPressed() (uint8, bool) {
	for i, pressed := range k.keys {
		if pressed {
			return uint8(i), true
		}
	}
	return 0, false
}
