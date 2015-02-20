package ecorest

import (
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"github.com/azavea/ecobenefits/ecorest/endpoints"
	"net/url"
)

type Config struct {
	Database eco.DBInfo
	Data     struct {
		Path string
	}
	Server struct {
		Host string
		Port string
	}
}

type restManager struct {
	ITreeCodesGET   (func() *endpoints.ITreeCodes)
	EcoGET          (func(url.Values) (*endpoints.BenefitsWrapper, error))
	EcoSummaryPOST  (func(*endpoints.SummaryPostData) (*endpoints.BenefitsWrapper, error))
	EcoScenarioPOST (func(*endpoints.ScenarioPostData) (*endpoints.Scenario, error))
}

func GetManager(cfg Config) *restManager {
	regiondata := eco.LoadFiles(cfg.Data.Path)
	speciesdata, err := eco.LoadSpeciesMap(cfg.Data.Path + "/species.json")
	PanicOnError(err)

	dbraw, err := eco.OpenDatabaseConnection(&cfg.Database)
	PanicOnError(err)

	defer dbraw.Close()

	db := (*eco.DBContext)(dbraw)

	overrides, err := db.GetOverrideMap()
	PanicOnError(err)

	eco.InitGeos()
	regiongeometry, err := db.GetRegionGeoms()
	PanicOnError(err)

	ecoCache := cache.MakeCache(regiondata, regiongeometry, overrides, speciesdata)

	return &restManager{endpoints.ITreeCodesGET(ecoCache),
		endpoints.EcoGET(ecoCache),
		endpoints.EcoSummaryPOST(*db, ecoCache),
		endpoints.EcoScenarioPOST(*db, ecoCache)}
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
