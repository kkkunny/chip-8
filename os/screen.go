package os

import (
	"image/color"

	stlval "github.com/kkkunny/stl/value"
)

type screen struct {
	onUpdate func(uint16, uint16, color.Color)
	onReset  func()
	pixels   [64][32]bool
}

func newScreen() *screen {
	return &screen{}
}

func (s *screen) Reset() {
	s.pixels = [64][32]bool{}
	if s.onReset != nil {
		s.onReset()
	}
}

func (s *screen) SetOnScreenUpdate(onScreenUpdate func(uint16, uint16, color.Color)) {
	s.onUpdate = onScreenUpdate
}

func (s *screen) SetOnReset(onReset func()) {
	s.onReset = onReset
}

func (s *screen) UpdatePixelsFromMemory(x, y uint8, memory []uint8) (collision bool) {
	for j, row := range memory {
		for i := range uint16(8) {
			newPixel := (row >> (7 - i) & 0x01) != 0
			if newPixel {
				xi, yj := (uint16(x)+i)%64, (uint16(y)+uint16(j))%32
				oldPixel := s.pixels[xi][yj]
				collision = collision || oldPixel
				newPixel = newPixel != oldPixel
				s.pixels[xi][yj] = newPixel
				if s.onUpdate != nil {
					clr := stlval.Ternary(newPixel, color.RGBA{R: 255, G: 255, B: 255, A: 255}, color.RGBA{R: 0, G: 0, B: 0, A: 255})
					s.onUpdate(xi, yj, clr)
				}
			}
		}
	}
	return collision
}
