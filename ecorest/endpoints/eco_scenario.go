package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"net/http"
	"strconv"
	"time"
)

type scenarioTree struct {
	Otmcode    string
	Species_id int
	Region     string
	Diameters  []float64
}

// Take an array of prospective trees where each tree contains
// an array of diamaters, one for each year the tree is alive,
// and return an array of eco calulations, one for each year
//
// Trees will die of as part of the scenario, so the
// `diameters` arrays for the trees may have different
// lengths. Trees that die may be replaced with other trees,
// so there will be trees that appear in the scenario at t >
// 0, so the `diameters` array may have initial elements set
// to 0.
//
// Specifying a "region" for an individual tree will override the
// scenario-level "region" value.
//
// The "years" parameter must be >= the length of the longest
// "diameters" array under "scenario_trees".
//
// Request (with bogus example parameters):
//
// POST /eco_scenario.json
//
// {
//   "region": "NoEastXXX",
//   "instance_id": 1,
//   "years": 3
//   "scenario_trees": [
//     {
//       "otmcode": "CACO",
//       "species_id": 1,
//       "region": "NoEastXXX",
//       "diameters": [1, 1.3, 1.7]
//     }
//   ]
// }
//
// Response (with bogus example values):
//
// {
//   "Years": [
//     {
//       "aq_nox_avoided":     0.01548490,
// 	 "aq_nox_dep":         0.00771784,
// 	 "aq_pm10_avoided":    0.00546863
//     },
//     {
//       "aq_nox_avoided":     0.02548420,
// 	 "aq_nox_dep":         0.01973722,
// 	 "aq_pm10_avoided":    0.00676823
//     },
//     {
//       "aq_nox_avoided":     0.05484902,
// 	 "aq_nox_dep":         0.04774471,
// 	 "aq_pm10_avoided":    0.00946822
//     }
//   ],
//   "Total": {
//     "aq_nox_avoided": ... ,
//     "aq_nox_dep": ... ,
//     "aq_pm10_avoided": ...
//   }
// }
func EcoScenarioPOST(cache *cache.Cache) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		if request.Method != "POST" {
			http.Error(writer, "", http.StatusMethodNotAllowed)
			return
		}

		request.ParseForm()
		data := request.PostForm
		t := time.Now()

		yearsLen, err := strconv.ParseInt(data["Years"][0], 10, 0)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO: this might not be right
		scenarioTrees := data["Scenario_trees"]
		scenarioRegion := data["Region"][0]

		instanceId, err := strconv.Atoi(data["Instance_id"][0])

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if len(scenarioRegion) == 0 {
			var regions []eco.Region
			regions, err = cache.Db.GetRegionsForInstance(
				cache.RegionGeometry, instanceId)

			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
			}

			if len(regions) == 1 {
				scenarioRegion = regions[0].Code
			}
		}

		yearTotals := make([][]float64, yearsLen)
		grandTotals := make([]float64, len(eco.Factors))
		for i := range yearTotals {
			yearTotals[i] = make([]float64, len(eco.Factors))
		}

		for _, treeJSON := range scenarioTrees {
			var tree scenarioTree
			err = json.Unmarshal([]byte(treeJSON), &tree)

			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			effectiveRegion := scenarioRegion
			if len(tree.Region) != 0 {
				effectiveRegion = tree.Region
			}

			factorDataForRegion, found := cache.RegionData[effectiveRegion]
			if !found {
				err = errors.New("No data is available for the iTree region with code " + effectiveRegion)
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}

			itreecode, err := cache.GetITreeCode(tree.Otmcode,
				tree.Species_id, effectiveRegion, instanceId)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}

			for i, diameter := range tree.Diameters {
				factorSum := make([]float64, len(eco.Factors))
				eco.CalcOneTree(
					factorDataForRegion,
					itreecode,
					diameter,
					factorSum)
				for j, value := range factorSum {
					yearTotals[i][j] = value
					grandTotals[j] += value
				}
			}
		}

		// The requests are written to stdout like this:
		// 2014/07/15 14:06:10 POST /eco_scenario.json
		// Indenting the timing report aligns it with the http
		// verb on the previous line.
		fmt.Println("                   ",
			int64(time.Since(t)/time.Millisecond), "ms (total)")

		years := make([]map[string]float64, yearsLen)
		for i, a := range yearTotals {
			years[i] = eco.FactorArrayToMap(a)
		}

		scenarioMap := map[string]interface{}{
			"Total": eco.FactorArrayToMap(grandTotals),
			"Years": years,
		}
		j, err := json.Marshal(scenarioMap)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(writer, string(j))

	}
}
