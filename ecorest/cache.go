package ecorest

import (
	"errors"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
)

func MakeItreeCodeCache(overrides map[int]map[string]map[int]string,
	speciesdata map[string]map[string]string) (func(string, int, string, int) (string, error)) {
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
