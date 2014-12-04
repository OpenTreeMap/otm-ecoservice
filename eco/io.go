package eco

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"
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

// A datafile a particular set of dbh breaks and data points
type Datafile struct {
	Breaks []float64
	Values map[string][]float64
}

func indexOf(value string, l []string) int {
	for p, v := range l {
		if v == value {
			return p
		}
	}
	return -1
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
// The returned data stucture maps is a map of
// regions to an array indexed by factor id
//
// For instance, Factors[3] = hydro_interception so
// the datafile for NoCalXXX region and hydro interception
// would be:
//
// files = LoadFiles()
// hydro_data = files['NoCalXXX'][3]
//
func LoadFiles(basePath string) map[string][]*Datafile {
	m := make(map[string][]*Datafile)

	files, _ := ioutil.ReadDir(basePath)
	for _, f := range files {
		// We only care about "output" files that have
		// been generated from itree streets
		if strings.Contains(f.Name(), "output") {
			parts := strings.Split(f.Name(), "__")
			region := parts[1]
			factor_with_csv := parts[2]
			// strip .csv
			factor := factor_with_csv[0 : len(factor_with_csv)-4]

			fidx := indexOf(factor, Factors)

			if fidx >= 0 {
				if m[region] == nil {
					m[region] = make([]*Datafile, len(Factors))
				}

				m[region][fidx] = LoadFile(basePath + f.Name())
			}
		}
	}

	return m
}

// TODO - Parse these all out to json
// The data files are in shambles due to the xls exporter
// we should clean these up once and for all
func LoadFile(path string) *Datafile {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// Scanner used to read line by line
	scanner := bufio.NewReader(fi)
	str, _ := scanner.ReadString(0x0A)

	str = strings.TrimSpace(str)
	headerFields := strings.Split(str, ",")[1:]
	breaks := make([]float64, 0, len(headerFields))

	// The first line is always the list of
	// diameter breaks
	for _, l := range headerFields {
		if len(l) > 0 {
			n, _ := strconv.ParseFloat(l, 64)
			breaks = append(breaks, n)
		}
	}

	tgtLen := len(breaks)

	// This maps from itree code to the
	// list of values at
	m := make(map[string][]float64)

	for {
		str, err := scanner.ReadString(0x0A)

		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		line := strings.Split(strings.TrimSpace(str), ",")

		if len(line) >= tgtLen {
			code := line[0]
			breaks := make([]float64, 0, len(breaks))

			for _, l := range line[1 : tgtLen+1] {
				n, _ := strconv.ParseFloat(l, 64)
				breaks = append(breaks, n)
			}

			m[code] = breaks
		}
	}

	file := &Datafile{breaks, m}

	return file
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
