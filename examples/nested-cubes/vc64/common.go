package vc64

import (
	"errors"
	"fmt"
	"os"
)

type DUMPSTYLE int

const (
	ACTION_REPLAY DUMPSTYLE = 0x01
	KICK_ASS      DUMPSTYLE = 0x02
)

func UInt8ToHexString(inputValue byte, prefix string, postfix string) string {
	bytesReadHexStr := fmt.Sprintf("%x", inputValue)
	if len(bytesReadHexStr) == 1 {
		return prefix + "0" + bytesReadHexStr
	}
	return prefix + bytesReadHexStr + postfix
}

func UInt16ToHexString(inputValue uint16, prefix string, postfix string) string {
	bytesReadHexStr := fmt.Sprintf("%x", inputValue)
	length := len(bytesReadHexStr)
	if len(bytesReadHexStr) < 4 {
		for c0 := 0; c0 < (4 - length); c0++ {
			bytesReadHexStr = "0" + bytesReadHexStr
		}
	}
	return prefix + bytesReadHexStr + postfix
}

func UInt32ToHexString(inputValue uint32, prefix string, postfix string) string {
	bytesReadHexStr := fmt.Sprintf("%x", inputValue)
	length := len(bytesReadHexStr)
	if len(bytesReadHexStr) < 8 {
		for c0 := 0; c0 < (8 - length); c0++ {
			bytesReadHexStr = "0" + bytesReadHexStr
		}
	}
	return prefix + bytesReadHexStr + postfix
}

// DoesFileExist
func DoesFileExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// ----------------------------------------------------------------
// Low level io interface
// ----------------------------------------------------------------
func ReadFileAsBytes(path string) ([]byte, error) {

	if !DoesFileExist(path) {
		return nil, errors.New("Provided file path doesn't exist: " + path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	bufferSize := uint64(fileinfo.Size())

	buffer := make([]byte, bufferSize)

	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func WriteBytesToFile(filePath string, data []byte) error {
	//--------------------------------------------------------
	// open file for binary write
	//--------------------------------------------------------
	fileHandle, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Println("Unable to write to ", filePath)
			fmt.Println(err)
			return err
		}
	}
	if fileHandle == nil {
		// something goes wrong
		return errors.New(filePath)
	}

	//--------------------------------------------------------
	// write file
	_, err = fileHandle.Write(data)
	if err != nil {
		return err
	}
	//--------------------------------------------------------
	// close file
	fileHandle.Close()
	return nil
}
