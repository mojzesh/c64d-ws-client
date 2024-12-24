package tests

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mojzesh/c64d-ws-client/c64dws"
	"gotest.tools/assert"
)

const DefaultURL = "ws://localhost:3563/stream"

func TestDefaultClient(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	assert.Equal(t, client.GetURL(), DefaultURL)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)
}

func TestCustomClient(t *testing.T) {
	const hostName = "localhost"
	const port = 3563
	const supportedScheme = "ws"
	const unsupportedScheme = "wss"
	// test: invalid port
	_, err := c64dws.GetCustomHost(hostName, 0, supportedScheme)
	assert.Error(t, err, "Invalid host description")

	// test: invalid scheme
	_, err = c64dws.GetCustomHost(hostName, port, unsupportedScheme)
	assert.Error(t, err, "Currently only 'ws' scheme is supported")

	// test: valid host description
	hostDesc, err := c64dws.GetCustomHost(hostName, port, supportedScheme)
	assert.NilError(t, err)

	// test: valid custom client
	client := c64dws.NewCustomClient(c64dws.EmulatorC64, c64dws.StreamAPI, c64dws.TokenTypeAutoIncrement, "ID:%s", hostDesc)
	assert.Check(t, client != nil)
	defer client.Close()

	assert.Equal(t, client.GetURL(), DefaultURL)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)
}

func TestGetAPIFn(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	apiFn := client.GetAPIFn(c64dws.APIFnLoad)
	assert.Equal(t, apiFn, "load")

	apiFn = client.GetAPIFn(c64dws.APIFnContinue)
	assert.Equal(t, apiFn, "c64/continue")
}

func TestHardReset(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	hardOrSoftReset(t, client, hardReset)
}

func TestSoftReset(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	hardOrSoftReset(t, client, softReset)
}

func TestDetachEverything(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	err = client.DetachEverything()
	assert.NilError(t, err)
}

func TestCPUStatus(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: CPUStatus
	// -------------------------------------------------------------
	err = client.CPUStatus()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Check(t, requestResult.Result != nil)
	// t.Log(requestResult.Result)

	// -------------------------------------------------------------
	// test if the response contains expected keys
	// -------------------------------------------------------------
	expectedKeys := []string{
		"p",
		"a",
		"x",
		"y",
		"pc",
		"sp",
		"memory0001",
		"instructionCycle",
		"rasterCycle",
		"rasterX",
		"rasterY",
		"game",
		"exrom",
	}
	for _, key := range expectedKeys {
		_, exist := (*requestResult.Result)[key]
		assert.Check(t, exist)
	}

	assert.Equal(t, requestResult.Status, 200)
}

func TestCPUMakeJMP(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	jmpAddr := uint16(0x0815)
	infiniteLoopProcedure := []byte{
		0x78,                                                       // $0815: sei
		0x4c, byte((jmpAddr + 1) & 0xff), byte((jmpAddr + 1) >> 8), // $0816: jmp $0816
	}
	err = client.RAMWriteBlock(jmpAddr, infiniteLoopProcedure)
	assert.NilError(t, err)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	_, _ = assertSuccessResponse(t, msgType, msg)

	// Run the program
	err = client.CPUMakeJMP(jmpAddr)
	assert.NilError(t, err)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	_, _ = assertSuccessResponse(t, msgType, msg)
}

func TestCPUCounters(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: CPUCounters
	// -------------------------------------------------------------
	err = client.CPUCounters()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Check(t, requestResult.Result != nil)
	// t.Log(requestResult.Result)

	// -------------------------------------------------------------
	// test if the response contains expected keys
	// -------------------------------------------------------------
	expectedKeys := []string{
		"cycle",
		"frame",
		"instruction",
	}
	for _, key := range expectedKeys {
		_, exist := (*requestResult.Result)[key]
		assert.Check(t, exist)
	}

	assert.Equal(t, requestResult.Status, 200)
}

func TestRAMClearBlock(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	const address = 0x1000
	const bytesCount = 256

	//-------------------------------------------------------------
	// test: RAMClear and RAMReadBlock
	//-------------------------------------------------------------
	value := byte(0xff)
	testClearRAMForGivenDevice(t, client, address, bytesCount, value, c64RAM)
	fetchedData := testReadMemoryForGivenDevice(t, client, address, bytesCount, c64RAM)
	testArrayHasOnlyExpectedValue(t, fetchedData, value)

	value = byte(0x00)
	testClearRAMForGivenDevice(t, client, address, bytesCount, value, c64RAM)
	fetchedData = testReadMemoryForGivenDevice(t, client, address, bytesCount, c64RAM)
	testArrayHasOnlyExpectedValue(t, fetchedData, value)
}

func TestMemoryWriteThenReadBlock(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	var address uint16
	address = 0x1000
	const bytesCount = 256

	testData := getSliceOfConsecutiveBytes(0, bytesCount)

	// -------------------------------------------------------------
	// test: RAMWriteBlock and RAMReadBlock
	// -------------------------------------------------------------
	testWriteMemoryForGivenDevice(t, client, address, testData, c64RAM)
	fetchedData := testReadMemoryForGivenDevice(t, client, address, bytesCount, c64RAM)
	assert.DeepEqual(t, testData, fetchedData)

	// -------------------------------------------------------------
	// test: CPUMemoryWriteBlock and CPUMemoryReadBlock
	// -------------------------------------------------------------
	// set $01 to $34 - makes the whole RAM visible
	client.CPUMemoryWriteBlock(0x01, []byte{0x34})
	assert.NilError(t, err)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	_, _ = assertSuccessResponse(t, msgType, msg)

	address = 0xd000
	testWriteMemoryForGivenDevice(t, client, address, testData, c64CPUMemory)
	fetchedData = testReadMemoryForGivenDevice(t, client, address, bytesCount, c64CPUMemory)
	assert.DeepEqual(t, testData, fetchedData)
}

//-------------------------------------------------------------
// TODO: this test is not working, fix it
//-------------------------------------------------------------
// func TestSegmentWriteThenReadBlock(t *testing.T) {
// 	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
// 	assert.Check(t, client != nil)
// 	defer client.Close()

// 	response, err := client.Connect()
// 	assertSuccessfullConnection(t, response, err)

// 	const bytesCount = 256

// 	testData := getSliceOfConsecutiveBytes(0, bytesCount)

// 	err = client.WriteSegment("segment1", testData)
// 	assert.NilError(t, err)

// 	// Receive response (blocks until message is received)
// 	msgType, msg, err := client.ReceiveMessage()
// 	assert.NilError(t, err)
// 	_, _ = assertSuccessResponse(t, msgType, msg)
// }

func TestVICWriteThenReadRegisters(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	type testCase struct {
		registers c64dws.RegistersMap // this is written to the VIC registers
		expected  map[byte]byte       // this is expected to be read from the VIC registers
	}

	vicTestCases := []testCase{
		{
			registers: c64dws.RegistersMap{
				"0xD020": 0x00,
				"0xD021": 0x00,
			},
			expected: map[byte]byte{
				0x20: 0xf0,
				0x21: 0xf0,
			},
		},
		{
			registers: c64dws.RegistersMap{
				"0xd020": 0x01,
				"0xd021": 0x01,
			},
			expected: map[byte]byte{
				0x20: 0xf1,
				0x21: 0xf1,
			},
		},
		{
			registers: c64dws.RegistersMap{
				"$d020": 0x02,
				"$d021": 0x02,
			},
			expected: map[byte]byte{
				0x20: 0xf2,
				0x21: 0xf2,
			},
		},
		{
			registers: c64dws.RegistersMap{
				"$d020": 0x03,
				"$21":   0x03,
			},
			expected: map[byte]byte{
				0x20: 0xf3,
				0x21: 0xf3,
			},
		},
		{
			registers: c64dws.RegistersMap{
				"53280": 0x04,
				"33":    0x04,
			},
			expected: map[byte]byte{
				0x20: 0xf4,
				0x21: 0xf4,
			},
		},
	}

	registersToRead := []uint16{0xd020, 0xd021}
	for _, testCase := range vicTestCases {
		// Write the VIC registers
		err = client.VICWrite(testCase.registers)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err := client.ReceiveMessage()
		assert.NilError(t, err)
		_, _ = assertSuccessResponse(t, msgType, msg)

		// Read the VIC registers
		err = client.VICRead(registersToRead)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err = client.ReceiveMessage()
		assert.NilError(t, err)
		requestResult, _ := assertSuccessResponse(t, msgType, msg)
		assert.Check(t, requestResult.Result != nil)

		regiters := (*requestResult.Result)["registers"]
		registersMap := convertArrayOfArraysToRegistersNBitMap[uint8](regiters.([]any))

		// test if the response contains expected values
		for k, v := range testCase.expected {
			// t.Logf("Register key: 0x%02x, has: 0x%02x, expect: 0x%02x", k, registersMap[k], v)
			assert.Equal(t, registersMap[k], v)
		}
	}
}

func TestCIAWriteThenReadRegisters(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	type testCase struct {
		ciaNum    c64dws.CIANum       // this is the CIA number
		registers c64dws.RegistersMap // this is written to the CIA registers
		expected  map[byte]byte       // this is expected to be read from the CIA registers
	}

	ciaTestCases := []testCase{
		{
			ciaNum: c64dws.CIAInfer,
			registers: c64dws.RegistersMap{
				"0xDD02": 0b00000011,
			},
			expected: map[byte]byte{
				0x02: 0b00000011,
			},
		},
		{
			ciaNum: c64dws.CIAInfer,
			registers: c64dws.RegistersMap{
				"0xdd02": 0b00000010,
			},
			expected: map[byte]byte{
				0x02: 0b00000010,
			},
		},
		{
			ciaNum: c64dws.CIAInfer,
			registers: c64dws.RegistersMap{
				"$dd02": 0b00000001,
			},
			expected: map[byte]byte{
				0x02: 0b00000001,
			},
		},
		{
			ciaNum: c64dws.CIAInfer,
			registers: c64dws.RegistersMap{
				"56578": 0b00000000,
			},
			expected: map[byte]byte{
				0x02: 0b00000000,
			},
		},
		{
			ciaNum: c64dws.CIA2,
			registers: c64dws.RegistersMap{
				"$02": 0b00000011,
			},
			expected: map[byte]byte{
				0x02: 0b00000011,
			},
		},
		{
			ciaNum: c64dws.CIA2,
			registers: c64dws.RegistersMap{
				"2": 0b00000010,
			},
			expected: map[byte]byte{
				0x02: 0b00000010,
			},
		},
	}

	registersToRead := []uint16{0xdd02}
	for _, testCase := range ciaTestCases {
		// Write the CIA registers
		err = client.CIAWrite(testCase.ciaNum, testCase.registers)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err := client.ReceiveMessage()
		assert.NilError(t, err)
		_, _ = assertSuccessResponse(t, msgType, msg)

		time.Sleep(100 * time.Millisecond)

		// Read the CIA registers
		err = client.CIARead(testCase.ciaNum, registersToRead)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err = client.ReceiveMessage()
		assert.NilError(t, err)
		requestResult, _ := assertSuccessResponse(t, msgType, msg)
		assert.Check(t, requestResult.Result != nil)

		// convert array of arrays to map
		regiters := (*requestResult.Result)["registers"]
		registersMap := convertArrayOfArraysToRegistersNBitMap[uint8](regiters.([]any))

		// test if the response contains expected values
		for k, v := range testCase.expected {
			// t.Logf("Register key: 0x%02x, has: 0x%02x, expect: 0x%02x", k, registersMap[k], v)
			assert.Equal(t, registersMap[k], v)
		}
	}
}

func TestDrive1541VIAWriteThenReadRegisters(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	type testCase struct {
		driveNum  c64dws.DriveNum     // this is the drive number
		viaNum    c64dws.VIANum       // this is the VIA number
		registers c64dws.RegistersMap // this is written to the VIA registers
		expected  map[byte]byte       // this is expected to be read from the VIA registers
	}

	viaTestCases := []testCase{
		{
			driveNum: c64dws.DriveDefault,
			viaNum:   c64dws.VIAInfer,
			registers: c64dws.RegistersMap{
				"0x1C00": 0b00001100,
			},
			expected: map[byte]byte{
				0x00: 0b00001100,
			},
		},
		{
			driveNum: c64dws.DriveDefault,
			viaNum:   c64dws.VIAInfer,
			registers: c64dws.RegistersMap{
				"0x1c00": 0b00001100,
			},
			expected: map[byte]byte{
				0x00: 0b00001100,
			},
		},
		{
			driveNum: c64dws.Drive0,
			viaNum:   c64dws.VIA2,
			registers: c64dws.RegistersMap{
				"$0": 0b00001100,
			},
			expected: map[byte]byte{
				0x00: 0b00001100,
			},
		},
		{
			driveNum: c64dws.Drive0,
			viaNum:   c64dws.VIA2,
			registers: c64dws.RegistersMap{
				"0": 0b00001100,
			},
			expected: map[byte]byte{
				0x00: 0b00001100,
			},
		},
	}

	registersToRead := []uint16{0x1c00}
	for _, testCase := range viaTestCases {
		// Write the VIA registers
		err = client.Drive1541VIAWrite(testCase.driveNum, testCase.viaNum, testCase.registers)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err := client.ReceiveMessage()
		assert.NilError(t, err)
		_, _ = assertSuccessResponse(t, msgType, msg)

		// Read the VIA registers
		err = client.Drive1541VIARead(testCase.driveNum, testCase.viaNum, registersToRead)
		assert.NilError(t, err)

		// Receive response (blocks until message is received)
		msgType, msg, err = client.ReceiveMessage()
		assert.NilError(t, err)
		requestResult, _ := assertSuccessResponse(t, msgType, msg)
		assert.Check(t, requestResult.Result != nil)

		// convert array of arrays to map
		regiters := (*requestResult.Result)["registers"]
		registersMap := convertArrayOfArraysToRegistersNBitMap[uint8](regiters.([]any))

		// test if the response contains expected values
		for k, v := range testCase.expected {
			// t.Logf("Register key: 0x%02x, has: 0x%02x, expect: 0x%02x", k, registersMap[k], v)
			assert.Equal(t, registersMap[k], v)
		}
	}
}

// -------------------------------------------------------------
// TODO: this test is not working, fix it
// -------------------------------------------------------------
// func TestSIDWriteThenReadRegisters(t *testing.T) {
// 	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
// 	assert.Check(t, client != nil)
// 	defer client.Close()

// 	response, err := client.Connect()
// 	assertSuccessfullConnection(t, response, err)

// 	type testCase struct {
// 		registers       c64dws.SIDRegistersMap // this is written to the SID registers
// 		registersToRead c64dws.Registers       // this is the SID registers to read
// 		expected        map[uint16]byte        // this is expected to be read from the SID registers
// 	}

// 	sidNum := c64dws.SIDDefault
// 	sidTestCases := []testCase{
// 		{
// 			registers: c64dws.SIDRegistersMap{
// 				"SID0": {
// 					Num: sidNum,
// 					Registers: c64dws.RegistersMap{
// 						"0xD400": 0xff,
// 					},
// 				},
// 			},
// 			expected: map[uint16]byte{
// 				0xd400: 0xff,
// 			},
// 		},
// 	}

// 	registersToRead := []uint16{0xd400}
// 	for _, testCase := range sidTestCases {
// 		// Write the SID registers
// 		err = client.SIDWrite(testCase.registers)
// 		assert.NilError(t, err)

// 		// Receive response (blocks until message is received)
// 		msgType, msg, err := client.ReceiveMessage()
// 		assert.NilError(t, err)
// 		_, _ = assertSuccessResponse(t, msgType, msg)

// 		// Read the SID registers
// 		err = client.SIDRead(sidNum, registersToRead)
// 		assert.NilError(t, err)

// 		// Receive response (blocks until message is received)
// 		msgType, msg, err = client.ReceiveMessage()
// 		assert.NilError(t, err)
// 		requestResult, _ := assertSuccessResponse(t, msgType, msg)
// 		assert.Check(t, requestResult.Result != nil)

// 		// convert array of arrays to map
// 		regiters := (*requestResult.Result)["registers"]
// 		registersMap := convertArrayOfArraysToRegistersNBitMap[uint16](regiters.([]any))

// 		// test if the response contains expected values
// 		for k, v := range testCase.expected {
// 			// t.Logf("Register key: 0x%04x, has: 0x%02x, expect: 0x%02x", k, registersMap[k], v)
// 			assert.Equal(t, registersMap[k], v)
// 		}
// 	}
// }

func TestDrive1541RAMClear(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	const address = 0x0200
	const bytesCount = 256

	//-------------------------------------------------------------
	// test: Drive1541RAMClear and Drive1541RAMReadBlock
	//-------------------------------------------------------------
	value := byte(0xff)
	testClearRAMForGivenDevice(t, client, address, bytesCount, value, drive1541RAM)
	fetchedData := testReadMemoryForGivenDevice(t, client, address, bytesCount, drive1541RAM)
	testArrayHasOnlyExpectedValue(t, fetchedData, value)

	value = byte(0x00)
	testClearRAMForGivenDevice(t, client, address, bytesCount, value, drive1541RAM)
	fetchedData = testReadMemoryForGivenDevice(t, client, address, bytesCount, drive1541RAM)
	testArrayHasOnlyExpectedValue(t, fetchedData, value)
}

func testDrive1541MemoryWriteThenReadBlock(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	const address = 0x1000
	const bytesCount = 256

	testData := getSliceOfConsecutiveBytes(0, bytesCount)

	// -------------------------------------------------------------
	// test: Drive1541RAMWriteBlock and Drive1541RAMReadBlock
	// -------------------------------------------------------------
	testWriteMemoryForGivenDevice(t, client, address, testData, drive1541RAM)
	fetchedData := testReadMemoryForGivenDevice(t, client, address, bytesCount, drive1541RAM)
	assert.DeepEqual(t, testData, fetchedData)

	// -------------------------------------------------------------
	// test: Drive1541CPUMemoryWriteBlock and Drive1541CPUMemoryReadBlock
	// -------------------------------------------------------------
	testWriteMemoryForGivenDevice(t, client, address, testData, drive1541CPUMemory)
	fetchedData = testReadMemoryForGivenDevice(t, client, address, bytesCount, drive1541CPUMemory)
	assert.DeepEqual(t, testData, fetchedData)
}

// -------------------------------------------------------------
// TODO:
// improve this test to force CPU to stop on breakpoint
// Also add capturing Server event when CPU stops on breakpoint
// -------------------------------------------------------------
func TestCPUBreakpointAddRemove(t *testing.T) {
	const address = 0x1000

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: AddCPUBreakpoint
	// -------------------------------------------------------------
	err = client.AddCPUBreakpoint(address)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: RemoveCPUBreakpoint
	// -------------------------------------------------------------
	err = client.RemoveCPUBreakpoint(address)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

// -------------------------------------------------------------
// TODO:
// improve this test to force CPU to stop on breakpoint
// Also add capturing Server event when CPU stops on breakpoint
// -------------------------------------------------------------
func TestCPUMemoryBreakpointAddRemove(t *testing.T) {
	const address = 0x1000
	const value byte = 0x80

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: AddCPUMemoryBreakpoint
	// -------------------------------------------------------------
	err = client.AddCPUMemoryBreakpoint(address, value, c64dws.MemoryBreakpointAccessWrite, "<")
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: RemoveCPUMemoryBreakpoint
	// -------------------------------------------------------------
	err = client.RemoveCPUMemoryBreakpoint(address, value)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestVICAddRemoveRasterBreakpoint(t *testing.T) {
	const rasterLine = 100

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: AddRasterBreakpoint
	// -------------------------------------------------------------
	err = client.AddRasterBreakpoint(rasterLine)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: RemoveRasterBreakpoint
	// -------------------------------------------------------------
	err = client.RemoveRasterBreakpoint(rasterLine)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestStepCycle(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: StepCycle
	// -------------------------------------------------------------
	err = client.StepCycle()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestStepInstruction(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: StepInstruction
	// -------------------------------------------------------------
	err = client.StepInstruction()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestStepSubroutine(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: StepSubroutine
	// -------------------------------------------------------------
	err = client.StepSubroutine()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestInputJoystick(t *testing.T) {
	const joystickPort = 0
	const joystickAxis = c64dws.JoystickAxisN

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: InputJoystickUp
	// -------------------------------------------------------------
	err = client.InputJoystickUp(joystickAxis, joystickPort)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: InputJoystickDown
	// -------------------------------------------------------------
	err = client.InputJoystickDown(joystickAxis, joystickPort)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestInputKey(t *testing.T) {
	const key = 0x20 // space key

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// test: InputKeyDown
	// -------------------------------------------------------------
	err = client.InputKeyDown(key)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: InputKeyUp
	// -------------------------------------------------------------
	err = client.InputKeyUp(key)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	assert.NilError(t, err)
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)
}

func TestSavePRGFile(t *testing.T) {
	const PRGFileRelPath = "./output/SavePRG-Test.prg"
	const bytesCount = 256
	const jumpAddress uint16 = 0x1000
	const startAddress uint16 = 0x1000
	const endAddress uint16 = 0x1100

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	// -------------------------------------------------------------
	// Prepare the test data in RAM
	// -------------------------------------------------------------
	testData := getSliceOfConsecutiveBytes(0, bytesCount)
	testWriteMemoryForGivenDevice(t, client, startAddress, testData, c64RAM)
	fetchedData := testReadMemoryForGivenDevice(t, client, startAddress, bytesCount, c64RAM)
	assert.DeepEqual(t, testData, fetchedData)

	// -------------------------------------------------------------
	// Prepare the PRG file path and remove file if it exists
	// -------------------------------------------------------------
	PRGFileAbsPath, err := filepath.Abs(PRGFileRelPath)
	assert.NilError(t, err)

	if _, err := os.Stat(PRGFileAbsPath); err == nil {
		err = os.Remove(PRGFileAbsPath)
		assert.NilError(t, err)
	}

	// -------------------------------------------------------------
	// test: SavePRG
	// -------------------------------------------------------------
	err = client.SavePRG(PRGFileAbsPath, startAddress, endAddress, false, jumpAddress)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// Read the saved file
	data, err := os.ReadFile(PRGFileAbsPath)
	assert.NilError(t, err)

	// Check if the saved file has the correct start address
	loadAddress := uint16(data[1])<<8 | uint16(data[0])
	assert.Equal(t, loadAddress, jumpAddress)

	// Check if the saved file contains the expected data
	fileData := data[2:]
	assert.DeepEqual(t, fileData, testData)
}

// -------------------------------------------------------------
// TODO:
// improve this test to confirm that emulation is paused/continued
// -------------------------------------------------------------
func TestPauseContinueEmulation(t *testing.T) {
	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)
	defer client.Close()

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	hardOrSoftReset(t, client, softReset)

	// -------------------------------------------------------------
	// test: PauseEmulation
	// -------------------------------------------------------------
	err = client.PauseEmulation()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// -------------------------------------------------------------
	// test: ContinueEmulation
	// -------------------------------------------------------------
	err = client.ContinueEmulation()
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err = client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ = assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	hardOrSoftReset(t, client, softReset)
}

func TestLoadAndWarpMode(t *testing.T) {
	const AmanitaSamarURL = "https://csdb.dk/release/download.php?id=269202"
	const AmanitaSamarZipPath = "./output/Amanita-Samar.zip"
	const AmanitaSamarDiskA = "Amanita_by_Samar_Disk_A.d64"

	client := c64dws.NewDefaultClient(c64dws.EmulatorC64, c64dws.StreamAPI)
	assert.Check(t, client != nil)

	response, err := client.Connect()
	assertSuccessfullConnection(t, response, err)

	hardOrSoftReset(t, client, softReset)

	// Enable warp mode
	client.SetWarpMode(true)

	// Download Amanita-Samar demo
	err = downloadFile(AmanitaSamarURL, AmanitaSamarZipPath)
	assert.NilError(t, err)

	// Unzip the demo
	amanitaPath, err := unzipToTempDir(AmanitaSamarZipPath)
	assert.NilError(t, err)

	// -------------------------------------------------------------
	// test: LoadFile
	// -------------------------------------------------------------
	amanitaD64Path := fmt.Sprintf("%s/%s", amanitaPath, AmanitaSamarDiskA)
	err = client.LoadFile(amanitaD64Path)
	assert.NilError(t, err)

	// Receive response (blocks until message is received)
	msgType, msg, err := client.ReceiveMessage()
	if err != nil {
		log.Fatal(err)
	}
	requestResult, _ := assertSuccessResponse(t, msgType, msg)
	assert.Equal(t, requestResult.Status, 200)

	// Wait for 10 seconds to let the demo run
	time.Sleep(10 * time.Second)

	// Disable warp mode
	client.SetWarpMode(false)

	// Reset the emulator
	hardOrSoftReset(t, client, softReset)
}
