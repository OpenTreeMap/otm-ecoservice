package endpoints

import (
	"github.com/OpenTreeMap/ecoservice/eco"
	"github.com/OpenTreeMap/ecoservice/ecorest/cache"
)

type ITreeCodes struct {
	Codes map[string][]string
}

func ITreeCodesGET(cache *cache.Cache) func() *ITreeCodes {
	return func() *ITreeCodes {
		codes := eco.GetITreeCodesByRegion(cache.RegionData)
		return &ITreeCodes{Codes: codes}
	}
}
