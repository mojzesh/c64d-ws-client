package vc64

type BUSType int

const (
	CPU_BUS BUSType = iota
	VIC_BUS
)

const (
	MEM_CONFIG_0001         uint16 = 0x0001
	MEM_CONFIG_0001_DEFAULT byte   = 0x37
)

// --------------------------------------------------------------
// 0x0001 - Memory config:
// Bits #0-#2: Configuration for memory areas:
// - $A000-$BFFF
// - $D000-$DFFF
// - $E000-$FFFF
// Values:
// %x00: RAM visible in all three areas.
// %x01: RAM visible at $A000-$BFFF and $E000-$FFFF.
// %x10: RAM visible at $A000-$BFFF; KERNAL ROM visible at $E000-$FFFF.
// %x11: BASIC ROM visible at $A000-$BFFF; KERNAL ROM visible at $E000-$FFFF.
// %0xx: Character ROM visible at $D000-$DFFF. (Except for the value %000, see above.)
// %1xx: I/O area visible at $D000-$DFFF. (Except for the value %100, see above.)
// --------------------------------------------------------------
type C64RAM struct {
	ram       [0x10000]byte // RAM: $0000-$FFFF
	vicRAM    [0x10000]byte // RAM: $0000-$FFFF
	io        [0x1000]byte  // RAM: $D000-$DFFF
	basicROM  [0x2000]byte  // RAM: $A000-$BFFF
	charROM   [0x1000]byte  // RAM: $D000-$DFFF
	kernalROM [0x2000]byte  // RAM: $E000-$FFFF
}

func NewC64RAM() *C64RAM {
	ram := new(C64RAM)
	ram.CPUBusSetUByte8(MEM_CONFIG_0001, MEM_CONFIG_0001_DEFAULT)
	return ram
}

func (ram *C64RAM) clearCPURAM(initialValue byte) {
	for idx := range ram.ram {
		ram.CPUBusSetUByte8(uint16(idx), initialValue)
	}
}

func (ram *C64RAM) clearVICRAM(initialValue byte) {
	for idx := range ram.vicRAM {
		ram.VICBusSetUByte8(uint16(idx), initialValue)
	}
}

func (ram *C64RAM) CPUFillRAM(value byte, startAddr uint16, count uint16) {
	for idx := startAddr; idx < startAddr+count; idx++ {
		ram.CPUBusSetUByte8(idx, value)
	}
}

func (ram *C64RAM) FillVICRAM(value byte, startAddr uint16, count uint16) {
	for idx := startAddr; idx < startAddr+count; idx++ {
		ram.VICBusSetUByte8(idx, value)
	}
}

func (ram *C64RAM) CPUBusGetUByte8(address uint16) byte {
	return ram.ram[address]
}

func (ram *C64RAM) CPUGetRAMSlice(startAddr uint16, count uint16) []byte {
	return ram.ram[startAddr : startAddr+count]
}

func (ram *C64RAM) CPUBusSetUByte8(address uint16, b byte) {
	ram.ram[address] = b
}

func (ram *C64RAM) VICBusGetUByte8(address uint16) byte {
	return ram.vicRAM[address]
}

func (ram *C64RAM) VICBusSetUByte8(address uint16, b byte) {
	ram.vicRAM[address] = b
}
