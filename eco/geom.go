package eco

// #cgo LDFLAGS: -lgeos_c
// #include <stdlib.h>
// #include <geos_c.h>
import "C"

import (
	"errors"
	"unsafe"
)

type Geom struct {
	geom   *[0]byte
	preped *[0]byte
}

type Point struct {
	coordseqptr *[0]byte
	pointptr    *[0]byte
}

type Region struct {
	Code string
	geom Geom
}

// initialize the geos system
// this must be called before any other geometry functions
var didInit = false

func InitGeos() {
	if !didInit {
		C.initGEOS(nil, nil)
	}
	didInit = true
}

// Determine if p intersects g
func Intersects(g Geom, p Point) (bool, error) {
	r := C.GEOSPreparedContains(g.preped, p.pointptr)

	if r == 1 {
		return true, nil
	}

	if r == 0 {
		return false, nil
	}

	// Note that r == 2 is the explicit error call here
	return false, errors.New("C call failed")
}

// Create a new geometry from the given
// wkt string
//
// The geometry will be 'prepared' to make
// intersects/contains queries faster
//
// The caller is responsible for destroying
// the returned geometry with "GeosDestroy"
func MakeGeosGeom(wkt string) Geom {
	reader := C.GEOSWKTReader_create()

	cwkt := C.CString(wkt)
	geom := C.GEOSWKTReader_read(reader, cwkt)

	prepgeom := C.GEOSPrepare(geom)

	C.free(unsafe.Pointer(cwkt))
	C.GEOSWKTReader_destroy(reader)

	return Geom{geom, prepgeom}
}

// The caller is responsible for destroying
// the returned geometry with "DestroyPt"
func CreateGeosPtWithXY(x float64, y float64) Point {
	coordseqptr := C.GEOSCoordSeq_create(1, 2)

	C.GEOSCoordSeq_setX(coordseqptr, 0, C.double(x))
	C.GEOSCoordSeq_setY(coordseqptr, 0, C.double(y))

	pointptr := C.GEOSGeom_createPoint(coordseqptr)

	return Point{coordseqptr, pointptr}
}

func GetXYOnSurface(g Geom) (float64, float64) {
	pt := C.GEOSGetCentroid(g.geom)

	var x C.double
	var y C.double

	coord := C.GEOSGeom_getCoordSeq(pt)

	err := C.GEOSCoordSeq_getX(coord, 0, &x)

	if err == -1 {
		panic(err)
	}

	err = C.GEOSCoordSeq_getY(coord, 0, &y)

	if err == -1 {
		panic(err)
	}

	return float64(x), float64(y)
}

func GeosDestroy(geom Geom) {
	C.GEOSPreparedGeom_destroy(geom.preped)
	C.GEOSGeom_destroy(geom.geom)
}

func DestroyPt(p Point) {
	C.GEOSGeom_destroy(p.pointptr)
}
