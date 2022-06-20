package airtelindia

import (
	"io/ioutil"
	"strings"

	"github.com/LuD1161/is-your-isp-blocking-you/internal/constants"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/interfaces"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/models"
	"github.com/rs/zerolog/log"
)

var metadata = map[string]string{
	"name":       "Airtel India Validator",
	"tags":       "airtel, optical fiber, broadband",
	"asn":        "AS24560",
	"asn_org":    "Bharti Airtel Ltd., Telemedia Services",
	"references": "https://github.com/captn3m0/airtel-blocked-hosts/blob/airtel-fiber/airtel-fiber-blocked-hosts.txt",
	"method":     "connection reset, http filtering, DNS poisoning",
	"country":    "India",
}

type validator struct {
	metadata map[string]string
}

func (v *validator) Validate(data models.ValidatorData) (int, error) {
	err := data.Err
	resp := data.Response
	finalURL := ""

	if err == nil {
		finalURL = resp.Request.URL.String()
	}

	log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)
	// Method 1 : Check PR_CONNECTION_RESET first
	if err != nil && strings.Contains(err.Error(), "connection reset by peer") {
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
			return constants.OTHER_ERROR, err
		}
		// Wherever it redirects, this is the domain that hosts the Department of Telecom's (DoT) notice
		if strings.Contains(string(body), "www.airtel.in/dot/") {
			log.Debug().Msgf("URL : %s | Body : %s", finalURL, body)
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
