package global

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate
var d sync.Once

func GetValidator() *validator.Validate {
	d.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})
	return validate
}
