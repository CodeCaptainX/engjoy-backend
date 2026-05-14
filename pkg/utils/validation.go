package utils

import (
	"fmt"
	"reflect"
	"regexp"

	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ValidationUtilStruct struct {
	v *validator.Validate
}

func sliceIsNotEmpty(fl validator.FieldLevel) bool {
	slice, ok := fl.Field().Interface().([]int)
	if !ok {
		return false
	}
	return len(slice) > 0
}

func uniqueValidator(fl validator.FieldLevel) bool {
	value := fl.Field()
	if value.Kind() != reflect.Slice {
		return false
	}

	seen := make(map[interface{}]struct{})

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i).Interface()
		if _, exists := seen[elem]; exists {
			return false // Duplicate found
		}
		seen[elem] = struct{}{}
	}

	return true
}
func NewCustomValidator() *ValidationUtilStruct {
	validate := validator.New()

	return &ValidationUtilStruct{
		v: validate,
	}
}
func (u *ValidationUtilStruct) Bind(c *fiber.Ctx, payload interface{}) error {
	if err := c.QueryParser(payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success":     false,
			"message":     "Failed to parse query parameters.",
			"status_code": fiber.StatusUnprocessableEntity,
			"errors":      map[string]string{"query": "Failed to parse query parameters."},
			"data":        nil,
		})
	}

	// Validate the parsed query parameters
	if err := u.v.Struct(payload); err != nil {
		return u.HandleValidationError(c, err)
	}
	return nil
}

func (u *ValidationUtilStruct) HandleValidationError(c *fiber.Ctx, err error) error {
	errors := map[string]string{}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			errors[fieldErr.Field()] = formatValidationError(fieldErr.Field(), fieldErr.Tag())
		}
	} else {
		errors["validation"] = err.Error()
	}

	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"success":     false,
		"message":     "Validation failed.",
		"status_code": fiber.StatusUnprocessableEntity,
		"errors":      errors,
		"data":        nil,
	})
}

// IsISO8601Date checks if a given string is a valid ISO8601 formatted date.
// It returns true if the string matches the ISO8601 date format, otherwise false.
func IsISO8601Date(fl validator.FieldLevel) bool {
	//! Example: Correct format:  Date: "2024-08-30T12:34:56Z"
	ISO8601DateRegexString := `^(?:[1-9]\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d{1,9})?(?:Z|[+-][01]\d:[0-5]\d)$`
	ISO8601DateRegex := regexp.MustCompile(ISO8601DateRegexString)
	return ISO8601DateRegex.MatchString(fl.Field().String())
}

type ErrorResponse struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Tag     string `json:"tag,omitempty"`
}

func (u *ValidationUtilStruct) Struct(payload interface{}) []*ErrorResponse {
	u.v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("msg"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	var errors []*ErrorResponse

	err := u.v.Struct(payload)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, err := range validationErrors {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Message = getErrorMessage(err)
			errors = append(errors, &element)
		}
	}

	return errors
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "sliceIsNotEmpty":
		return err.Field() + " " + "should not be empty"
	case "uniqueValidator":
		return err.Field() + " " + "should not be duplicate"
	case "isodate":
		return err.Field() + " " + "not correct format"
	default:
		return err.Field() + " is " + err.Tag()
	}
}

type ErrorResponseSingleColumn struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

// NotEmpty validates that the specified field in the payload is not empty.
func (u *ValidationUtilStruct) NotEmpty(payload interface{}, field string) []*ErrorResponseSingleColumn {
	var errors []*ErrorResponseSingleColumn
	err := u.v.Var(payload, fmt.Sprintf("required=%s", field))
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, err := range validationErrors {
			var element ErrorResponseSingleColumn
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

func formatValidationError(fieldName, tag string) string {
	// Customize messages based on validation tag
	switch tag {
	case "required":
		return fieldName + " is required."
	case "min":
		return fieldName + " must have a minimum value."
	case "max":
		return fieldName + " must not exceed the maximum value."
	case "email":
		return fieldName + " must be a valid email address."
	case "oneof":
		return fieldName + " must be one of the allowed values."
	default:
		return fieldName + " has an invalid value."
	}
}

// ValidateDateRange validates that the end date is after the start date (if EndDate is not nil)
func ValidateDateRange(sl validator.StructLevel) {
	packagePlan := sl.Current().Interface().(struct {
		StartDate *time.Time
		EndDate   *time.Time `validate:"omitempty,gtfield=StartDate"`
	})

	// If EndDate is nil, skip validation
	if packagePlan.EndDate == nil {
		return
	}

	// If StartDate is nil, skip validation (assuming StartDate is required elsewhere)
	if packagePlan.StartDate == nil {
		return
	}

	// Check if EndDate is after StartDate
	if packagePlan.EndDate.Before(*packagePlan.StartDate) {
		sl.ReportError(packagePlan.EndDate, "EndDate", "EndDate", "gtfield", "StartDate")
	}
}

// NewValidatorDateRange creates a new validator and registers custom validations
func NewValidatorDateRange() *validator.Validate {
	validate := validator.New()
	validate.RegisterStructValidation(ValidateDateRange, struct {
		StartDate *time.Time
		EndDate   *time.Time `validate:"omitempty,gtfield=StartDate"`
	}{})
	return validate
}
