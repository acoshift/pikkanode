package validator

import (
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
	return arpc.WrapError(v.Validator.Error())
}

var rxTag = regexp.MustCompile(`^[\w\-]+$`)

func IsTag(str string) bool {
	return rxTag.MatchString(str)
}
