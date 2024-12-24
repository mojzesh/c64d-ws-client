package vc64

import (
	"fmt"
	"log"
	"strings"
)

// -------------------------------------------
//
// -------------------------------------------
type VC64 struct {
	ram *C64RAM
}

// -------------------------------------------
//
// -------------------------------------------
func NewVC64() *VC64 {
	vc64 := new(VC64)
	vc64.ram = NewC64RAM()
	return vc64
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) GetRAM() *C64RAM {
	return vc64.ram
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) LoadPRG(fileName string) int {
	data, err := ReadFileAsBytes(fileName)
	if err != nil {
		panic(err)
	}
	var dataLength = len(data)
	startAddress := uint16(data[0]) | uint16(data[1])<<8
	for c0 := 0; c0 < dataLength-2; c0++ {
		vc64.ram.CPUBusSetUByte8(startAddress+uint16(c0), data[c0+2])
	}
	return dataLength
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) SavePRG(fileName string, startAddress uint16, endAddress uint16) {
	var bytesCount int = int(endAddress) - int(startAddress) + 1
	if bytesCount < 0 {
		panic("endAddress is smaller than startAddress!")
	}
	outputArray := make([]byte, bytesCount+2)
	outputArray[0] = byte(startAddress & 0xFF)
	outputArray[1] = byte((startAddress >> 8) & 0xFF)
	for c0 := 0; c0 < int(bytesCount); c0++ {
		outputArray[c0+2] = vc64.ram.CPUBusGetUByte8(startAddress + uint16(c0))
	}
	err := WriteBytesToFile(fileName, outputArray)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) LoadBIN(fileName string, startAddress uint16) uint16 {
	data, err := ReadFileAsBytes(fileName)
	if err != nil {
		panic(err)
	}
	dataLength := len(data)
	for c0 := 0; c0 < dataLength; c0++ {
		vc64.ram.CPUBusSetUByte8(startAddress+uint16(c0), data[c0])
	}
	return uint16(dataLength)
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) SaveBIN(fileName string, startAddress uint16, endAddress uint16) {
	var bytesCount int = int(endAddress) - int(startAddress) + 1
	if bytesCount < 0 {
		panic("endAddress is smaller than startAddress!")
	}
	var outputArray = make([]byte, bytesCount)

	for c0 := 0; c0 < int(bytesCount); c0++ {
		outputArray[c0] = vc64.ram.CPUBusGetUByte8(startAddress + uint16(c0))
	}
	err := WriteBytesToFile(fileName, outputArray)
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) DumpRAM(startAddress uint16, endAddress uint16, columnsCount uint16, flags DUMPSTYLE) {
	log.Println("RAM Dumped")

	if columnsCount == 0 {
		columnsCount = 16
	}

	if startAddress > endAddress {
		panic("Error: StartAddress can't be higher than EndAddress !")
	}

	log.Println("")
	log.Println("Virtual C64 - Starting memory dump")
	log.Println("RAM Dump start: " + UInt16ToHexString(startAddress, "0x", "") + " end: " + UInt16ToHexString(endAddress, "0x", ""))
	log.Println("")

	valuesPrefix := ""
	separatorAfterLastColumn := false

	if (flags & ACTION_REPLAY) != 0 {
		valuesPrefix = ""
		separatorAfterLastColumn = true
	} else if (flags & KICK_ASS) != 0 {
		valuesPrefix = "$"
		separatorAfterLastColumn = false
	}

	bytesString := ""
	addressString := ""

	outputString := ""
	var bytesCount uint16 = endAddress - startAddress + 1
	var offset uint16 = 0
	var address uint16
	stopFlag := false

	var columnsTmp uint16

	for !stopFlag {
		address = startAddress + offset
		if offset > (bytesCount - 1) {
			break
		} else {
			if (bytesCount - offset) >= columnsCount {
				columnsTmp = columnsCount
			} else {
				columnsTmp = ((bytesCount - offset) % columnsCount)
				stopFlag = true
			}
		}

		offset += columnsCount

		bytesString = vc64.formatOneLine(address, columnsTmp, "", valuesPrefix, separatorAfterLastColumn)
		addressString = UInt16ToHexString(address, "", "")

		if bytesString != "" {

			if (flags & ACTION_REPLAY) != 0 {
				outputString += addressString + "  " + bytesString + "\n"
			} else if (flags & KICK_ASS) != 0 {
				outputString += ".byte " + bytesString + "\n"
			}
		}
	}
	outputString += "\n"

	fmt.Println(outputString)
}

// -------------------------------------------
//
// -------------------------------------------
func (vc64 *VC64) formatOneLine(startAddress uint16, bytesCount uint16, valuesPrefix string, valuesSeparator string, separatorAfterLastColumn bool) string {
	var address uint16
	var offset uint16 = 0
	var byteValue byte
	var bytesString string = ""

	for c0 := 0; c0 < int(bytesCount); c0++ {
		address = startAddress + offset
		byteValue = vc64.ram.CPUBusGetUByte8(address)
		if (c0 == int((bytesCount - 1))) && !separatorAfterLastColumn {
			bytesString += UInt8ToHexString(byteValue, valuesPrefix, "")
		} else {
			bytesString += UInt8ToHexString(byteValue, valuesPrefix, "") + valuesSeparator
		}
		offset++
	}
	return strings.TrimSpace(bytesString)
}
