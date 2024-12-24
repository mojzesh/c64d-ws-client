package tests

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mojzesh/c64d-ws-client/c64dws"
	"gotest.tools/assert"
)

// ----------------------------------------------------------------------
// Assert Successfull Connection
// ----------------------------------------------------------------------
func assertSuccessfullConnection(t *testing.T, response []byte, err error) {
	assert.NilError(t, err)
	assert.Equal(t, string(response), "")
}

// ----------------------------------------------------------------------
// Helper function to assert success response
// ----------------------------------------------------------------------
func assertSuccessResponse(t *testing.T, msgType c64dws.C64DMessageType, msg any) (*c64dws.RequestResult, *c64dws.RequestError) {
	assert.Equal(t, msgType, c64dws.C64DRequestResponse)
	requestResult, requestError := c64dws.GetResultOrError(msg)
	assert.Check(t, requestResult != nil)
	assert.Check(t, requestError == nil)
	assert.Equal(t, requestResult.Status, 200)
	return requestResult, requestError
}

// ----------------------------------------------------------------------
// Reset types
// ----------------------------------------------------------------------
type resetType int

const (
	softReset resetType = iota
	hardReset
)

// ----------------------------------------------------------------------
// HardReset helper function
// ----------------------------------------------------------------------
func hardOrSoftReset(t *testing.T, client *c64dws.Client, resetType resetType) {
	if resetType == softReset {
		err := client.SoftReset()
		assert.NilError(t, err)
	} else {
		err := client.HardReset()
		assert.NilError(t, err)
	}

	// Receive response (this blocks until a message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}

	assertSuccessResponse(t, msgType, msg)
}

// ----------------------------------------------------------------------
// Device types
// ----------------------------------------------------------------------
type deviceType int

const (
	c64RAM deviceType = iota
	c64CPUMemory
	drive1541RAM
	drive1541CPUMemory
)

// ----------------------------------------------------------------------
// Helper function to test RAMClear for given device
// ----------------------------------------------------------------------
func testClearRAMForGivenDevice(t *testing.T, client *c64dws.Client, address uint16, bytesCount uint16, value byte, deviceType deviceType) {
	var err error
	// -------------------------------------------------------------
	// test: RAMClear or Drive1541RAMClear
	// -------------------------------------------------------------
	switch deviceType {
	case c64RAM:
		err = client.RAMClear(address, bytesCount, value)
	case drive1541RAM:
		err = client.Drive1541RAMClear(address, bytesCount, value)
	default:
		err = fmt.Errorf("unknown device type")
	}
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	_, _ = assertSuccessResponse(t, msgType, msg)
}

// -------------------------------------------------------------
// test one of the following:
// - RAMWriteBlock
// - CPUMemoryWriteBlock
// - Drive1541RAMWriteBlock
// - Drive1541CPUMemoryWriteBlock
// -------------------------------------------------------------
func testWriteMemoryForGivenDevice(t *testing.T, client *c64dws.Client, address uint16, data []byte, deviceType deviceType) {
	var err error
	switch deviceType {
	case c64RAM:
		err = client.RAMWriteBlock(address, data)
	case c64CPUMemory:
		err = client.CPUMemoryWriteBlock(address, data)
	case drive1541RAM:
		err = client.Drive1541RAMWriteBlock(address, data)
	case drive1541CPUMemory:
		err = client.Drive1541CPUMemoryWriteBlock(address, data)
	default:
		err = fmt.Errorf("unknown device type")
	}
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	_, _ = assertSuccessResponse(t, msgType, msg)

}

// -------------------------------------------------------------
// test one of the following:
// - RAMReadBlock
// - CPUMemoryReadBlock
// - Drive1541RAMReadBlock
// - Drive1541CPUMemoryReadBlock
// -------------------------------------------------------------
func testReadMemoryForGivenDevice(t *testing.T, client *c64dws.Client, address uint16, bytesCount uint16, deviceType deviceType) []byte {
	var err error
	switch deviceType {
	case c64RAM:
		err = client.RAMReadBlock(address, bytesCount)
	case c64CPUMemory:
		err = client.CPUMemoryReadBlock(address, bytesCount)
	case drive1541RAM:
		err = client.Drive1541RAMReadBlock(address, bytesCount)
	case drive1541CPUMemory:
		err = client.Drive1541CPUMemoryReadBlock(address, bytesCount)
	default:
		err = fmt.Errorf("unknown device type")
	}
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)

	return requestResult.BinaryData
}

// ----------------------------------------------------------------------
//
// ----------------------------------------------------------------------
func convertArrayOfArraysToRegistersNBitMap[T uint8 | uint16](registers []any) map[T]byte {
	var registersMap = map[T]byte{}
	for _, v := range registers {
		address := T(v.([]any)[0].(float64))
		value := byte(v.([]any)[1].(float64))
		registersMap[address] = value
	}

	return registersMap
}

// ----------------------------------------------------------------------
// Test if the array only has the expected value and nothing else
// ----------------------------------------------------------------------
func testArrayHasOnlyExpectedValue(t *testing.T, data []byte, expectedValue byte) {
	for i := 0; i < len(data); i++ {
		assert.Equal(t, data[i], expectedValue)
	}
}

// ----------------------------------------------------------------------
// Get a slice of consecutive bytes
// ----------------------------------------------------------------------
func getSliceOfConsecutiveBytes(start byte, count int) []byte {
	data := make([]byte, count)
	for i := 0; i < count; i++ {
		data[i] = start + byte(i)
	}
	return data
}

// ----------------------------------------------------------------------
// Download a file from the internet
// ----------------------------------------------------------------------
func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------
// Unzip a file to a temporary directory
// ----------------------------------------------------------------------
func unzipToTempDir(zipFilePath string) (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "unzipped")
	if err != nil {
		return "", err
	}

	// Open the zip file
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", err
	}
	defer zipReader.Close()

	// Iterate through each file in the zip archive
	for _, file := range zipReader.File {
		filePath := filepath.Join(tempDir, file.Name)

		// Create directories if necessary
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// Create the file
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", err
		}

		// Open the file inside the zip archive
		rc, err := file.Open()
		if err != nil {
			return "", err
		}

		// Copy the contents of the file
		_, err = io.Copy(outFile, rc)

		// Close the file and the zip archive file
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}
	}

	return tempDir, nil
}
