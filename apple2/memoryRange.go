package apple2

type memoryRange struct {
	base uint16
	data []uint8
}

func newMemoryRange(base uint16, data []uint8) *memoryRange {
	var m memoryRange
	m.base = base
	m.data = data
	return &m
}

func (m *memoryRange) peek(address uint16) uint8 {
	return m.data[address-m.base]
}

func (m *memoryRange) poke(address uint16, value uint8) {
	m.data[address-m.base] = value
}

func (m *memoryRange) subRange(a, b uint16) []uint8 {
	return m.data[a-m.base : b-m.base]
}
