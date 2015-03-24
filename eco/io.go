package eco

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

var (
	// List of "factor" files that we will load. Files generally have the form:
	// output__{region}__{factor}.csv
	Factors = []string{"natural_gas", "electricity",
		"hydro_interception", "co2_sequestered",
		"co2_avoided", "co2_storage", "aq_nox_dep", "aq_ozone_dep",
		"aq_nox_avoided", "aq_pm10_dep", "aq_pm10_avoided",
		"aq_sox_dep", "aq_sox_avoided", "aq_voc_avoided", "bvoc"}
)

// A datafile contains a particular set of dbh breaks and data points
// For example, the datafile:
// Datafile{[12, 15, 17, 200], [5, 10, 15, 20]}
// would imply that diameter values of 0-12 have a benefit value
// of 5, 12-15 have a benefit of 10 value of 10, and so on.
//
// TODO: clarify whether breaks are inclusive/exclusive and at
// which ends of the range.
type Datafile struct {
	// Breaks are values that form the endpoints of a range
	// of diameter values that all receive the same ecobenefit value.
	Breaks []float64
	Values map[string][]float64
}

// Load the master species map
//
// The output map combines regions with species codes
// and returns the iTree Code
func LoadSpeciesMap(speciesMasterList string) (map[string]map[string]string, error) {
	bytes, err := ioutil.ReadFile(speciesMasterList)

	if err != nil {
		return nil, err
	}

	data := make(map[string]map[string]string)
	err = json.Unmarshal(bytes, &data)

	return data, err
}

// Load the data files
//
// the relevant data files are stored in the format:
// output__<regioncode>__<factor_with_csv??>.csv
//
// for example:
// output__TropicPacXXX__property_value.csv
//
// The returned data stucture has region codes as keys and an array of data files
// as values.
// the data file array is indexed by factor id, as determined by the global `Factors`.
//
// For instance, Factors[3] = hydro_interception so
// the datafile for NoCalXXX region and hydro interception
// would be:
//
// files = LoadFiles('/path/to/data/folder')
// hydro_data = files['NoCalXXX'][3]
//
func LoadFiles(basePath string) map[string][]*Datafile {
	prefix := "output"
	extension := ".json"
	m := make(map[string][]*Datafile)

	files, _ := ioutil.ReadDir(basePath)
	for _, f := range files {
		// We only care about "output" files that have
		// been generated from itree streets
		if strings.Contains(f.Name(), prefix) {
			name := f.Name()[0 : len(f.Name())-len(extension)]
			parts := strings.Split(name, "__")
			region := parts[1]
			factor := parts[2]

			for fidx, fcat := range Factors {
				if fcat == factor {
					if m[region] == nil {
						m[region] = make([]*Datafile, len(Factors))
					}
					jdata, _ := ioutil.ReadFile(basePath + f.Name())

					var val Datafile
					json.Unmarshal(jdata, &val)

					m[region][fidx] = &val
					break
				}
			}
		}
	}
	return m
}

func GetITreeCodesByRegion(regionData map[string][]*Datafile) map[string][]string {
	// Return valid i-Tree codes for each i-Tree region, e.g.:
	//     { 'CaNCCoJBK': ['AB', 'AC', ...],
	//       'CenFlaXXX': ['ACAC2', 'ACNE', ...],
	//       ...}
	codes := make(map[string][]string, len(regionData))
	for regionCode, data := range regionData {
		// All value maps for a region use the same i-Tree codes, so just use the first one
		valueMap := data[0].Values
		keys := make([]string, 0, len(valueMap))
		for k := range valueMap {
			keys = append(keys, k)
		}
		codes[regionCode] = keys
	}
	return codes
}
