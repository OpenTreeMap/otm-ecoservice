package eco

import (
	"strconv"
)

var CentimetersPerInch = 2.54

// Data backends are used to fetch the actual tree
// data. The main backend right now is a postgres database.
//
// The testing code also provides a mock implementation
type DataBackend interface {
	// Given a mapping of region ids to region geoms
	// return the set of regions that intersect the given
	// instance
	GetRegionsForInstance(
		regions map[int]Region, instance int) ([]Region, error)

	// Get a fetchable with the given sql string
	ExecSql(string) (Fetchable, error)

	// Get a map for all of the overrides on the database
	// The map should end up looking something like:
	// instanceid -> region -> species id -> itreecode
	GetOverrideMap() (map[int]map[string]map[int]string, error)

	// Get all itree geometries
	GetRegionGeoms() (map[int]Region, error)
}

// Fetchables come out of the database backend and essentially
// wrap a database rowset
//
// The testing code also provides a mock implementation
type Fetchable interface {
	// This method should only be called on a "region" fetchable
	// object and will get the current record's data
	//
	// The diameter will be in centimeters
	GetDataWithRegion(
		id *int, diameter *float64, otmcode *string,
		speciesid *int, x *float64, y *float64) error

	// This method can be called on any fetchable object and
	// will get the current record's data
	//
	// The diameter will be in centimeters
	GetDataWithoutRegion(
		id *int, diameter *float64, otmcode *string, speciesid *int) error

	// Closes this fetchable
	Close() error

	// Move to the next item in the internal iterator
	// returns false if there are no more records
	Next() bool
}

// Calculate ecobenefits over an instance in the given backend
//
// Regions are a list of intersecting regions the check. This can
// be nil or empty only if the "region" parameter is specified
//
// To use a fixed region pass in a valid "region" parameter
// if this parameter is passed in the regions array will be
// ignored
//
// Rows is the fetchable set to use
//
// speciesdata is a map to itreecode:
// region --> otmcode --> itreecode
//
// Static data can be loaded for this via LoadSpeciesMap
//
// regiondata maps regions to factor lists
// region -> slice of datafiles
//
// Overrides is a map like:
// region -> species id -> itree code
//
// That allows itree overrides on a per-species/instance level
//
// Note that the ith element of the datafiles slice is
// the ith factor from eco.Factors
//
func CalcBenefitSummaryWithData(
	regions []Region,
	rows Fetchable,
	region string,
	speciesdata map[string]map[string]string,
	regiondata map[string][]*Datafile,
	overrides map[string]map[int]string) (map[string]float64, error) {

	return calcBenefitsWithData(regions, rows, region, speciesdata, regiondata, overrides, nil)
}

// CalcFullBenefitsWithData performs the same operation as CalcSummaryBenefitsWithData
// but returns data on a per-tree basis instead of summarized for every tree
func CalcFullBenefitsWithData(
	regions []Region,
	rows Fetchable,
	region string,
	speciesdata map[string]map[string]string,
	regiondata map[string][]*Datafile,
	overrides map[string]map[int]string) (map[string]map[string]float64, error) {

	outputData := make(map[string]map[string]float64)
	_, err := calcBenefitsWithData(regions, rows, region, speciesdata, regiondata, overrides, outputData)

	return outputData, err
}

func calcBenefitsWithData(
	regions []Region,
	rows Fetchable,
	region string,
	speciesdata map[string]map[string]string,
	regiondata map[string][]*Datafile,
	overrides map[string]map[int]string,
	outputData map[string]map[string]float64) (map[string]float64, error) {

	useFixedRegion := len(region) > 0
	factorsum := make([]float64, len(Factors))
	ntrees := 0

	id := 0
	diameter := 0.0
	otmcode := ""
	speciesid := 0
	x := 0.0
	y := 0.0

	// Last index of found polygon
	lastidx := 0
	regionlen := len(regions)

	var speciesDataForRegion map[string]string
	var factorDataForRegion []*Datafile
	var overridesForRegion map[int]string

	if useFixedRegion {
		speciesDataForRegion = speciesdata[region]
		factorDataForRegion = regiondata[region]

		if overrides != nil {
			overridesForRegion = overrides[region]
		}
	}

	itreecode := ""
	var err error

	for rows.Next() {
		if useFixedRegion {
			err = rows.GetDataWithoutRegion(
				&id, &diameter, &otmcode, &speciesid)
		} else {
			err = rows.GetDataWithRegion(
				&id, &diameter, &otmcode, &speciesid, &x, &y)

			region = ""

			pt := CreateGeosPtWithXY(x, y)
			defer DestroyPt(pt)

			for i := 0; i < regionlen; i += 1 {
				// consecutive trees have a high spatial
				// correlation so try the last successful
				// polygon first
				calcidx := (i + lastidx) % regionlen
				regiongeom := regions[calcidx]

				intersects, err := Intersects(
					regiongeom.geom, pt)

				if err != nil {
					return nil, err
				}

				if intersects {
					lastidx = calcidx
					region = regiongeom.Code
					break
				}
			}

			speciesDataForRegion = speciesdata[region]
			factorDataForRegion = regiondata[region]

			if overrides != nil {
				overridesForRegion = overrides[region]
			}
		}

		if err != nil {
			return nil, err
		}

		itreecode = speciesDataForRegion[otmcode]

		if overridesForRegion != nil {
			itreecodeOver, found := overridesForRegion[speciesid]

			if found {
				itreecode = itreecodeOver
			}
		}

		if itreecode != "" {
			CalcOneTree(
				factorDataForRegion,
				itreecode,
				diameter,
				factorsum)

			if outputData != nil {
				outputData[strconv.Itoa(id)] = FactorArrayToMap(factorsum)
				factorsum = make([]float64, len(Factors))
			}

			ntrees += 1
		}
	}

	factormap := FactorArrayToMap(factorsum)
	factormap["n_trees"] = float64(ntrees)

	return factormap, nil
}

// Convert an array of factors into a map by
// matching up their indicies
func FactorArrayToMap(factors []float64) map[string]float64 {
	factormap := make(map[string]float64)
	for i, factor := range Factors {
		factormap[factor] = factors[i]
	}

	return factormap
}

// Calculate benefits for a single tree
//
// The diameter must be in centimeters
//
// The benefits will be added to the factorsum slice
func CalcOneTree(
	factorDataForRegion []*Datafile,
	itreecode string,
	diameter float64,
	factorsum []float64) {
	nfactors := len(factorsum)

	for fidx := 0; fidx < nfactors; fidx++ {
		data := factorDataForRegion[fidx]

		// This is the slowest part of this function (the
		// has lookup in the map). If we could make this
		// faster it would do the algo good
		values := data.Values[itreecode]
		breaks := data.Breaks

		if len(values) == 0 {
			continue
		}

		dbhRangeMin := 0.0
		dbhRangeMax := 1.0

		factorMinValue := 0.0
		factorMaxValue := 0.0

		lastBreakIdx := len(breaks) - 1
		maxdiameter := breaks[lastBreakIdx]

		// Clamp to maximum diameter, as
		// per eco.py/itree strees spec
		if diameter >= maxdiameter {
			dbhRangeMin = breaks[lastBreakIdx-1]
			dbhRangeMax = breaks[lastBreakIdx]

			factorMinValue = values[lastBreakIdx-1]
			factorMaxValue = values[lastBreakIdx]
		} else {
			for i, diameterBreak := range breaks {
				// If we're in this break, be fit between
				// break[i-1] and break[i]
				if diameter < diameterBreak {
					// Since we don't have zero values
					// if diameter < break[0] we have
					// we used the fixed value at
					// of values[0]
					if i == 0 {
						// Treated as a special
						// case below
						dbhRangeMin = breaks[0]
						dbhRangeMax = breaks[0]

						factorMinValue = values[0]
						factorMaxValue = values[0]
					} else {
						dbhRangeMin = breaks[i-1]
						dbhRangeMax = breaks[i]

						factorMinValue = values[i-1]
						factorMaxValue = values[i]
					}

					break
				}
			}
		}

		var factorValue float64

		// Fixed point, use the factorMinValue
		if dbhRangeMin == dbhRangeMax {
			factorValue = factorMinValue
		} else {
			// m = Δy/Δx
			factorPerUnitDiameter :=
				(factorMaxValue - factorMinValue) / (dbhRangeMax - dbhRangeMin)

			// b = y₀ - mx₀
			factorIntercept :=
				factorMaxValue - factorPerUnitDiameter*dbhRangeMax

			// y = mx + b
			factorValue = factorPerUnitDiameter*diameter + factorIntercept
		}

		factorsum[fidx] = factorsum[fidx] + factorValue
	}
}
