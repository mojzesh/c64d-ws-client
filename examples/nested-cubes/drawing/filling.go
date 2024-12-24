package drawing

import (
	"github.com/mojzesh/nested-cubes/vc64"
)

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func EORFillBitmap(
	frameBufferAddr uint16,
	bitmapAddr uint16,
	initValue byte,
	columns uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	for column := uint16(0); column < columns; column++ {
		value := initValue
		for row := uint16(0); row < rows; row++ {
			fragAddr := column*8 + ((uint16(row) >> 3) * 320) + (uint16(row) & 0x07)
			value = value ^ ram.CPUBusGetUByte8(frameBufferAddr+fragAddr)
			ram.CPUBusSetUByte8(bitmapAddr+fragAddr, value)
		}
	}

}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func EORFillChars(
	frameBufferAddr uint16,
	bitmapAddr uint16,
	initValue byte,
	columns uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	for column := uint16(0); column < columns; column++ {
		value := initValue
		for row := uint16(0); row < rows; row++ {
			value = value ^ ram.CPUBusGetUByte8(frameBufferAddr+(column*rows)+row)
			ram.CPUBusSetUByte8(bitmapAddr+(column*rows)+row, value)
		}
	}

}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func EORFillEvenOddBitmap(
	frameBufferAddr uint16,
	bitmapAddr uint16,
	initValue byte,
	columns uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	oddBufferOffsetY := uint16(8) // sync this value with line drawing routine
	for column := uint16(0); column < columns; column++ {
		evenValue := initValue
		oddValue := initValue
		for row := uint16(0); row < rows; row++ {
			inputRow := uint16(row)
			outputRow := uint16((row * 2) + 0)
			fragInputAddr := column*8 + ((inputRow >> 3) * 320) + (inputRow & 0x07)
			fragOutputAddr := column*8 + ((outputRow >> 3) * 320) + (outputRow & 0x07)
			evenValue = evenValue ^ ram.CPUBusGetUByte8(frameBufferAddr+fragInputAddr)
			ram.CPUBusSetUByte8(bitmapAddr+fragOutputAddr, evenValue)

			inputRow = uint16(row + (rows / 2) + oddBufferOffsetY)
			outputRow = uint16((row * 2) + 1)
			fragInputAddr = column*8 + ((inputRow >> 3) * 320) + (inputRow & 0x07)
			fragOutputAddr = column*8 + ((outputRow >> 3) * 320) + (outputRow & 0x07)
			oddValue = oddValue ^ ram.CPUBusGetUByte8(frameBufferAddr+fragInputAddr)
			ram.CPUBusSetUByte8(bitmapAddr+fragOutputAddr, oddValue)
		}
	}
}
