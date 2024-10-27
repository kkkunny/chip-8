package os

type opcode uint16

func newOpcode(high, low uint8) opcode {
	return opcode(high)<<8 | opcode(low)
}

func (op opcode) High() uint8 { return uint8((op >> 8) & 0x00ff) }
func (op opcode) Low() uint8  { return uint8(op & 0x00ff) }
func (op opcode) Op() uint8   { return (op.High() >> 4) & 0x0F }
func (op opcode) N() uint8    { return op.Low() & 0x0F }
func (op opcode) NN() uint8   { return op.Low() }
func (op opcode) NNN() uint16 { return uint16(op.X())<<8 | uint16(op.Low()) }
func (op opcode) X() uint8    { return op.High() & 0x0F }
func (op opcode) Y() uint8    { return (op.Low() >> 4) & 0x0F }
