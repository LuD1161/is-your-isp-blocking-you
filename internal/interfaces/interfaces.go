package interfaces

import "github.com/LuD1161/is-your-isp-blocking-you/cmd"

type Validator interface {
	Validate(cmd.ValidatorData) bool
	GetMetadata() map[string]string
}
