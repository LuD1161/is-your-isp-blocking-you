package generic

import (
	"io/ioutil"
	"strings"

	"github.com/LuD1161/is-your-isp-blocking-you/internal/constants"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/interfaces"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/models"
	"github.com/rs/zerolog/log"
)

var metadata = map[string]string{
	"name":       "Generic Validator",
	"tags":       "generic",
	"asn":        "",
	"asn_org":    "",
	"references": "",
	"method":     "connection reset, website redirection, DNS poisoning",
	"country":    "",
}

type validator struct {
	metadata map[string]string
}

func (v *validator) Validate(data models.ValidatorData) (int, error) {
	err := data.Err
	resp := data.Response
	finalURL := resp.Request.URL.String()

	// Method 1 : Check PR_CONNECTION_RESET first
	if strings.Contains(err.Error(), "connection reset by peer") {
		log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)
		return constants.CONN_RESET, nil
	}

	// Method 2 : Check Final URL
	if finalURL == "http://www.airtel.in/dot/" || finalURL == "https://www.airtel.in/dot/" {
		return constants.CONN_BLOCKED, nil
	}

	// Method 3 : Check Response Body
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error().Msgf("Error in reading response body : %s", err.Error())
			return constants.CONN_OK, err
		}
		if strings.Contains(string(body), "The website has been blocked as per order of Ministry of Electronics") {
			log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)
			return constants.CONN_BLOCKED, nil
		}
	}
	return constants.CONN_OK, nil
}

func (v *validator) GetMetadata() map[string]string {
	return v.metadata
}

func NewValidator() interfaces.Validator {
	return &validator{
		metadata: metadata,
	}
}
