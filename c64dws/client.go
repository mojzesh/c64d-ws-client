// # Retro Debugger WebSocket API client
//
// This package makes it easy to interact with the Retro Debugger via WebSocket API.
// It works in async and sync mode.
// It allows you to manipulate the state of the following components of the Retro Debugger:
// This will be rendered as a list:
//   - Base hardware: CPU, RAM, Memory Segments
//   - Chipsets: VIC, CIA, SID
//   - 1541 Drive: CPU / RAM / VIA
//   - Peripherals: Joystick, Keyboard
//   - Breakpoints: Raster and CPU
package c64dws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Default host, port and scheme of Retro Debugger WebSocket API
// Default URL: ws://localhost:3563
const WS_SCHEME = "ws"
const WS_HOST = "localhost"
const WS_PORT_0x0DEB = 3563

const WS_DEFAULT_TOKEN_FORMAT = "id-%d"

// API types
type APIType string

const (
	StreamAPI APIType = "/stream"
)

// API function paths
type APIFn string

const (
	APIFnLoad                         APIFn = "load"
	APIFnDetachEverything             APIFn = "%s/detachEverything"
	APIFnContinue                     APIFn = "%s/continue"
	APIFnPause                        APIFn = "%s/pause"
	APIFnCPUBreakpointAdd             APIFn = "%s/cpu/breakpoint/add"
	APIFnCPUBreakpointRemove          APIFn = "%s/cpu/breakpoint/remove"
	APIFnCPUCountersRead              APIFn = "%s/cpu/counters/read"
	APIFnCPUMakeJmp                   APIFn = "%s/cpu/makejmp"
	APIFnCPUMemoryBreakpointAdd       APIFn = "%s/cpu/memory/breakpoint/add"
	APIFnCPUMemoryBreakpointRemove    APIFn = "%s/cpu/memory/breakpoint/remove"
	APIFnCPUMemoryReadBlock           APIFn = "%s/cpu/memory/readBlock"
	APIFnCPUMemoryWriteBlock          APIFn = "%s/cpu/memory/writeBlock"
	APIFnCPUStatus                    APIFn = "%s/cpu/status"
	APIFnDrive1541CPUMemoryReadBlock  APIFn = "%s/drive1541/cpu/memory/readBlock"
	APIFnDrive1541CPUMemoryWriteBlock APIFn = "%s/drive1541/cpu/memory/writeBlock"
	APIFnDrive1541RAMClear            APIFn = "%s/drive1541/ram/clear"
	APIFnDrive1541RAMReadBlock        APIFn = "%s/drive1541/ram/readBlock"
	APIFnDrive1541RAMWriteBlock       APIFn = "%s/drive1541/ram/writeBlock"
	APIFnDrive1541VIARead             APIFn = "%s/drive1541/via/read"
	APIFnDrive1541VIAWrite            APIFn = "%s/drive1541/via/write"
	APIFnInputJoystickDown            APIFn = "%s/input/joystick/down"
	APIFnInputJoystickUp              APIFn = "%s/input/joystick/up"
	APIFnInputKeyDown                 APIFn = "%s/input/key/down"
	APIFnInputKeyUp                   APIFn = "%s/input/key/up"
	APIFnRAMClear                     APIFn = "%s/ram/clear"
	APIFnRAMReadBlock                 APIFn = "%s/ram/readBlock"
	APIFnRAMWriteBlock                APIFn = "%s/ram/writeBlock"
	APIFnResetHard                    APIFn = "%s/reset/hard"
	APIFnResetSoft                    APIFn = "%s/reset/soft"
	APIFnSavePRG                      APIFn = "%s/savePrg"
	APIFnSegmentRead                  APIFn = "%s/segment/read"
	APIFnSegmentWrite                 APIFn = "%s/segment/write"
	APIFnStepCycle                    APIFn = "%s/step/cycle"
	APIFnStepInstruction              APIFn = "%s/step/instruction"
	APIFnStepSubroutine               APIFn = "%s/step/subroutine"
	APIFnCIARead                      APIFn = "%s/cia/read"
	APIFnCIAWrite                     APIFn = "%s/cia/write"
	APIFnSIDRead                      APIFn = "%s/sid/read"
	APIFnSIDWrite                     APIFn = "%s/sid/write"
	APIFnVICAddRasterBreakpoint       APIFn = "%s/vic/breakpoint/add"
	APIFnVICRemoveRasterBreakpoint    APIFn = "%s/vic/breakpoint/remove"
	APIFnVICRead                      APIFn = "%s/vic/read"
	APIFnVICWrite                     APIFn = "%s/vic/write"
	APIFnWarpSet                      APIFn = "%s/warp/set"
)

// Emulator types
type EmulatorType string

const (
	EmulatorC64      EmulatorType = "c64"
	EmulatorATARI800 EmulatorType = "atari800"
	EmulatorNES      EmulatorType = "nes"
)

// Token type
type TokenType string

const (
	TokenTypeAutoIncrement TokenType = "autoincrement"
	TokenTypeUUID          TokenType = "uuid"
)

// C64D Debugger message types
type C64DMessageType string

const (
	C64DRequestResponse C64DMessageType = "RequestResponse"
	C64DServerEvent     C64DMessageType = "ServerEvent"
	C64DUnknown         C64DMessageType = "Unknown"
)

// Message types (follows gorilla/websocket)
type WSMessageType int

const (
	WSMessageTypeText   WSMessageType = websocket.TextMessage
	WSMessageTypeBinary WSMessageType = websocket.BinaryMessage
	WSMessageTypeClose  WSMessageType = websocket.CloseMessage
	WSMessageTypePing   WSMessageType = websocket.PingMessage
	WSMessageTypePong   WSMessageType = websocket.PongMessage
)

// WebSocket message type to string
func (mt WSMessageType) String() string {
	switch mt {
	case WSMessageTypeText:
		return "Text"
	case WSMessageTypeBinary:
		return "Binary"
	case WSMessageTypeClose:
		return "Close"
	case WSMessageTypePing:
		return "Ping"
	case WSMessageTypePong:
		return "Pong"
	default:
		return "Unknown"
	}
}

// Host description
type HostDesc struct {
	hostName string
	port     int
	scheme   string
}

// Get the default host
func GetDefaultHost() HostDesc {
	return HostDesc{
		hostName: WS_HOST,
		port:     WS_PORT_0x0DEB,
		scheme:   WS_SCHEME,
	}
}

func GetCustomHost(hostName string, port int, scheme string) (HostDesc, error) {
	if scheme == "" || port == 0 || hostName == "" {
		return HostDesc{}, errors.New("Invalid host description")
	}
	if scheme != "ws" {
		return HostDesc{}, errors.New("Currently only 'ws' scheme is supported")
	}
	return HostDesc{
		hostName: hostName,
		port:     port,
		scheme:   scheme,
	}, nil
}

// C64D WebSocket Client
type Client struct {
	emulator      EmulatorType
	host          HostDesc
	apiType       APIType
	conn          *websocket.Conn
	tokenType     TokenType
	tokenFormat   string
	autoincrement int64 // autoincrement token used when TokenTypeAutoIncrement
}

// Create a new client with custom host, port and scheme
func NewCustomClient(emulator EmulatorType, apiType APIType, tokenType TokenType, tokenFormat string, host HostDesc) *Client {
	return &Client{
		emulator:      emulator,
		apiType:       apiType,
		host:          host,
		tokenType:     tokenType,
		tokenFormat:   tokenFormat,
		autoincrement: 0,
	}
}

// Create a new client with default host, port and scheme
func NewDefaultClient(emulator EmulatorType, apiType APIType) *Client {
	return NewCustomClient(
		emulator,
		apiType,
		TokenTypeAutoIncrement,
		WS_DEFAULT_TOKEN_FORMAT,
		GetDefaultHost(),
	)
}

// Close the connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Get the URL to the Retro Debugger WebSocket API endpoint
func (c *Client) GetURL() string {
	url := url.URL{
		Scheme: c.host.scheme,
		Host:   c.GetAPIAddr(c.host.hostName, c.host.port),
		Path:   c.GetAPIEndpoint(c.apiType),
	}

	return url.String()
}

// Connect to the Retro Debugger WebSocket API
func (c *Client) Connect() ([]byte, error) {
	conn, resp, err := websocket.DefaultDialer.Dial(c.GetURL(), nil)
	if err != nil {
		return nil, errors.Join(errors.New("Dial error"), err)
	}

	// get the response
	responseBody := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, responseBody)
	if err != nil {
		return nil, errors.Join(errors.New("Read response error"), err)
	}

	c.conn = conn

	return responseBody, nil
}

// Get the API address
func (c *Client) GetAPIAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// Get the API endpoint
func (c *Client) GetAPIEndpoint(api APIType) string {
	return string(api)
}

// Get the API function path
func (c *Client) GetAPIFn(apiPathFmt APIFn) string {
	if strings.Contains(string(apiPathFmt), "%s") {
		return fmt.Sprintf(string(apiPathFmt), c.emulator)
	}
	return string(apiPathFmt)
}

// Read raw message from the WebSocket connection
func (c *Client) getRAWMessage() (WSMessageType, []byte, []byte, error) {
	var textPart []byte
	var binaryPart []byte

	messageType, message, err := c.conn.ReadMessage()
	if err != nil {
		return WSMessageType(messageType), nil, nil, err
	}

	switch messageType {
	case websocket.TextMessage:
		// Only text message is expected
		return WSMessageType(messageType), message, nil, nil
	case websocket.BinaryMessage:
		// Mix of text and binary data is expected
		endOfTextPartOffset := bytes.IndexByte(message, 0)
		if endOfTextPartOffset == -1 {
			// Message has only text part
			textPart = message
		} else {
			// Message has text part and binary part
			textPart = message[:endOfTextPartOffset]
			if endOfTextPartOffset < len(message)-1 {
				binaryPart = message[endOfTextPartOffset+1:]
			}
		}
	}

	return WSMessageType(messageType), textPart, binaryPart, nil
}

// Receive message from the WebSocket connection
func (c *Client) ReceiveMessage() (C64DMessageType, any, error) {
	msgType, textPart, binaryPart, err := c.getRAWMessage()
	if err != nil {
		return C64DUnknown, nil, err
	}
	// fmt.Printf("Type: %s, Text: %s, Bin: %v\n", msgType.String(), textPart, binaryPart)

	switch msgType {
	//----------------------------------------------
	// Server events always send as a text message
	//----------------------------------------------
	case websocket.TextMessage:
		// Base Server Event message
		var serverEvent ServerEventBase
		err := json.Unmarshal(textPart, &serverEvent)
		if err != nil {
			return C64DUnknown, nil, err
		}
		switch serverEvent.Event {
		case string(ServerEventTypeBreakpoint):
			// Breakpoint Event message
			var breakpointEvent BreakpointEventBase
			err := json.Unmarshal(textPart, &breakpointEvent)
			if err != nil {
				return C64DServerEvent, nil, err
			}
			switch breakpointEvent.Type {
			case string(BreakpointEventRaster):
				// Raster Breakpoint Event message
				var rasterBreakpointEvent RasterBreakpointEvent
				err := json.Unmarshal(textPart, &rasterBreakpointEvent)
				if err != nil {
					return C64DServerEvent, nil, err
				}
				return C64DServerEvent, rasterBreakpointEvent, nil
			case string(BreakpointEventCPUData):
				// CPU Data Breakpoint Event message
				var cpuAddrBreakpointEvent CPUAddrBreakpointEvent
				err := json.Unmarshal(textPart, &cpuAddrBreakpointEvent)
				if err != nil {
					return C64DServerEvent, nil, err
				}
				return C64DServerEvent, cpuAddrBreakpointEvent, nil
			case string(BreakpointEventCPUAddr):
				// CPU Address Breakpoint Event message
				var cpuDataBreakpointEvent CPUDataBreakpointEvent
				err := json.Unmarshal(textPart, &cpuDataBreakpointEvent)
				if err != nil {
					return C64DServerEvent, nil, err
				}
				return C64DServerEvent, cpuDataBreakpointEvent, nil
			default:
				return C64DServerEvent, breakpointEvent, nil
			}
		}
	//----------------------------------------------
	// RequestResponse messages are always sent as binary message
	//----------------------------------------------
	case websocket.BinaryMessage:
		// Response Payload message
		var requestResultBase RequestResultBase
		err := json.Unmarshal(textPart, &requestResultBase)
		if err != nil {
			return C64DRequestResponse, nil, err
		}

		if requestResultBase.Status == 200 {
			// Success Response Payload message
			var requestResult RequestResult
			err := json.Unmarshal(textPart, &requestResult)
			if err != nil {
				return C64DRequestResponse, nil, err
			}
			requestResult.BinaryData = binaryPart
			return C64DRequestResponse, requestResult, nil
		} else {
			// Error Response Payload message
			var requestError RequestError
			err := json.Unmarshal(textPart, &requestError)
			if err != nil {
				return C64DRequestResponse, nil, err
			}
			return C64DRequestResponse, requestError, nil
		}
	//----------------------------------------------
	// Unknown message type
	//----------------------------------------------
	default:
		return C64DUnknown, nil, errors.New("Unknown message type")
	}

	return C64DUnknown, nil, nil
}

// Prepare message
func (c *Client) prepareMessage(apiFn APIFn, params *Params, binaryData []byte, token string) ([]byte, error) {
	var requestPayload any

	requestPayloadBase := RequestPayloadBase{
		Fn:     c.GetAPIFn(apiFn),
		Params: params,
	}

	if token == "" {
		requestPayload = requestPayloadBase
	} else {
		requestPayload = RequestPayloadWithToken{
			RequestPayloadBase: requestPayloadBase,
			Token:              token,
		}
	}

	requestPayloadBytes, err := json.Marshal(&requestPayload)
	if err != nil {
		return nil, err
	}

	if binaryData != nil {
		requestPayloadBytes = append(requestPayloadBytes, 0)
		requestPayloadBytes = append(requestPayloadBytes, binaryData...)
	}

	return requestPayloadBytes, nil
}

// Send message over the WebSocket connection
func (c *Client) sendMessage(message []byte) error {
	err := c.conn.WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to prepare and send message
func (c *Client) prepareAndSendMessage(apiFn APIFn, params *Params, binaryData []byte, token string) error {
	requestPayloadBytes, err := c.prepareMessage(apiFn, params, binaryData, token)
	if err != nil {
		return err
	}

	return c.sendMessage(requestPayloadBytes)
}

// Load file
func (c *Client) LoadFile(path string, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnLoad,
		&Params{
			"path": path,
		},
		nil,
		c.extractToken(token),
	)
}

// Save PRG
func (c *Client) SavePRG(path string, fromAddr uint16, toAddr uint16, exomizer bool, jmpAddr uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnSavePRG,
		&Params{
			"path":     path,
			"fromAddr": fromAddr,
			"toAddr":   toAddr,
			"exomizer": exomizer,
			"jmpAddr":  jmpAddr,
		},
		nil,
		c.extractToken(token),
	)
}

// Hard reset
func (c *Client) HardReset(token ...string) error {
	return c.prepareAndSendMessage(APIFnResetHard, nil, nil, c.extractToken(token))
}

// Soft reset
func (c *Client) SoftReset(token ...string) error {
	return c.prepareAndSendMessage(APIFnResetSoft, nil, nil, c.extractToken(token))
}

// Detach everything
func (c *Client) DetachEverything(token ...string) error {
	return c.prepareAndSendMessage(APIFnDetachEverything, nil, nil, c.extractToken(token))
}

// Pause emulation
func (c *Client) PauseEmulation(token ...string) error {
	return c.prepareAndSendMessage(APIFnPause, nil, nil, c.extractToken(token))
}

// Continue emulation
func (c *Client) ContinueEmulation(token ...string) error {
	return c.prepareAndSendMessage(APIFnContinue, nil, nil, c.extractToken(token))
}

// Set warp mode
func (c *Client) SetWarpMode(warpMode bool, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnWarpSet,
		&Params{
			"warp": warpMode,
		},
		nil,
		c.extractToken(token),
	)
}

// CPU status
func (c *Client) CPUStatus(token ...string) error {
	return c.prepareAndSendMessage(APIFnCPUStatus, nil, nil, c.extractToken(token))
}

// CPU counters
func (c *Client) CPUCounters(token ...string) error {
	return c.prepareAndSendMessage(APIFnCPUCountersRead, nil, nil, c.extractToken(token))
}

// Make CPU JMP
func (c *Client) CPUMakeJMP(address uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUMakeJmp,
		&Params{
			"address": address,
		},
		nil,
		c.extractToken(token),
	)
}

// CPU Memory Write Block
func (c *Client) CPUMemoryWriteBlock(address uint16, binaryData []byte, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUMemoryWriteBlock,
		&Params{
			"address": address,
		},
		binaryData,
		c.extractToken(token),
	)
}

// CPU Memory Read Block
func (c *Client) CPUMemoryReadBlock(address uint16, size uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUMemoryReadBlock,
		&Params{
			"address": address,
			"size":    size,
		},
		nil,
		c.extractToken(token),
	)
}

// Clear RAM
func (c *Client) RAMClear(address uint16, size uint16, value uint8, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnRAMClear,
		&Params{
			"address": address,
			"size":    size,
			"value":   value,
		},
		nil,
		c.extractToken(token),
	)
}

// Read RAM block
func (c *Client) RAMWriteBlock(address uint16, binaryData []byte, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnRAMWriteBlock,
		&Params{
			"address": address,
		},
		binaryData,
		c.extractToken(token),
	)
}

// Read RAM block
func (c *Client) RAMReadBlock(address uint16, size uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnRAMReadBlock,
		&Params{
			"address": address,
			"size":    size,
		},
		nil,
		c.extractToken(token),
	)
}

type Registers []uint16
type RegistersMap map[string]uint8

// CIA Number - CIA1, CIA2 or CIAInfer (infer from 16-bit address)
type CIANum int

const (
	CIAInfer CIANum = -1
	CIA1     CIANum = 0
	CIA2     CIANum = 1
)

// Read CIA Registers
func (c *Client) CIARead(ciaNum CIANum, registers Registers, token ...string) error {
	params := Params{
		"registers": registers,
	}

	if ciaNum != CIAInfer {
		params["num"] = ciaNum
	}

	return c.prepareAndSendMessage(
		APIFnCIARead,
		&params,
		nil,
		c.extractToken(token),
	)
}

// Write CIA Registers
func (c *Client) CIAWrite(ciaNum CIANum, registersMap RegistersMap, token ...string) error {
	params := Params{
		"registers": registersMap,
	}

	if ciaNum != CIAInfer {
		params["num"] = ciaNum
	}

	return c.prepareAndSendMessage(
		APIFnCIAWrite,
		&params,
		nil,
		c.extractToken(token),
	)
}

// Read VIC Registers
func (c *Client) VICRead(registers Registers, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnVICRead,
		&Params{
			"registers": registers,
		},
		nil,
		c.extractToken(token),
	)
}

// Write VIC Registers
func (c *Client) VICWrite(registersMap RegistersMap, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnVICWrite,
		&Params{
			"registers": registersMap,
		},
		nil,
		c.extractToken(token),
	)
}

// SID Registers and SID Registers Map
type SIDRegistersMap map[string]SIDRegisters
type SIDRegisters struct {
	Num       SIDNum       `json:"num"`       // SID number
	Registers RegistersMap `json:"registers"` // SID registers map
}

// SID Number - SID0, SID1, SID2, SID3 or SIDDefault (server defaults to SID0)
type SIDNum int

const (
	SIDDefault SIDNum = -1
	SID0       SIDNum = 0
	SID1       SIDNum = 1
	SID2       SIDNum = 2
	SID3       SIDNum = 3
)

// Read SID Registers
func (c *Client) SIDRead(sidNum SIDNum, registers Registers, token ...string) error {
	params := Params{
		"registers": registers,
	}

	if sidNum != SIDDefault {
		params["num"] = sidNum
	}

	return c.prepareAndSendMessage(
		APIFnSIDRead,
		&params,
		nil,
		c.extractToken(token),
	)
}

// Write SID Registers
func (c *Client) SIDWrite(registersPerSid SIDRegistersMap, token ...string) error {
	// Make sure SID0 is used if SIDDefault is used
	for key, sidRegisters := range registersPerSid {
		if sidRegisters.Num == SIDDefault {
			currentSIDRegs := registersPerSid[key]
			currentSIDRegs.Num = SID0
			registersPerSid[key] = currentSIDRegs
		}
	}

	return c.prepareAndSendMessage(
		APIFnSIDWrite,
		&Params{
			"sids": registersPerSid,
		},
		nil,
		c.extractToken(token),
	)
}

// Read Segment
func (c *Client) ReadSegment(segment string, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnSegmentRead,
		&Params{
			"segment": segment,
		},
		nil,
		c.extractToken(token),
	)
}

// Write Segment
func (c *Client) WriteSegment(segment string, binaryData []byte, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnSegmentWrite,
		&Params{
			"segment": segment,
		},
		binaryData,
		c.extractToken(token),
	)
}

// Input Joystick Up
func (c *Client) InputJoystickUp(axis JoystickAxisType, port int, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnInputJoystickUp,
		&Params{
			"axis": axis,
			"port": port,
		},
		nil,
		c.extractToken(token),
	)
}

// Input Joystick Down
func (c *Client) InputJoystickDown(axis JoystickAxisType, port int, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnInputJoystickDown,
		&Params{
			"axis": axis,
			"port": port,
		},
		nil,
		c.extractToken(token),
	)
}

// Input Key Up
func (c *Client) InputKeyUp(keyCode int, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnInputKeyUp,
		&Params{
			"keyCode": keyCode,
		},
		nil,
		c.extractToken(token),
	)
}

// Input Key Down
func (c *Client) InputKeyDown(keyCode int, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnInputKeyDown,
		&Params{
			"keyCode": keyCode,
		},
		nil,
		c.extractToken(token),
	)
}

// Drive 1541 CPU Memory Read Block
func (c *Client) Drive1541CPUMemoryReadBlock(address uint16, size uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnDrive1541CPUMemoryReadBlock,
		&Params{
			"address": address,
			"size":    size,
		},
		nil,
		c.extractToken(token),
	)
}

// Drive 1541 CPU Memory Write Block
func (c *Client) Drive1541CPUMemoryWriteBlock(address uint16, binaryData []byte, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnDrive1541CPUMemoryWriteBlock,
		&Params{
			"address": address,
		},
		binaryData,
		c.extractToken(token),
	)
}

// Drive 1541 RAM Clear
func (c *Client) Drive1541RAMClear(address uint16, size uint16, value uint8, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnDrive1541RAMClear,
		&Params{
			"address": address,
			"size":    size,
			"value":   value,
		},
		nil,
		c.extractToken(token),
	)
}

// Drive 1541 RAM Read Block
func (c *Client) Drive1541RAMReadBlock(address uint16, size uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnDrive1541RAMReadBlock,
		&Params{
			"address": address,
			"size":    size,
		},
		nil,
		c.extractToken(token),
	)
}

// Drive 1541 RAM Write Block
func (c *Client) Drive1541RAMWriteBlock(address uint16, binaryData []byte, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnDrive1541RAMWriteBlock,
		&Params{
			"address": address,
		},
		binaryData,
		c.extractToken(token),
	)
}

// VIA Number - VIA1, VIA2 or VIAInfer (infer from 16-bit address)
type VIANum int

const (
	VIAInfer VIANum = -1
	VIA1     VIANum = 0
	VIA2     VIANum = 1
)

type DriveNum int

const (
	DriveDefault DriveNum = -1
	Drive0       DriveNum = 0 // Device 8
	Drive1       DriveNum = 1 // Device 9
	Drive2       DriveNum = 2 // Device 10
	Drive3       DriveNum = 3 // Device 11
)

// Drive 1541 VIA Read
func (c *Client) Drive1541VIARead(driveNum DriveNum, viaNum VIANum, registers Registers, token ...string) error {
	params := Params{
		"registers": registers,
	}

	if viaNum != VIAInfer {
		params["num"] = viaNum
	}

	if driveNum != DriveDefault {
		params["drive"] = driveNum
	}

	return c.prepareAndSendMessage(
		APIFnDrive1541VIARead,
		&params,
		nil,
		c.extractToken(token),
	)
}

// Drive 1541 VIA Write
func (c *Client) Drive1541VIAWrite(driveNum DriveNum, viaNum VIANum, registersMap RegistersMap, token ...string) error {
	params := Params{
		"registers": registersMap,
	}

	if viaNum != VIAInfer {
		params["num"] = viaNum
	}

	if driveNum != DriveDefault {
		params["drive"] = driveNum
	}

	return c.prepareAndSendMessage(
		APIFnDrive1541VIAWrite,
		&params,
		nil,
		c.extractToken(token),
	)
}

// Step cycle
func (c *Client) StepCycle(token ...string) error {
	return c.prepareAndSendMessage(APIFnStepCycle, nil, nil, c.extractToken(token))
}

// Step instruction
func (c *Client) StepInstruction(token ...string) error {
	return c.prepareAndSendMessage(APIFnStepInstruction, nil, nil, c.extractToken(token))
}

// Step subroutine
func (c *Client) StepSubroutine(token ...string) error {
	return c.prepareAndSendMessage(APIFnStepSubroutine, nil, nil, c.extractToken(token))
}

// Add CPU breakpoint
func (c *Client) AddCPUBreakpoint(address uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUBreakpointAdd,
		&Params{
			"addr": address,
		},
		nil,
		c.extractToken(token),
	)
}

// Remove CPU breakpoint
func (c *Client) RemoveCPUBreakpoint(address uint16, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUBreakpointRemove,
		&Params{
			"addr": address,
		},
		nil,
		c.extractToken(token),
	)
}

// Add CPU memory breakpoint
func (c *Client) AddCPUMemoryBreakpoint(address uint16, value uint8, access MemoryBreakpointAccess, comparison string, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUMemoryBreakpointAdd,
		&Params{
			"addr":       address,
			"value":      value,
			"access":     access,
			"comparison": comparison,
		},
		nil,
		c.extractToken(token),
	)
}

// Remove CPU memory breakpoint
func (c *Client) RemoveCPUMemoryBreakpoint(address uint16, value uint8, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnCPUMemoryBreakpointRemove,
		&Params{
			"addr": address,
		},
		nil,
		c.extractToken(token),
	)
}

// Add raster breakpoint
func (c *Client) AddRasterBreakpoint(rasterline uint8, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnVICAddRasterBreakpoint,
		&Params{
			"rasterLine": rasterline,
		},
		nil,
		c.extractToken(token),
	)
}

// Remove raster breakpoint
func (c *Client) RemoveRasterBreakpoint(rasterline uint8, token ...string) error {
	return c.prepareAndSendMessage(
		APIFnVICRemoveRasterBreakpoint,
		&Params{
			"rasterLine": rasterline,
		},
		nil,
		c.extractToken(token),
	)
}

// Get the token, allows to use custom token format per request
func (c *Client) GetToken(tokenFormat ...string) string {
	var tokenFormatStr string
	if len(tokenFormat) > 0 {
		tokenFormatStr = tokenFormat[0]
	} else {
		tokenFormatStr = c.tokenFormat
	}
	switch c.tokenType {
	case TokenTypeUUID:
		return fmt.Sprintf(tokenFormatStr, uuid.New().String())
	case TokenTypeAutoIncrement:
		c.autoincrement++
		return fmt.Sprintf(tokenFormatStr, c.autoincrement)
	default:
		panic("Unknown token type")
	}
}

// Extract token from the token slice
func (c *Client) extractToken(token []string) string {
	tokenLen := len(token)
	if tokenLen == 0 {
		return ""
	} else if len(token) == 1 {
		return token[0]
	}

	panic("Token must be a single string")
}
