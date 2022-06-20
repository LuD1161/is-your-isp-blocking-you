package interfaces

import "github.com/LuD1161/is-your-isp-blocking-you/internal/models"

type Validator interface {
	Validate(models.ValidatorData) (int, error)
	GetMetadata() map[string]string
}
