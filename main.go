package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"fmt"
	"github.com/azavea/ecobenefits/ecorest"
	"github.com/azavea/ecobenefits/ecorest/config"
	"log"
	"net/http"
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

	var cfg config.Config

	err := gcfg.ReadFileInto(&cfg, *configpath)
	config.PanicOnError(err)

	endpoints := ecorest.GetManager(cfg)

	http.HandleFunc("/itree_codes.json", endpoints.ITreeCodesGET)
	http.HandleFunc("/eco.json", endpoints.EcoGET)
	http.HandleFunc("/eco_summary.json", endpoints.EcoSummaryPOST)
	http.HandleFunc("/eco_scenario.json", endpoints.EcoScenarioPOST)
	http.HandleFunc("/invalidate_cache", endpoints.InvalidateCacheGET)

	hostInfo := fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port)

	log.Println("Server listening at ", hostInfo)
	err = http.ListenAndServe(hostInfo, nil)

	if err != nil {
		log.Fatal(err)
	}
}
