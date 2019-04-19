package validator

import (
	"net/http"

	"github.com/acoshift/arpc"
	"github.com/moonrhythm/validator"
)

type Validator struct {
	validator.Validator
}

func New() *Validator {
	return new(Validator)
}

func (v *Validator) Error() error {
	err := v.Validator.Error()
	if err != nil {
		return arpc.NewError(http.StatusBadRequest, err.Error())
	}
	return nil
}
