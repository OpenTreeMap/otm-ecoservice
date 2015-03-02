package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"net/http"
	"net/url"
	"strconv"
)

// Given a values list return the single value
// associated with a given key or an error
func getSingleValue(in url.Values, key string) (string, error) {
	if keys, ok := in[key]; ok {
		if len(keys) == 1 {
			return keys[0], nil
		}
	}

	return "", errors.New(
		fmt.Sprintf("Missing or invalid %v parameter", key))
}

func getSingleIntValue(in url.Values, key string) (int, error) {
	str, err := getSingleValue(in, key)

	if err != nil {
		return 0, err
	}

	intv, err := strconv.Atoi(str)

	if err != nil {
		return 0, err
	}

	return intv, nil
}

func EcoGET(cache *cache.Cache) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		instanceid, err := getSingleIntValue(request.Form, "instanceid")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		speciesid, err := getSingleIntValue(request.Form, "speciesid")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		otmcode, err := getSingleValue(request.Form, "otmcode")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		diameterstr, err := getSingleValue(request.Form, "diameter")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		diameter, err := strconv.ParseFloat(diameterstr, 64)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		diameter = diameter * eco.CentimetersPerInch

		region, err := getSingleValue(request.Form, "region")

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		factorDataForRegion, found := cache.RegionData[region]

		if !found {
			http.Error(writer, errors.New("invalid region").Error(),
				http.StatusBadRequest)
			return
		}
		itreecode, err := cache.GetITreeCode(otmcode, speciesid, region, instanceid)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		factorsum := make([]float64, len(eco.Factors))

		eco.CalcOneTree(
			factorDataForRegion,
			itreecode,
			diameter,
			factorsum)

		benefits := eco.FactorArrayToMap(factorsum)
		benefitsMap := map[string]map[string]float64{"Benefits": benefits}
		j, err := json.Marshal(benefitsMap)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(writer, string(j))
	}
}
