package drawing

import (
	"github.com/mojzesh/nested-cubes/vc64"
)

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func DrawCharsLineToFillNormal(
	x1 uint16,
	y1 uint16,
	x2 uint16,
	y2 uint16,
	color byte,
	startAddr uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	var deltaX float64
	var deltaY float64
	var stepY float64
	var stopX float64
	var x float64
	var y float64

	x1f := float64(x1)
	x2f := float64(x2)
	y1f := float64(y1)
	y2f := float64(y2)

	tabPix := getMultiColorTabPix(color)

	if x1f == x2f {
		return
	} else if x1f > x2f {
		deltaX = x2f - x1f
		deltaY = y2f - y1f
		x = x2f
		y = y2f
		stopX = x1f - 1
	} else {
		deltaX = x1f - x2f
		deltaY = y1f - y2f
		x = x1f
		y = y1f
		stopX = x2f - 1
	}

	stepY = deltaY / deltaX
	for {
		if x <= stopX {
			valAddr := startAddr + ((uint16(x) >> 2) * rows) + uint16(y)
			val := ram.CPUBusGetUByte8(valAddr)
			ram.CPUBusSetUByte8(valAddr, val^tabPix[uint16(x)&3])
			y = float64(y) + stepY
			x++
			continue
		}
		break
	}
}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func DrawBitmapLineToFillNormal(
	x1 uint16,
	y1 uint16,
	x2 uint16,
	y2 uint16,
	color byte,
	startAddr uint16,
	cols uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	var deltaX float64
	var deltaY float64
	var stepY float64
	var stopX float64
	var x float64
	var y float64

	x1f := float64(x1)
	x2f := float64(x2)
	y1f := float64(y1)
	y2f := float64(y2)

	tabPix := getMultiColorTabPix(color)

	if x1f == x2f {
		return
	} else if x1f > x2f {
		deltaX = x2f - x1f
		deltaY = y2f - y1f
		x = x2f
		y = y2f
		stopX = x1f - 1
	} else {
		deltaX = x1f - x2f
		deltaY = y1f - y2f
		x = x1f
		y = y1f
		stopX = x2f - 1
	}

	stepY = deltaY / deltaX
	for {
		if x <= stopX {
			valAddr := startAddr + ((uint16(x) & 0b11111100) << 1) + ((uint16(y) >> 3) * 320) + (uint16(y) & 0b00000111)
			val := ram.CPUBusGetUByte8(valAddr)
			ram.CPUBusSetUByte8(valAddr, val^tabPix[uint16(x)&3])
			y = float64(y) + stepY
			x++
			continue
		}
		break
	}
}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func DrawBitmapLineToFillEvenOdd(
	x1 uint16,
	y1 uint16,
	x2 uint16,
	y2 uint16,
	patternValue byte,
	startAddr uint16,
	cols uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	var deltaX float64
	var deltaY float64
	var stepY float64
	var stopX float64
	var x float64
	var y float64

	x1f := float64(x1)
	x2f := float64(x2)
	y1f := float64(y1)
	y2f := float64(y2)

	tabPixEven, tabPixOdd := getMultiColorTabPixEvenOdd(patternValue)

	if x1f == x2f {
		return
	} else if x1f > x2f {
		deltaX = x2f - x1f
		deltaY = y2f - y1f
		x = x2f
		y = y2f
		stopX = x1f - 1
	} else {
		deltaX = x1f - x2f
		deltaY = y1f - y2f
		x = x1f
		y = y1f
		stopX = x2f - 1
	}

	stepY = deltaY / deltaX
	for {
		if x <= stopX {
			valAddr := startAddr + ((uint16(x) & 0b11111100) << 1) + ((uint16(y) >> 3) * 320) + (uint16(y) & 0b00000111)
			val := ram.CPUBusGetUByte8(valAddr)
			ram.CPUBusSetUByte8(valAddr, val^tabPixEven[uint16(x)&3])

			yy := uint16(y) + (rows / 2)
			valAddr = startAddr + ((uint16(x) & 0b11111100) << 1) + ((yy >> 3) * 320) + ((yy) & 0b00000111)
			val = ram.CPUBusGetUByte8(valAddr)
			ram.CPUBusSetUByte8(valAddr, val^tabPixOdd[uint16(x)&3])

			y = float64(y) + stepY
			x++
			continue
		}
		break
	}
}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func DrawBitmapLineToFillEvenOddWithClipping(
	x1, y1, x2, y2 int16,
	patternValue byte,
	startAddr uint16,
	cols, rows uint16,
	maxX, maxY uint16,
	ram *vc64.C64RAM,
) {
	var deltaX float64
	var deltaY float64
	var stepY float64
	var stopX float64
	var x float64
	var y float64

	x1f := float64(x1)
	x2f := float64(x2)
	y1f := float64(y1)
	y2f := float64(y2)

	tabPixEven, tabPixOdd := getMultiColorTabPixEvenOdd(patternValue)

	if x1f == x2f {
		return
	} else if x1f > x2f {
		deltaX = x2f - x1f
		deltaY = y2f - y1f
		x = x2f
		y = y2f
		stopX = x1f - 1
	} else {
		deltaX = x1f - x2f
		deltaY = y1f - y2f
		x = x1f
		y = y1f
		stopX = x2f - 1
	}

	oddBufferOffsetY := uint16(8)
	stepY = deltaY / deltaX
	for {
		if x <= stopX {
			currY := y
			currX := x
			if currY < 0 {
				currY = 0
			}

			if currY > float64(maxY) {
				currY = float64(maxY)
			}

			if currX >= 0 && currX < float64(maxX) {
				valAddr := startAddr + ((uint16(currX) & 0b11111100) << 1) + ((uint16(currY) >> 3) * 320) + (uint16(currY) & 0b00000111)
				val := ram.CPUBusGetUByte8(valAddr)
				ram.CPUBusSetUByte8(valAddr, val^tabPixEven[uint16(currX)&3])

				yy := uint16(currY) + (rows / 2) + oddBufferOffsetY
				valAddr = startAddr + ((uint16(currX) & 0b11111100) << 1) + ((yy >> 3) * 320) + ((yy) & 0b00000111)
				val = ram.CPUBusGetUByte8(valAddr)
				ram.CPUBusSetUByte8(valAddr, val^tabPixOdd[uint16(x)&3])
			}

			y = float64(y) + stepY
			x++
			continue
		}
		break
	}
}
