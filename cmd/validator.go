package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Validator struct{}

var (
	fYaml FilteringYAML
)

func (v *Validator) Validate(data ValidatorData) (string, string, error) {
	resp := data.Response
	finalURL := ""

	// Read filtering.yaml to load all filters
	yfile, err := ioutil.ReadFile("filtering.yaml")
	if err != nil {
		return OTHER_ERROR, "", fmt.Errorf("error parsing filtering.yaml : %s", err.Error())
	}

	if err := yaml.Unmarshal(yfile, &fYaml); err != nil {
		return OTHER_ERROR, "", fmt.Errorf("error parsing filtering.yaml : %s", err.Error())
	}

	if data.Err == nil {
		finalURL = resp.Request.URL.String()
	}

	log.Debug().Msgf("URL : %s | Error : %+v", finalURL, data.Err)
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
