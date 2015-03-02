package endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"net/http"
	"strconv"
	"time"
)

func EcoSummaryPOST(cache *cache.Cache) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		query := request.PostForm["Query"][0]
		region := request.PostForm["Region"][0]

		instanceid, err := strconv.Atoi(request.PostForm["Instance_id"][0])

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now()

		// Using a fixed region lets us avoid costly
		// hash lookups. While we don't yet cache this value, we should
		// consider it since instance geometries change so rarely
		var regions []eco.Region

		if len(region) == 0 {
			regions, err = cache.Db.GetRegionsForInstance(
				cache.RegionGeometry, instanceid)

			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
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
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		factorsums, err :=
			eco.CalcBenefitsWithData(
				regions, rows, region,
				cache.SpeciesData, cache.RegionData, instanceOverrides)

		s = time.Since(now)
		fmt.Println(int64(s/time.Millisecond), "ms (total)")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		benefitsMap := map[string]map[string]float64{"Benefits": factorsums}
		j, err := json.Marshal(benefitsMap)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(writer, string(j))
	}
}
