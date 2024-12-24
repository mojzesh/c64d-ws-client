package drawing

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func getMultiColorTabPix(color byte) [4]byte {
	return [4]byte{
		(color & 0b00000011) << 6,
		(color & 0b00000011) << 4,
		(color & 0b00000011) << 2,
		(color & 0b00000011) << 0,
	}
}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func getMultiColorTabPixMask() [4]byte {
	return [4]byte{
		(0b00000011) << 6,
		(0b00000011) << 4,
		(0b00000011) << 2,
		(0b00000011) << 0,
	}
}

// ---------------------------------------------------------
//
// ---------------------------------------------------------
func getMultiColorTabPixEvenOdd(patternValue byte) ([4]byte, [4]byte) {
	tabPixEven := [4]byte{
		((patternValue >> 2) & 0b00000011) << 6,
		((patternValue >> 0) & 0b00000011) << 4,
		((patternValue >> 2) & 0b00000011) << 2,
		((patternValue >> 0) & 0b00000011) << 0,
	}

	tabPixOdd := [4]byte{
		((patternValue >> 6) & 0b00000011) << 6,
		((patternValue >> 4) & 0b00000011) << 4,
		((patternValue >> 6) & 0b00000011) << 2,
		((patternValue >> 4) & 0b00000011) << 0,
	}
	return tabPixEven, tabPixOdd
}
