package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest"
	"github.com/azavea/ecobenefits/ecorest/endpoints"
	"github.com/ungerik/go-rest"
	"log"
	"os"
	"runtime/pprof"
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

	getItreeCode := ecorest.MakeItreeCodeCache(overrides, speciesdata)

	rest.HandleGET("/itree_codes.json",
		endpoints.ItreeCodesGET(regiondata))

	rest.HandleGET("/eco.json",
		endpoints.EcoGET(regiondata, getItreeCode))

	rest.HandlePOST("/eco_summary.json",
		endpoints.EcoSummaryPOST(regiondata, db, regiongeometry, overrides, speciesdata, getItreeCode))

	rest.HandlePOST("/eco_scenario.json",
		endpoints.EcoScenarioPOST(regiondata, db, regiongeometry, getItreeCode))

	rest.RunServer(
		fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port), nil)
}
