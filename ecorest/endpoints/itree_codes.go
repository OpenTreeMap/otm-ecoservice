package endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/azavea/ecobenefits/eco"
	"github.com/azavea/ecobenefits/ecorest/cache"
	"net/http"
)

func ITreeCodesGET(cache *cache.Cache) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		codes := eco.GetITreeCodesByRegion(cache.RegionData)
		j, _ := json.Marshal(map[string]map[string][]string{"Codes": codes})
		fmt.Fprint(writer, string(j))
	}
}
