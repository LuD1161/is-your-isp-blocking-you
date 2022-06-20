package validators

import (
	"github.com/LuD1161/is-your-isp-blocking-you/internal/interfaces"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/models"
	airtelindia "github.com/LuD1161/is-your-isp-blocking-you/internal/validators/AirtelIndia"
	generic "github.com/LuD1161/is-your-isp-blocking-you/internal/validators/Generic"
)

var (
	allValidators = []interfaces.Validator{
		airtelindia.NewValidator(),
	}
)

func ValidatorResolver(ispResult models.IfConfigResponse) interfaces.Validator {
	for _, validator := range allValidators {
		// TODO : Add check on the basis of tags, name substring match etc
		if (ispResult.Asn == validator.GetMetadata()["asn"]) || (ispResult.AsnOrg == validator.GetMetadata()["asn_org"]) || (ispResult.AsnOrg == validator.GetMetadata()["asn_org"]) {
			return validator
		}
	}
	// If no validator is found, return the generic validator
	return generic.NewValidator()
}
