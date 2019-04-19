package validator

import (
	"net/http"
	"regexp"

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

var rxTag = regexp.MustCompile(`^[\w\-]+$`)

func IsTag(str string) bool {
	return rxTag.MatchString(str)
}
