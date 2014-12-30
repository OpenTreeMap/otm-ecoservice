package endpoints

import (
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
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
