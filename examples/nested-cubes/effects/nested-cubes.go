package effects

import (
	"math"
	"time"

	"math/rand"

	"github.com/mojzesh/nested-cubes/drawing"
	"github.com/mojzesh/nested-cubes/transform"
	"github.com/mojzesh/nested-cubes/vc64"
)

// -------------------------------------------------------------
// Cube count
// -------------------------------------------------------------
const cubesCount = 9
const facesCount = 6

const viewportHeight = 200
const viewportColumns = 40
const frontbufferFrameBufferSize = viewportColumns * viewportHeight
const backbufferFrameBufferSize = frontbufferFrameBufferSize + (viewportColumns * 16)

// -------------------------------------------------------------
// Cube vertexes
// -------------------------------------------------------------
var vertexes = []transform.Vector3D{
	{X: -1, Y: -1, Z: +1}, // 0 - x, y, z
	{X: +1, Y: -1, Z: +1}, // 1 - x, y, z
	{X: +1, Y: +1, Z: +1}, // 2 - x, y, z
	{X: -1, Y: +1, Z: +1}, // 3 - x, y, z
	{X: -1, Y: -1, Z: -1}, // 4 - x, y, z
	{X: +1, Y: -1, Z: -1}, // 5 - x, y, z
	{X: +1, Y: +1, Z: -1}, // 6 - x, y, z
	{X: -1, Y: +1, Z: -1}, // 7 - x, y, z
}

// -------------------------------------------------------------
// Cube faces indexes
// -------------------------------------------------------------
var faces = [][]byte{
	{0, 1, 2, 3},
	{1, 5, 6, 2},
	{5, 4, 7, 6},
	{4, 0, 3, 7},
	{7, 3, 2, 6},
	{4, 5, 1, 0},
}

// -------------------------------------------------------------
// Cube colors
// -------------------------------------------------------------
var c64CubeColors [][]byte

// -------------------------------------------------------------
//
// -------------------------------------------------------------
func init() {
	// Generate random colors for each cube face
	rnd := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	c64CubeColors = make([][]byte, cubesCount)
	for c0 := 0; c0 < cubesCount; c0++ {
		c64CubeColors[c0] = make([]byte, facesCount)
		for c1 := 0; c1 < facesCount; c1++ {
			c64CubeColors[c0][c1] = byte(rnd.Uint32())
		}
	}
}

// -------------------------------------------------------------
//
// -------------------------------------------------------------
func DrawNestedCubes(currentTime int, frameBufferAddr uint16, bitmapAddr uint16, ram *vc64.C64RAM) []byte {
	const firstCubeSize int = 30
	const speedFactor float64 = 0.5

	frameWithSpeed := float64(currentTime) * speedFactor

	amp := 180.0
	for currC0 := 0; currC0 < cubesCount; currC0++ {
		cubeColors := c64CubeColors[currC0]
		angle := frameWithSpeed + float64(currC0*2)
		angleX := amp * math.Cos(math.Pi/180.0*angle)
		angleY := amp * math.Sin(math.Pi/180.0*angle)
		angleZ := amp * math.Sin(math.Pi/180.0*angle)
		scale := float64(firstCubeSize - (firstCubeSize / cubesCount * currC0))
		drawC64Cube(angleX, angleY, angleZ, scale, frameBufferAddr, viewportColumns, viewportHeight, vertexes, faces, cubeColors, ram)
	}

	// drawing.EORFillBitmap(frameBufferAddr, bitmapAddr, 0, 40, 200, ram)
	drawing.EORFillEvenOddBitmap(frameBufferAddr, bitmapAddr, 0, viewportColumns, viewportHeight, ram)

	// copy frame buffer
	frameBuffer := make([]byte, backbufferFrameBufferSize)
	copy(frameBuffer, ram.CPUGetRAMSlice(bitmapAddr, backbufferFrameBufferSize))

	// clear backbuffer
	ram.CPUFillRAM(0, frameBufferAddr, backbufferFrameBufferSize)

	return frameBuffer
}

// -------------------------------------------------------------
//
// -------------------------------------------------------------
func drawC64Cube(
	angleX float64,
	angleY float64,
	angleZ float64,
	scale float64,
	frameBufferAddr uint16,
	columns uint16,
	rows uint16,
	vertexes []transform.Vector3D,
	faces [][]byte,
	c64CubeColors []byte,
	ram *vc64.C64RAM,
) {
	var fVtx0 transform.Vector3D
	var fVtx1 transform.Vector3D
	var fVtx2 transform.Vector3D
	var fVtx3 transform.Vector3D
	var fVtx4 transform.Vector3D
	var fVtx5 transform.Vector3D
	var patternValue byte

	var edge1LenX float64
	var edge1LenY float64
	var edge1LenZ float64
	var edge2LenX float64
	var edge2LenY float64
	var edge2LenZ float64

	width := 160
	height := 100
	dist := 64.0
	zoom := 140.0

	resVtx := make([]transform.Vector3D, 8)
	vCenterX := float64(width / 2.0)
	vCenterY := float64(height / 2.0)

	primIndexes := []int{0, 1, 3, 4}
	// secIndexes := []int{2, 6, 5, 7}
	allResVtxPersp := make([]transform.Vector3D, 8)

	for c0 := 0; c0 < 4; c0++ {
		vIdx := primIndexes[c0]
		inputVertex := vertexes[vIdx]
		inputVertex = transform.ScaleVector3D(inputVertex, scale)
		// inputVertex = Vector3D.Translate(inputVertex, new Vector3D(10, 10, 10))
		inputVertex = transform.RotateXYZ(angleX, angleY, angleZ, inputVertex)
		perspVertex := transform.Perspective(inputVertex, zoom, dist)
		resVtx[vIdx] = inputVertex
		allResVtxPersp[vIdx] = perspVertex
	}

	//-----------------------------------------
	fVtx0 = resVtx[0]
	fVtx3 = resVtx[3]
	fVtx4 = resVtx[4]

	fVtx1 = resVtx[1]

	edge1LenX = (fVtx4.X - fVtx0.X)
	edge1LenY = (fVtx4.Y - fVtx0.Y)
	edge1LenZ = (fVtx4.Z - fVtx0.Z)

	edge2LenX = (fVtx3.X - fVtx0.X)
	edge2LenY = (fVtx3.Y - fVtx0.Y)
	edge2LenZ = (fVtx3.Z - fVtx0.Z)

	//-----------------------------------------
	resVtx[7] = transform.Vector3D{
		X: fVtx3.X + edge1LenX,
		Y: fVtx3.Y + edge1LenY,
		Z: fVtx3.Z + edge1LenZ,
	}

	//-----------------------------------------
	resVtx[2] = transform.Vector3D{
		X: fVtx1.X + edge2LenX,
		Y: fVtx1.Y + edge2LenY,
		Z: fVtx1.Z + edge2LenZ,
	}

	//-----------------------------------------
	resVtx[5] = transform.Vector3D{
		X: fVtx1.X + edge1LenX,
		Y: fVtx1.Y + edge1LenY,
		Z: fVtx1.Z + edge1LenZ,
	}

	//-----------------------------------------
	fVtx5 = resVtx[5]
	resVtx[6] = transform.Vector3D{
		X: fVtx5.X + edge2LenX,
		Y: fVtx5.Y + edge2LenY,
		Z: fVtx5.Z + edge2LenZ,
	}

	//-----------------------------------------
	allResVtxPersp[2] = transform.Perspective(resVtx[2], zoom, dist)
	allResVtxPersp[7] = transform.Perspective(resVtx[7], zoom, dist)
	allResVtxPersp[5] = transform.Perspective(resVtx[5], zoom, dist)
	allResVtxPersp[6] = transform.Perspective(resVtx[6], zoom, dist)
	//-----------------------------------------
	for c0 := 0; c0 < 6; c0++ {
		fVtx1 = allResVtxPersp[faces[c0][0]]
		fVtx2 = allResVtxPersp[faces[c0][1]]
		fVtx3 = allResVtxPersp[faces[c0][2]]
		fVtx4 = allResVtxPersp[faces[c0][3]]

		patternValue = c64CubeColors[c0]

		if transform.FaceVisibilityTest(fVtx1, fVtx2, fVtx3) > 0 {
			drawFaceC64NormalInt(
				patternValue,
				int16(fVtx1.X+vCenterX),
				int16(fVtx1.Y+vCenterY),
				int16(fVtx2.X+vCenterX),
				int16(fVtx2.Y+vCenterY),
				int16(fVtx3.X+vCenterX),
				int16(fVtx3.Y+vCenterY),
				int16(fVtx4.X+vCenterX),
				int16(fVtx4.Y+vCenterY),
				frameBufferAddr,
				columns,
				rows,
				ram,
			)
		}
	}
}

// -------------------------------------------------------------
//
// -------------------------------------------------------------
func drawFaceC64NormalInt(
	patternValue byte,
	x1 int16,
	y1 int16,
	x2 int16,
	y2 int16,
	x3 int16,
	y3 int16,
	x4 int16,
	y4 int16,
	frameBufferAddr uint16,
	columns uint16,
	rows uint16,
	ram *vc64.C64RAM,
) {
	viewPortWidth := uint16(160)
	viewPortHeight := uint16(100)
	drawing.DrawBitmapLineToFillEvenOddWithClipping(x1, y1, x2, y2, patternValue, frameBufferAddr, columns, rows, viewPortWidth, viewPortHeight, ram)
	drawing.DrawBitmapLineToFillEvenOddWithClipping(x2, y2, x3, y3, patternValue, frameBufferAddr, columns, rows, viewPortWidth, viewPortHeight, ram)
	drawing.DrawBitmapLineToFillEvenOddWithClipping(x3, y3, x4, y4, patternValue, frameBufferAddr, columns, rows, viewPortWidth, viewPortHeight, ram)
	drawing.DrawBitmapLineToFillEvenOddWithClipping(x4, y4, x1, y1, patternValue, frameBufferAddr, columns, rows, viewPortWidth, viewPortHeight, ram)
}
