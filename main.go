//
// Todo
//
// * http://10.0.0.10:8080/eco.json?otmcode=ULAM&diameter=11&region=LoMidWXXX
//   negative natural gas?
// * eco.py says that diameter should be in cm?! but database is in inches
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type config struct {
	Database eco.DBInfo
	Server   struct {
		Host string
		Port string
	}
}

// We can't marshall maps directly with
// go-rest so we just wrap it here
type wrapper struct {
	Benefits map[string]float64
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
	err := gcfg.ReadFileInto(&cfg, "./config.gcfg")

	if err != nil {
		panic(err)
	}

	regiondata := eco.LoadFiles("./data/")
	speciesdata, err := eco.LoadSpeciesMap("data/species.json")

	if err != nil {
		panic(err)
	}

	dbraw, err := eco.OpenDatabaseConnection(&cfg.Database)

	if err != nil {
		panic(err)
	}

	defer dbraw.Close()

	db := (*eco.DBContext)(dbraw)

	rest.HandleGET("/eco.json", func(in url.Values) (*wrapper, error) {
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

		itreecode, found := speciesDataForRegion[otmcode]

		if !found {
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

	rest.HandleGET("/eco_summary.json", func(in url.Values) (*wrapper, error) {
		instancestr, err := getSingleValue(in, "instance_id")

		if err != nil {
			return nil, err
		}

		instanceid, err := strconv.Atoi(instancestr)

		if err != nil {
			return nil, err
		}

		// Each instance can override itree codes
		// ITreeCodeOverride
		// TODO... bummer...

		now := time.Now()

		// Contains the running total of the various factors
		factorsums, err :=
			eco.CalcBenefits(db, instanceid, speciesdata, regiondata)

		s := time.Since(now)
		fmt.Println(int64(s/time.Millisecond), "ms")

		if err != nil {
			return nil, err
		}

		return &wrapper{Benefits: factorsums}, nil

	})

	rest.RunServer(
		fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
		stopServerChan)
}
