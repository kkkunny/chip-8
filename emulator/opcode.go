package emulator

type Opcode uint16

func NewOpcode(high, low uint8) Opcode {
	return Opcode(high)<<8 | Opcode(low)
}

func (op Opcode) High() uint8 { return uint8((op >> 8) & 0x00ff) }
func (op Opcode) Low() uint8  { return uint8(op & 0x00ff) }
func (op Opcode) Op() uint8   { return (op.High() >> 4) & 0x0F }
func (op Opcode) N() uint8    { return op.Low() & 0x0F }
func (op Opcode) NN() uint8   { return op.Low() }
func (op Opcode) NNN() uint16 { return uint16(op.X())<<8 | uint16(op.Low()) }
func (op Opcode) X() uint8    { return op.High() & 0x0F }
func (op Opcode) Y() uint8    { return (op.Low() >> 4) & 0x0F }
