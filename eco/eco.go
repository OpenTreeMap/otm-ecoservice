package eco

var CentimetersPerInch = 2.54

// Data backends are used to fetch the actual tree
// data. The main backend right now is a postgres database.
//
// The testing code also provides a mock implementation
type DataBackend interface {
	// If an instance only contains a single region or the
	// given instance has a region override set, return
	// the single instance
	//
	// If the instance has multiple overlapping regions, returns
	// the empty string. If the instance has no overlapping regions
	// this function also return the empty string
	GetRegionForInstance(int) (string, error)

	// Get a fetchable for all trees with a diameter and species in
	// a given instance - the region data will not be fetched
	RowsForTreesWithoutRegion(string, ...string) (Fetchable, error)

	// Get a fetchable for all trees with a diameter and species in
	// a given instance - including the spatial join over the
	// regions table
	RowsForTreesWithRegion(string, ...string) (Fetchable, error)
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
	GetDataWithRegion(diameter *float64, otmcode *string, region *string) error

	// This method can be called on any fetchable object and
	// will get the current record's data
	//
	// The diameter will be in centimeters
	GetDataWithoutRegion(diameter *float64, otmcode *string) error

	// Closes this fetchable
	Close() error

	// Move to the next item in the internal iterator
	// returns false if there are no more records
	Next() bool
}

// Calculate ecobenefits over an instance in the given backend
//
// speciesdata is a map to itreecode:
// region --> otmcode --> itreecode
//
// Static data can be loaded for this via LoadSpeciesMap
//
// regiondata maps regions to factor lists
// region -> slice of datafiles
//
// Note that the ith element of the datafiles slice is
// the ith factor from eco.Factors
//
func CalcBenefits(
	db DataBackend,
	speciesdata map[string]map[string]string,
	regiondata map[string][]*Datafile,
	instanceid int,
	where string,
	params ...string) (map[string]float64, error) {

	// Using a fixed region lets us avoid costly
	// hash lookups. While we don't yet cache this value, we should
	// consider it since instace geometries change so rarely
	region, err := db.GetRegionForInstance(instanceid)

	if err != nil {
		return nil, err
	}

	useFixedRegion := len(region) > 0

	var rows Fetchable

	if useFixedRegion {
		rows, err = db.RowsForTreesWithoutRegion(where, params...)
	} else {
		rows, err = db.RowsForTreesWithRegion(where, params...)
	}

	if err != nil {
		panic(err)
		return nil, err
	}

	defer rows.Close()

	factorsum := make([]float64, len(Factors))
	ntrees := 0

	diameter := 0.0
	otmcode := ""

	var speciesDataForRegion map[string]string
	var factorDataForRegion []*Datafile

	if useFixedRegion {
		speciesDataForRegion = speciesdata[region]
		factorDataForRegion = regiondata[region]
	}

	itreecode := ""

	for rows.Next() {
		if useFixedRegion {
			err = rows.GetDataWithoutRegion(&diameter, &otmcode)
		} else {
			err = rows.GetDataWithRegion(&diameter, &otmcode, &region)

			speciesDataForRegion = speciesdata[region]
			factorDataForRegion = regiondata[region]
		}

		if err != nil {
			return nil, err
		}

		itreecode = speciesDataForRegion[otmcode]

		if itreecode != "" {
			CalcOneTree(
				factorDataForRegion,
				itreecode,
				diameter,
				factorsum)

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
