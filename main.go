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
type wrapper struct {
	Benefits map[string]float64
}

type PostData struct {
	Region      string
	Query       string
	Instance_id string
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

	// See RunServer below
	stopServerChan := make(chan bool)

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

	rest.HandleGET("/eco.json", func(in url.Values) (*wrapper, error) {
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

		speciesDataForRegion, found := speciesdata[region]

		if !found {
			return nil, errors.New("invalid region")
		}

		factorDataForRegion, found := regiondata[region]

		if !found {
			return nil, errors.New("invalid region")
		}

		itreecode, founditree := speciesDataForRegion[otmcode]

		overidesForInstance, found := overrides[instanceid]

		if found {
			overridesForRegion, found := overidesForInstance[region]

			if found {
				overrideCode, found := overridesForRegion[speciesid]

				if found {
					itreecode = overrideCode
					founditree = true
				}
			}
		}

		if !founditree {
			return nil, errors.New("invalid otm code for region")
		}

		factorsum := make([]float64, len(eco.Factors))

		eco.CalcOneTree(
			factorDataForRegion,
			itreecode,
			diameter,
			factorsum)

		return &wrapper{Benefits: eco.FactorArrayToMap(factorsum)}, nil
	})

	rest.HandlePOST("/eco_summary.json", func(data *PostData) (*wrapper, error) {
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

		return &wrapper{Benefits: factorsums}, nil
	})

	rest.RunServer(
		fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
		stopServerChan)
}
