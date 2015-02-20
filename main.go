package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"fmt"
	"github.com/azavea/ecobenefits/ecorest"
	"github.com/ungerik/go-rest"
	"log"
	"os"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	configpath = flag.String("configpath", "./", "path to the configuration")
)

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

	var cfg ecorest.Config

	err := gcfg.ReadFileInto(&cfg, *configpath)
	ecorest.PanicOnError(err)

	endpoints := ecorest.GetManager(cfg)

	rest.HandleGET("/itree_codes.json", endpoints.ITreeCodesGET)
	rest.HandleGET("/eco.json", endpoints.EcoGET)
	rest.HandlePOST("/eco_summary.json", endpoints.EcoSummaryPOST)
	rest.HandlePOST("/eco_scenario.json", endpoints.EcoScenarioPOST)

	rest.RunServer(fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port), nil)
}
