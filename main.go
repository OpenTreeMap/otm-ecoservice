package main

import (
	"flag"
	"fmt"
	"github.com/OpenTreeMap/otm-ecoservice/ecorest"
	"github.com/OpenTreeMap/otm-ecoservice/ecorest/config"
	"github.com/ungerik/go-rest"
	"log"
	"os"
	"runtime/pprof"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
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

	cfg := config.LoadConfig()

	endpoints := ecorest.GetManager(cfg)

	rest.HandleGET("/itree_codes.json", endpoints.ITreeCodesGET)
	rest.HandleGET("/eco.json", endpoints.EcoGET)
	rest.HandlePOST("/eco_summary.json", endpoints.EcoSummaryPOST)
	rest.HandlePOST("/eco_scenario.json", endpoints.EcoScenarioPOST)
	rest.HandleGET("/invalidate_cache", endpoints.InvalidateCacheGET)

	rest.RunServer(fmt.Sprintf("%v:%v", cfg.ServerHost, cfg.ServerPort), nil)
}
