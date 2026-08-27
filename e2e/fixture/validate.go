package fixture

import (
	playvalidator "github.com/go-playground/validator/v10"
)

// structValidator adapts go-playground/validator to excelizex.Validator.
type structValidator struct {
	v *playvalidator.Validate
}

// StructValidator returns a validator backed by validator.New().
// Wire it on Read via Validate(fixture.StructValidator()) so DTO validate tags apply.
func StructValidator() structValidator {
	return structValidator{v: playvalidator.New()}
}

// Validate implements excelizex.Validator. Row must be a pointer to the bound DTO.
func (s structValidator) Validate(row any) error {
	return s.v.Struct(row)
}
