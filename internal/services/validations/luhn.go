package validations

import "github.com/ShiraazMoollatjie/goluhn"

func LuhnValidate(str string) error {
	return goluhn.Validate(str)
}
