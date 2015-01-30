package cache

import (
	"errors"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/config"
)

type speciesDataMap map[string]map[string]string

type overridesMap map[int]map[string]map[int]string

type regionDataMap map[string][]*eco.Datafile

type regionGeometryMap map[int]eco.Region

type iTreeCodeRetrieverFunc func(string, int, string, int) (string, error)

type Cache struct {
	RegionData     regionDataMap
	RegionGeometry regionGeometryMap
	Overrides      overridesMap
	SpeciesData    speciesDataMap
	GetITreeCode   iTreeCodeRetrieverFunc
	Db             eco.DBContext
}

func Init(cfg config.Config) (*Cache, func()) {
	cache := &Cache{}
	return cache, func() {
		dbraw, err := eco.OpenDatabaseConnection(&cfg.Database)
		config.PanicOnError(err)

		// this is tricky. Though the db connection will close at
		// the end of this function, we cast it and keep a pointer in
		// the cache, so the actual struct will live on, and be used
		// for data to open new database connections
		db := (*eco.DBContext)(dbraw)
		defer dbraw.Close()

		eco.InitGeos()

		regiondata := eco.LoadFiles(cfg.Data.Path)
		speciesdata, err := eco.LoadSpeciesMap(cfg.Data.Path + "/species.json")
		config.PanicOnError(err)
		overrides, err := db.GetOverrideMap()
		config.PanicOnError(err)

		regiongeometry, err := db.GetRegionGeoms()
		config.PanicOnError(err)

		retriever := makeItreeCodeRetriever(overrides, speciesdata)
		cache.RegionData = regiondata
		cache.RegionGeometry = regiongeometry
		cache.Overrides = overrides
		cache.SpeciesData = speciesdata
		cache.GetITreeCode = retriever
		cache.Db = *db
	}
}

func makeItreeCodeRetriever(overrides overridesMap, speciesdata speciesDataMap) iTreeCodeRetrieverFunc {
	// This is implemented as a curried function so it can
	// close over the variables set at the start of the main
	// function, which can be expensive to load and only need to
	// be loaded once.
	return func(otmcode string, speciesId int, region string, instanceId int) (string, error) {
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
}
