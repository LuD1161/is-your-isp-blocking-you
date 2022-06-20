package validators

import (
	"github.com/LuD1161/is-your-isp-blocking-you/internal/interfaces"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/models"
	airtelindia "github.com/LuD1161/is-your-isp-blocking-you/internal/validators/AirtelIndia"
)

var (
	allValidators = []interfaces.Validator{
		airtelindia.NewValidator(),
	}
)

func ValidatorResolver(ispResult models.IfConfigResponse) []interfaces.Validator {
	var validators []interfaces.Validator
	for _, validator := range allValidators {
		if (ispResult.Asn == validator.GetMetadata()["asn"]) || (ispResult.AsnOrg == validator.GetMetadata()["asn_org"]) || (ispResult.AsnOrg == validator.GetMetadata()["asn_org"]) {
			validators = append(validators, validator)
		}
	}
	return validators
}
