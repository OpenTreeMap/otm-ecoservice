package endpoints

import (
	"fmt"
	"github.com/OpenTreeMap/otm-ecoservice/eco"
	"github.com/OpenTreeMap/otm-ecoservice/ecorest/cache"
	"strconv"
	"time"
)

type SummaryPostData struct {
	Region      string
	Query       string
	Instance_id string
}

type calculatorFn func([]eco.Region, eco.Fetchable, string, map[string]map[string]string, map[string][]*eco.Datafile, map[string]map[int]string) (map[string]map[string]float64, error)

func EcoSummaryPOST(cache *cache.Cache) func(*SummaryPostData) (*BenefitsWrapper, error) {
	return func(data *SummaryPostData) (*BenefitsWrapper, error) {
		factorsum, err := GetFactorsum(cache, eco.CalcBenefitSummaryWithData,
			data.Query, data.Region, data.Instance_id)

		if err != nil {
			return nil, err
		}

		return &BenefitsWrapper{Benefits: factorsum["summary"]}, nil
	}
}

func GetFactorsum(cache *cache.Cache, calculatorFn calculatorFn, query string, region string, instanceIdStr string) (map[string]map[string]float64, error) {

	instanceid, err := strconv.Atoi(instanceIdStr)

	if err != nil {
		return nil, err
	}


	now := time.Now()

	// Using a fixed region lets us avoid costly
	// hash lookups. While we don't yet cache this value, we should
	// consider it since instance geometries change so rarely
	var regions []eco.Region

	if len(region) == 0 {
		regions, err := cache.Db.GetRegionsForInstance(
			cache.RegionGeometry, instanceid)

		if err != nil {
			return nil, err
		}

		if len(regions) == 1 {
			region = regions[0].Code
		}
	}

	// Contains the running total of the various factors
	instanceOverrides := cache.Overrides[instanceid]

	rows, err := cache.Db.ExecSql(query)

	s := time.Since(now)
	fmt.Println(int64(s/time.Millisecond), "ms (query)")

	if err != nil {
		return nil, err
	}

	factorsums, err :=
		calculatorFn(
			regions, rows, region,
			cache.SpeciesData, cache.RegionData, instanceOverrides)

	s = time.Since(now)
	fmt.Println(int64(s/time.Millisecond), "ms (total)")

	if err != nil {
		return nil, err
	}

	return factorsums, nil
}

