package endpoints

import (
	"github.com/azavea/ecobenefits/eco"
)

type ITreeCodes struct {
	Codes map[string][]string
}

func ItreeCodesGET(regiondata map[string][]*eco.Datafile) func() (*ITreeCodes) {
	return func() *ITreeCodes {
		codes := eco.GetITreeCodesByRegion(regiondata)
		return &ITreeCodes{Codes: codes}
	}
}
