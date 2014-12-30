package endpoints

import (
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"strconv"
	"time"
)

type SummaryPostData struct {
	Region      string
	Query       string
	Instance_id string
}

func EcoSummaryPOST(regiondata map[string][]*eco.Datafile,
	db eco.DBContext,
	regiongeometry map[int]eco.Region,
	overrides map[int]map[string]map[int]string,
	speciesdata map[string]map[string]string,
	getItreeCode (func(string, int, string, int) (string, error))) func(*SummaryPostData) (*BenefitsWrapper, error) {
	return func(data *SummaryPostData) (*BenefitsWrapper, error) {
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
	}
}
