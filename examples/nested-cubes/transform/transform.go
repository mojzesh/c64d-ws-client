package transform

import "math"

type Vector3D struct {
	X float64
	Y float64
	Z float64
}

func NewVector3D(x, y, z float64) Vector3D {
	return Vector3D{
		X: x,
		Y: y,
		Z: z,
	}
}

const piDiv180 = math.Pi / 180.0

func ScaleVector3D(vector Vector3D, scale float64) Vector3D {
	return NewVector3D(vector.X*scale, vector.Y*scale, vector.Z*scale)
}

func TranslateVector3D(vector Vector3D, translate Vector3D) Vector3D {
	return NewVector3D(vector.X+translate.X, vector.Y+translate.Y, vector.Z+translate.Z)
}

func AngleToRadians(angle float64) float64 {
	return piDiv180 * angle
}

func FaceVisibilityTest(v1 Vector3D, v2 Vector3D, v3 Vector3D) float64 {
	return (v3.X-v2.X)*(v2.Y-v1.Y) - (v2.X-v1.X)*(v3.Y-v2.Y)
}

func Perspective(vertex Vector3D, zoom float64, distance float64) Vector3D {
	zPlusDist := vertex.Z + distance
	return NewVector3D(
		(vertex.X/(zPlusDist))*zoom,
		(vertex.Y/(zPlusDist))*zoom,
		vertex.Z,
	)
}

// perspX = (vertex.X / (vertex.Z+distance)) * zoom
// perspY = (vertex.Y / (vertex.Z+distance)) * zoom
// vertex.X = (perspX / zoom) * (vertex.Z+distance)
// vertex.Y = (perspY / zoom) * (vertex.Z+distance)
func PerspectiveInverse(vertex Vector3D, zoom float64, distance float64) Vector3D {
	zPlusDist := vertex.Z + distance
	return NewVector3D(
		(vertex.X/zoom)*(zPlusDist),
		(vertex.Y/zoom)*(zPlusDist),
		vertex.Z,
	)
}

func RotateX(angle float64, vector Vector3D) Vector3D {
	angleInRad := AngleToRadians(angle)
	sina := math.Sin(angleInRad)
	cosa := math.Cos(angleInRad)
	return NewVector3D(
		vector.X,
		vector.Y*cosa-vector.Z*sina,
		vector.Y*sina+vector.Z*cosa,
	)
}

func RotateY(angle float64, vector Vector3D) Vector3D {
	angleInRad := AngleToRadians(angle)
	sina := math.Sin(angleInRad)
	cosa := math.Cos(angleInRad)
	return NewVector3D(
		vector.Z*sina+vector.X*cosa,
		vector.Y,
		vector.Z*cosa-vector.X*sina,
	)
}

func RotateZ(angle float64, vector Vector3D) Vector3D {
	angleInRad := AngleToRadians(angle)
	sina := math.Sin(angleInRad)
	cosa := math.Cos(angleInRad)
	return NewVector3D(
		vector.X*cosa-vector.Y*sina,
		vector.X*sina+vector.Y*cosa,
		vector.Z,
	)
}

// -------------------------------------
// FORMULAS
// -------------------------------------
// For rotation around the x-axis:
// x1 = x
// y1 = y*cos(angleX) - z*sin(angleX);
// z1 = y*sin(angleX) + z*cos(angleX);
// -------------------------------------
// For rotation around the y-axis:
// x2 = x1*cos(angleY) - z1*sin(angleY);
// y2 = y1
// z2 = x1*sin(angleY) + z1*cos(angleY);
// -------------------------------------
// For rotation around the z-axis:
// x3 = x2*cos(angleZ) - y2*sin(angleZ);
// y3 = x2*sin(angleZ) + y2*cos(angleZ);
// z3 = z2
// -------------------------------------
func RotateXYZ(angleX float64, angleY float64, angleZ float64, vector Vector3D) Vector3D {
	angleXinRad := angleX * piDiv180
	angleYinRad := angleY * piDiv180
	angleZinRad := angleZ * piDiv180
	sinax := math.Sin(angleXinRad)
	cosax := math.Cos(angleXinRad)
	sinay := math.Sin(angleYinRad)
	cosay := math.Cos(angleYinRad)
	sinaz := math.Sin(angleZinRad)
	cosaz := math.Cos(angleZinRad)
	// RotateX
	tx1 := vector.X
	ty1 := vector.Y*cosax - vector.Z*sinax
	tz1 := vector.Y*sinax + vector.Z*cosax
	// RotateY
	tx2 := tx1*cosay - tz1*sinay
	ty2 := ty1
	tz2 := tx1*sinay + tz1*cosay
	// let tx2 = tz * sinay + tx * cosay;
	// let ty2 = ty;
	// let tz2 = tz * cosay - tx * sinay;
	return NewVector3D(
		// RotateZ
		tx2*cosaz-ty2*sinaz,
		tx2*sinaz+ty2*cosaz,
		tz2,
	)
}
