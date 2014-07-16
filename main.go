//
// Todo
//
// * http://10.0.0.10:8080/eco.json?otmcode=ULAM&diameter=11&region=LoMidWXXX
//   negative natural gas?
// * support itree override table
//
package main

import (
	"code.google.com/p/gcfg"
	"errors"
	"flag"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/ungerik/go-rest"
	"log"
	"net/url"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	configpath = flag.String("configpath", "./", "path to the configuration")
)

type config struct {
	Database eco.DBInfo
	Data     struct {
		Path string
	}
	Server struct {
		Host string
		Port string
	}
}

// We can't marshall maps directly with
// go-rest so we just wrap it here
type BenefitsWrapper struct {
	Benefits map[string]float64
}

type SummaryPostData struct {
	Region      string
	Query       string
	Instance_id string
}

type ScenarioTree struct {
	Otmcode    string
	Species_id int
	Region     string
	Diameters  []float64
}

type ScenarioPostData struct {
	Region         string
	Instance_id    string
	Years          int
	Scenario_trees []ScenarioTree
}

type Scenario struct {
	Total map[string]float64
	Years []map[string]float64
}

// Given a values list return the single value
// associated with a given key or an error
func getSingleValue(in url.Values, key string) (string, error) {
	if keys, ok := in[key]; ok {
		if len(keys) == 1 {
			return keys[0], nil
		}
	}

	return "", errors.New(
		fmt.Sprintf("Missing or invalid %v parameter", key))
}

func getSingleIntValue(in url.Values, key string) (int, error) {
	str, err := getSingleValue(in, key)

	if err != nil {
		return 0, err
	}

	intv, err := strconv.Atoi(str)

	if err != nil {
		return 0, err
	}

	return intv, nil
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var cfg config
	err := gcfg.ReadFileInto(&cfg, *configpath)

	if err != nil {
		panic(err)
	}

	regiondata := eco.LoadFiles(cfg.Data.Path)
	speciesdata, err := eco.LoadSpeciesMap(cfg.Data.Path + "/species.json")

	if err != nil {
		panic(err)
	}

	dbraw, err := eco.OpenDatabaseConnection(&cfg.Database)

	if err != nil {
		panic(err)
	}

	defer dbraw.Close()

	db := (*eco.DBContext)(dbraw)

	overrides, err := db.GetOverrideMap()

	if err != nil {
		panic(err)
	}

	eco.InitGeos()
	regiongeometry, err := db.GetRegionGeoms()

	if err != nil {
		panic(err)
	}

	// This is implemented as an anonymous function so it can
	// close over the variables set at the start of the main
	// function, which can be expensive to load and only need to
	// be loaded once.
	getItreeCode := func(otmcode string, speciesId int, region string, instanceId int) (string, error) {
		speciesDataForRegion, found := speciesdata[region]
		if !found {
			return "", errors.New(fmt.Sprintf("Species data not found for the %v region",
				region))
		}

		itreeCode, foundItree := speciesDataForRegion[otmcode]
		notFoundMessage := fmt.Sprintf("iTree code not found for otmcode %v in region %v",
			otmcode, region)

		overidesForInstance, found := overrides[instanceId]
		if found {
			overridesForRegion, found := overidesForInstance[region]
			if found {
				overrideCode, found := overridesForRegion[speciesId]
				if found {
					itreeCode = overrideCode
					foundItree = true
				} else {
					notFoundMessage = fmt.Sprintf("There are overrides "+
						"defined for instance %v in the %v region "+
						"but not for species ID %v", instanceId, region, speciesId)
				}
			} else {
				notFoundMessage = fmt.Sprintf("There are overrides defined for "+
					"the instance, but not for the %v region", region)
			}
		}
		// It is normal for an instance to not have any
		// overrides defined, so there is no else block to set
		// an error message in the not-found case.

		if !foundItree {
			return "", errors.New(notFoundMessage)
		} else {
			return itreeCode, nil
		}
	}

	rest.HandleGET("/eco.json", func(in url.Values) (*BenefitsWrapper, error) {
		instanceid, err := getSingleIntValue(in, "instanceid")

		if err != nil {
			return nil, err
		}

		speciesid, err := getSingleIntValue(in, "speciesid")

		if err != nil {
			return nil, err
		}

		otmcode, err := getSingleValue(in, "otmcode")

		if err != nil {
			return nil, err
		}

		diameterstr, err := getSingleValue(in, "diameter")

		if err != nil {
			return nil, err
		}

		diameter, err := strconv.ParseFloat(diameterstr, 64)

		if err != nil {
			return nil, err
		}

		diameter = diameter * eco.CentimetersPerInch

		region, err := getSingleValue(in, "region")

		if err != nil {
			return nil, err
		}

		factorDataForRegion, found := regiondata[region]

		if !found {
			return nil, errors.New("invalid region")
		}

		itreecode, err := getItreeCode(otmcode, speciesid, region, instanceid)
		if err != nil {
			return nil, err
		}

		factorsum := make([]float64, len(eco.Factors))

		eco.CalcOneTree(
			factorDataForRegion,
			itreecode,
			diameter,
			factorsum)

		return &BenefitsWrapper{Benefits: eco.FactorArrayToMap(factorsum)}, nil
	})

	rest.HandlePOST("/eco_summary.json", func(data *SummaryPostData) (*BenefitsWrapper, error) {
		query := data.Query
		region := data.Region

		instanceid, err := strconv.Atoi(data.Instance_id)

		if err != nil {
			return nil, err
		}

		now := time.Now()

		// Using a fixed region lets us avoid costly
		// hash lookups. While we don't yet cache this value, we should
		// consider it since instance geometries change so rarely
		var regions []eco.Region

		if len(region) == 0 {
			regions, err = db.GetRegionsForInstance(
				regiongeometry, instanceid)

			if err != nil {
				return nil, err
			}

			if len(regions) == 1 {
				region = regions[0].Code
			}
		}

		// Contains the running total of the various factors
		instanceOverrides := overrides[instanceid]

		rows, err := db.ExecSql(query)

		s := time.Since(now)
		fmt.Println(int64(s/time.Millisecond), "ms (query)")

		if err != nil {
			return nil, err
		}

		factorsums, err :=
			eco.CalcBenefitsWithData(
				regions, rows, region,
				speciesdata, regiondata, instanceOverrides)

		s = time.Since(now)
		fmt.Println(int64(s/time.Millisecond), "ms (total)")

		if err != nil {
			return nil, err
		}

		return &BenefitsWrapper{Benefits: factorsums}, nil
	})

	// Take an array of prospective trees where each tree contains
	// an array of diamaters, one for each year the tree is alive,
	// and return an array of eco calulations, one for each year
	//
	// Trees will die of as part of the scenario, so the
	// `diameters` arrays for the trees may have different
	// lengths. Trees that die may be replaced with other trees,
	// so there will be trees that appear in the scenario at t >
	// 0, so the `diameters` array may have initial elements set
	// to 0.
	//
	// Specifying a "region" for an individual tree will override the
	// scenario-level "region" value.
	//
	// The "years" parameter must be >= the length of the longest
	// "diameters" array under "scenario_trees".
	//
	// Request (with bogus example parameters):
	//
	// POST /eco_scenario.json
	//
	// {
	//   "region": "NoEastXXX",
	//   "instance_id": 1,
	//   "years": 3
	//   "scenario_trees": [
	//     {
	//       "otmcode": "CACO",
	//       "species_id": 1,
	//       "region": "NoEastXXX",
	//       "diameters": [1, 1.3, 1.7]
	//     }
	//   ]
	// }
	//
	// Response (with bogus example values):
	//
	// {
	//   "Years": [
	//     {
	//       "aq_nox_avoided":     0.01548490,
	// 	 "aq_nox_dep":         0.00771784,
	// 	 "aq_pm10_avoided":    0.00546863
	//     },
	//     {
	//       "aq_nox_avoided":     0.02548420,
	// 	 "aq_nox_dep":         0.01973722,
	// 	 "aq_pm10_avoided":    0.00676823
	//     },
	//     {
	//       "aq_nox_avoided":     0.05484902,
	// 	 "aq_nox_dep":         0.04774471,
	// 	 "aq_pm10_avoided":    0.00946822
	//     }
	//   ],
	//   "Total": {
	//     "aq_nox_avoided": ... ,
	//     "aq_nox_dep": ... ,
	//     "aq_pm10_avoided": ...
	//   }
	// }
	rest.HandlePOST("/eco_scenario.json", func(data *ScenarioPostData) (*Scenario, error) {
		t := time.Now()

		scenarioTrees := data.Scenario_trees
		scenarioRegion := data.Region

		instanceId, err := strconv.Atoi(data.Instance_id)

		if err != nil {
			return nil, err
		}

		if len(scenarioRegion) == 0 {
			var regions []eco.Region
			regions, err = db.GetRegionsForInstance(
				regiongeometry, instanceId)

			if err != nil {
				return nil, err
			}

			if len(regions) == 1 {
				scenarioRegion = regions[0].Code
			}
		}

		yearTotals := make([][]float64, data.Years)
		grandTotals := make([]float64, len(eco.Factors))
		for i := range yearTotals {
			yearTotals[i] = make([]float64, len(eco.Factors))
		}

		for _, tree := range scenarioTrees {
			effectiveRegion := scenarioRegion
			if len(tree.Region) != 0 {
				effectiveRegion = tree.Region
			}

			factorDataForRegion, found := regiondata[effectiveRegion]
			if !found {
				return nil, errors.New("No data is available for the iTree region with code " + effectiveRegion)
			}

			itreecode, err := getItreeCode(tree.Otmcode,
				tree.Species_id, effectiveRegion, instanceId)
			if err != nil {
				return nil, err
			}

			for i, diameter := range tree.Diameters {
				factorSum := make([]float64, len(eco.Factors))
				eco.CalcOneTree(
					factorDataForRegion,
					itreecode,
					diameter,
					factorSum)
				for j, value := range factorSum {
					yearTotals[i][j] = value
					grandTotals[j] += value
				}
			}
		}

		// The requests are written to stdout like this:
		// 2014/07/15 14:06:10 POST /eco_scenario.json
		// Indenting the timing report aligns it with the http
		// verb on the previous line.
		fmt.Println("                   ",
			int64(time.Since(t)/time.Millisecond), "ms (total)")

		years := make([]map[string]float64, data.Years)
		for i, a := range yearTotals {
			years[i] = eco.FactorArrayToMap(a)
		}
		return &Scenario{
			Total: eco.FactorArrayToMap(grandTotals),
			Years: years}, nil
	})

	rest.RunServer(
		fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port), nil)
}
