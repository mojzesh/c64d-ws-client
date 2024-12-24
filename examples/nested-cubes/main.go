package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mojzesh/c64d-ws-client/c64dws"

	"github.com/mojzesh/nested-cubes/cia"
	"github.com/mojzesh/nested-cubes/effects"
	"github.com/mojzesh/nested-cubes/vc64"
	"github.com/mojzesh/nested-cubes/vic"
	"github.com/mojzesh/nested-cubes/vic/colors"
)

const waitAfterReset = true                               // set to true for waiting after hard reset
const waitTimeAfterResetInSeconds = time.Millisecond * 20 // time to wait after hard reset

const silentMode = true          // set to true for less verbose output
const warpMode = false           // set to true to enable warp mode
const rasterLineBreakpoint = 255 // raster line to wait for VSync

// Color constants
const color00 uint8 = colors.COL_BLACK_IDX
const color01 uint8 = colors.COL_DARK_BLUE_IDX
const color10 uint8 = colors.COL_LIGHT_BLUE_IDX
const color11 uint8 = colors.COL_CYAN_IDX

func main() {
	// -------------------------------------------------------------
	// Create cancellable context
	// -------------------------------------------------------------
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	// -------------------------------------------------------------
	// Connect to C64D WebSocket Server
	// -------------------------------------------------------------
	client, err := connectToWS(c64dws.EmulatorC64, c64dws.StreamAPI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	vsyncChannel := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go listenForMessages(ctx, &wg, client, vsyncChannel)

	// -------------------------------------------------------------
	// Request Hard reset
	// -------------------------------------------------------------
	if err := resetEmulator(client, waitAfterReset); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Set raster breakpoint
	// -------------------------------------------------------------
	if err = client.AddRasterBreakpoint(rasterLineBreakpoint, client.GetToken("breakpoint-%d")); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Init register $01 and put CPU into infinite loop
	// -------------------------------------------------------------
	if err := putCPUIntoInfiniteLoop(client, vsyncChannel); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Wait one more frame to make sure CPU is in infinite loop
	// -------------------------------------------------------------
	<-vsyncChannel
	if err := client.ContinueEmulation(); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Init VIC registers
	// -------------------------------------------------------------
	log.Println("Init VIC registers")
	if err := client.VICWrite(map[string]uint8{
		"0xD011": vic.D011_BITMAP_MODE_VALUE | vic.D011_SCREEN_HEIGHT_25_ROWS_VALUE | vic.D011_SCREEN_ON_VALUE | 0b00000011,
		"0xD016": 0b11000000 | vic.D016_SCREEN_MULTI_VALUE | vic.D016_SCREEN_WIDTH_40_COLS_VALUE,
		"0xD018": vic.D018_BITMAP_MODE_2000_3FFF_VALUE | vic.D018_SCREEN_MEMORY_0400_07FF_VALUE,
		"0xD020": colors.COL_DARK_BLUE_IDX,
		"0xD021": colors.COL_BLACK_IDX,
	}); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Init CIA registers
	// -------------------------------------------------------------
	log.Println("Init CIA registers")
	if err := client.CIAWrite(
		c64dws.CIAInfer,
		map[string]uint8{
			"0xDD00": cia.DD00_BANK_0_VALUE,
		},
	); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Init Color RAM, Screen RAM and Bitmap
	// -------------------------------------------------------------
	log.Println("Init Color RAM, Screen RAM and Bitmap")
	const colorScreenSize = 40 * 25
	const bitmapSize = 40 * 200
	const screenAddr uint16 = 0x0400
	const bitmapAddr uint16 = 0x2000
	const colorRAM uint16 = 0xd800
	const frameBufferAddr uint16 = 0x8000

	// Set color RAM
	vicVideoRAM := make([]byte, colorScreenSize)
	for i := 0; i < colorScreenSize; i++ {
		vicVideoRAM[i] = color11
	}
	if err = client.CPUMemoryWriteBlock(colorRAM, vicVideoRAM); err != nil {
		log.Fatal(err)
	}

	// Clear screen
	if err = client.RAMClear(screenAddr, colorScreenSize, color01<<4|color10); err != nil {
		log.Fatal(err)
	}

	// Clear bitmap
	if err = client.RAMClear(bitmapAddr, bitmapSize, 0); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Set warp mode if enabled
	// -------------------------------------------------------------
	if err = client.SetWarpMode(warpMode); err != nil {
		log.Fatal(err)
	}

	// -------------------------------------------------------------
	// Handle SIGINT and SIGTERM
	// -------------------------------------------------------------
	log.Println("Starting termination handler")
	killSignalChan := make(chan os.Signal, 0)
	doneChan := make(chan bool, 0)
	signal.Notify(killSignalChan, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// wait for kill signal
		<-killSignalChan
		// signal that we're done
		doneChan <- true
	}()

	// -------------------------------------------------------------
	// Render loop
	// -------------------------------------------------------------
	log.Println("Starting render loop (press CTRL+C to exit)")
	ram := vc64.NewC64RAM()
	currentFrame := 0
	startTime := time.Now()
	for {
		select {
		case <-doneChan:
			// Received done signal
			doCleanExit(client, cancelContext, &wg)
		default:
			// Render next frame
			frameBuffer := effects.DrawNestedCubes(currentFrame, frameBufferAddr, bitmapAddr, ram)

			// Send frame buffer to C64 Debugger
			if err = client.CPUMemoryWriteBlock(bitmapAddr, frameBuffer); err != nil {
				log.Fatal(err)
			}

			// wait for VSync (raster breakpoint)
			if <-vsyncChannel {
				if err = client.ContinueEmulation(); err != nil {
					log.Fatal(err)
				}
			}

			currentFrame++
			log.Printf("Frame: %d, FPS: %.2f", currentFrame, float64(currentFrame)/time.Since(startTime).Seconds())
		}
	}
}

// -------------------------------------------------------------
// Clean exit
// -------------------------------------------------------------
func doCleanExit(client *c64dws.Client, cancelContext context.CancelFunc, wg *sync.WaitGroup) {
	log.Println("Exiting...")
	log.Println("Removing raster breakpoint")
	client.RemoveRasterBreakpoint(rasterLineBreakpoint)
	log.Println("Requested hard reset")
	client.HardReset()
	if warpMode {
		log.Println("Disabling warp mode")
		client.SetWarpMode(false)
	}
	cancelContext()
	wg.Wait()
	log.Println("Bye!")
	os.Exit(0)
}

// -------------------------------------------------------------
// Reset C64D Emulator
// -------------------------------------------------------------
func resetEmulator(client *c64dws.Client, waitAfterReset bool) error {
	log.Println("Requested hard reset")
	if err := client.HardReset(); err != nil {
		return err
	}

	if waitAfterReset {
		log.Printf("Waiting %v after reset\n", waitTimeAfterResetInSeconds)
		time.Sleep(waitTimeAfterResetInSeconds)
	}
	return nil
}

// -------------------------------------------------------------
// Put CPU into infinite loop
// -------------------------------------------------------------
func putCPUIntoInfiniteLoop(client *c64dws.Client, vsyncChannel chan bool) error {
	// -------------------------------------------------------------
	// Wait for first VSync and pause emulation
	// -------------------------------------------------------------
	<-vsyncChannel

	// -------------------------------------------------------------
	// set $01 to $35 - whole RAM visible, except I/O area (0xD000-0xDFFF)
	// -------------------------------------------------------------
	if err := client.CPUMemoryWriteBlock(0x01, []byte{0x35}); err != nil {
		log.Fatal(err)
	}

	// Prepare simple assembly program which disables interrupts and then loops forever to keep CPU busy
	jmpAddr := uint16(0x0815)
	infiniteLoopProcedure := []byte{
		0x78,                                                       // $0815: sei
		0x4c, byte((jmpAddr + 1) & 0xff), byte((jmpAddr + 1) >> 8), // $0816: jmp $0816
	}
	if err := client.RAMWriteBlock(jmpAddr, infiniteLoopProcedure); err != nil {
		return err
	}

	// Run the program
	if err := client.CPUMakeJMP(jmpAddr); err != nil {
		return err
	}

	// -------------------------------------------------------------
	// Continue emulation
	// -------------------------------------------------------------
	if err := client.ContinueEmulation(); err != nil {
		log.Fatal(err)
	}

	return nil
}

// -------------------------------------------------------------
// Connect to C64D WebSocket Server
// -------------------------------------------------------------
func connectToWS(emulator c64dws.EmulatorType, apiType c64dws.APIType) (*c64dws.Client, error) {
	client := c64dws.NewDefaultClient(emulator, apiType)
	// Alternatively, you can create a custom client
	// client := c64dws.NewCustomClient(emulator, apiType, c64dws.TokenTypeUUID, "UUID:%s", c64dws.GetDefaultHost())
	log.Printf("Trying to connect to: %s", client.GetURL())
	_, err := client.Connect()
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to: %s", client.GetURL())

	return client, nil
}

// -------------------------------------------------------------
// Listen for messages from C64D WebSocket Server
// -------------------------------------------------------------
func listenForMessages(ctx context.Context, wg *sync.WaitGroup, client *c64dws.Client, vsyncChannel chan bool) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			// context cancelled
			return
		default:
			msgType, msg, err := client.ReceiveMessage()
			if err != nil {
				log.Fatal(err)
			}

			switch msgType {
			case c64dws.C64DRequestResponse:
				switch msg := msg.(type) {
				case c64dws.RequestResult:
					if !silentMode {
						log.Printf("Success: %d, result: %+v, token: '%s', binaryData: %v\n", msg.Status, msg.Result, msg.Token, msg.BinaryData)
					}
				case c64dws.RequestError:
					if !silentMode {
						log.Printf("Error: %d, error: %s\n", msg.Status, msg.Error)
					}
				}
			case c64dws.C64DServerEvent:
				switch msg := msg.(type) {
				case c64dws.RasterBreakpointEvent:
					if !silentMode {
						log.Printf("RasterBreakpointEvent: ID: %d, Raster: %d\n", msg.BreakpointId, msg.RasterLine)
					}
					vsyncChannel <- true
				}
			case c64dws.C64DUnknown:
				log.Println("Unknown message")
			}
		}
	}
}
