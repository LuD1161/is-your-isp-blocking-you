package interfaces

import "github.com/LuD1161/is-your-isp-blocking-you/internal/models"

type Validator interface {
	Validate(models.ValidatorData) bool
	GetMetadata() map[string]string
}
