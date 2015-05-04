package ecorest

import (
	"github.com/azavea/ecobenefits/ecorest/cache"
	"github.com/azavea/ecobenefits/ecorest/config"
	"github.com/azavea/ecobenefits/ecorest/endpoints"
	"net/http"
)

type restManager struct {
	ITreeCodesGET      http.HandlerFunc
	EcoGET             http.HandlerFunc
	EcoSummaryPOST     http.HandlerFunc
	EcoScenarioPOST    http.HandlerFunc
	InvalidateCacheGET http.HandlerFunc
}

func GetManager(cfg config.Config) *restManager {
	ecoCache, invalidateCache := cache.Init(cfg)
	invalidateCache()

	invalidateCacheGET := func(_ http.ResponseWriter, _ *http.Request) {
		invalidateCache()
	}

	return &restManager{endpoints.ITreeCodesGET(ecoCache),
		endpoints.EcoGET(ecoCache),
		endpoints.EcoSummaryPOST(ecoCache),
		endpoints.EcoScenarioPOST(ecoCache),
		invalidateCacheGET}
}
