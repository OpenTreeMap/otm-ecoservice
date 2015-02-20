package endpoints

import (
	"errors"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"strconv"
	"time"
)

type ScenarioPostData struct {
	Region         string
	Instance_id    string
	Years          int
	Scenario_trees []ScenarioTree
}

type ScenarioTree struct {
	Otmcode    string
	Species_id int
	Region     string
	Diameters  []float64
}

type Scenario struct {
	Total map[string]float64
	Years []map[string]float64
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
func EcoScenarioPOST(db eco.DBContext, cache *cache.Cache) func(*ScenarioPostData) (*Scenario, error) {
	return func(data *ScenarioPostData) (*Scenario, error) {
		t := time.Now()

		scenarioTrees := data.Scenario_trees
		scenarioRegion := data.Region

		instanceId, err := strconv.Atoi(data.Instance_id)

		if err != nil {
			return nil, err
		}

		if len(scenarioRegion) == 0 {
			var regions []eco.Region
			regions, err = db.GetRegionsForInstance(
				cache.RegionGeometry, instanceId)

			if err != nil {
				return nil, err
			}

			if len(regions) == 1 {
				scenarioRegion = regions[0].Code
			}
		}

		yearTotals := make([][]float64, data.Years)
		grandTotals := make([]float64, len(eco.Factors))
		for i := range yearTotals {
			yearTotals[i] = make([]float64, len(eco.Factors))
		}

		for _, tree := range scenarioTrees {
			effectiveRegion := scenarioRegion
			if len(tree.Region) != 0 {
				effectiveRegion = tree.Region
			}

			factorDataForRegion, found := cache.RegionData[effectiveRegion]
			if !found {
				return nil, errors.New("No data is available for the iTree region with code " + effectiveRegion)
			}

			itreecode, err := cache.GetITreeCode(tree.Otmcode,
				tree.Species_id, effectiveRegion, instanceId)
			if err != nil {
				return nil, err
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

		years := make([]map[string]float64, data.Years)
		for i, a := range yearTotals {
			years[i] = eco.FactorArrayToMap(a)
		}
		return &Scenario{
			Total: eco.FactorArrayToMap(grandTotals),
			Years: years}, nil
	}
}
