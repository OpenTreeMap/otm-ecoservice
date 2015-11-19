package endpoints

import (
	"github.com/OpenTreeMap/otm-ecoservice/eco"
	"github.com/OpenTreeMap/otm-ecoservice/ecorest/cache"
)

type FullBenefitsPostData struct {
	Region      string
	Query       string
	Instance_id string
}

// We can't marshall maps directly with
// go-rest so we just wrap it here
type FullBenefitsWrapper struct {
	Benefits map[string]map[string]float64
}

func EcoFullBenefitsPOST(cache *cache.Cache) func(*FullBenefitsPostData) (*FullBenefitsWrapper, error) {
	return func(data *FullBenefitsPostData) (*FullBenefitsWrapper, error) {
		factorsum, err := GetFactorsum(cache, eco.CalcFullBenefitsWithData,
			data.Query, data.Region, data.Instance_id)

		if err != nil {
			return nil, err
		}

		return &FullBenefitsWrapper{Benefits: factorsum}, nil
	}
}
