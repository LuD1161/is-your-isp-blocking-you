package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/rs/zerolog/log"
)

type Validator struct{}

func (v *Validator) Validate(data ValidatorData) (string, string, error) {
	resp := data.Response
	finalURL := ""

	if data.Err == nil {
		finalURL = resp.Request.URL.String()
	}

	log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)

	// Check DNS Filtering
	filtering, msg, err := v.CheckDNSFiltering(data.ResolvedIPs)
	if filtering != NO_FILTERING { // if no filtering then return from here. No need to check for HTTP or SNI filtering.
		return filtering, msg, err
	}

	// Method 1 : Check PR_CONNECTION_RESET first
	// TODO: Check with some un-censored source of truth. Fetch from database.
	if data.Err != nil && strings.Contains(data.Err.Error(), "connection reset by peer") {
		log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)
		return SNI_FILTERING, "", nil
	}

	// Method 2 : Check Final URL
	if finalURL == "http://www.airtel.in/dot/" || finalURL == "https://www.airtel.in/dot/" {
		return CONN_BLOCKED, "", nil
	}

	// Method 3 : Check HTTP Filtering
	if resp.StatusCode == 200 {
		return v.CheckHTTPFiltering(resp.Body)
	}

	return NO_FILTERING, "OK", nil
}

func (v *Validator) CheckHTTPFiltering(bodyReader io.ReadCloser) (string, string, error) {

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Error().Msgf("Error in reading response body : %s", err.Error())
		return OTHER_ERROR, "", err
	}
	// Wherever it redirects, this is the domain that hosts the Department of Telecom's (DoT) notice
	// Check body for strings
	for _, blockedString := range fYaml.HTTPFILTERING.Body {
		if strings.Contains(string(body), blockedString.Value) {
			return HTTP_FILTERING, blockedString.Value, nil
		}
	}
	return NO_FILTERING, "", nil
}

func (v *Validator) CheckDNSFiltering(resolvedIPS string) (string, string, error) {
	for _, blacklistedIP := range fYaml.DNSFILTERING {
		ips := strings.Split(resolvedIPS, ",")
		for _, ip := range ips {
			if ip == blacklistedIP.Value {
				validatorMsg := fmt.Sprintf("%s : %s", blacklistedIP.ISP, blacklistedIP.Value)
				return DNS_FILTERING, validatorMsg, nil
			}
		}
	}
	return NO_FILTERING, "", nil
}
