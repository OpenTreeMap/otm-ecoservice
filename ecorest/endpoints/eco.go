package endpoints

import (
	"errors"
	"fmt"
	"github.com/OpenTreeMap/otm-ecoservice/eco"
	"github.com/OpenTreeMap/otm-ecoservice/ecorest/cache"
	"net/url"
	"strconv"
)

// We can't marshall maps directly with
// go-rest so we just wrap it here
type BenefitsWrapper struct {
	Benefits map[string]float64
}

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

func EcoGET(cache *cache.Cache) func(url.Values) (*BenefitsWrapper, error) {
	return func(in url.Values) (*BenefitsWrapper, error) {
		instanceid, err := getSingleIntValue(in, "instanceid")

		if err != nil {
			return nil, err
		}

		speciesid, err := getSingleIntValue(in, "speciesid")

		if err != nil {
			return nil, err
		}

		otmcode, err := getSingleValue(in, "otmcode")

		if err != nil {
			return nil, err
		}

		diameterstr, err := getSingleValue(in, "diameter")

		if err != nil {
			return nil, err
		}

		diameter, err := strconv.ParseFloat(diameterstr, 64)

		if err != nil {
			return nil, err
		}

		diameter = diameter * eco.CentimetersPerInch

		region, err := getSingleValue(in, "region")

		if err != nil {
			return nil, err
		}

		factorDataForRegion, found := cache.RegionData[region]

		if !found {
			return nil, errors.New("invalid region")
		}

		itreecode, err := cache.GetITreeCode(otmcode, speciesid, region, instanceid)
		if err != nil {
			return nil, err
		}

		factorsum := make([]float64, len(eco.Factors))

		eco.CalcOneTree(
			factorDataForRegion,
			itreecode,
			diameter,
			factorsum)

		return &BenefitsWrapper{Benefits: eco.FactorArrayToMap(factorsum)}, nil
	}
}
