package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateBody[T any](r *http.Request) (T, error) {
	var data T

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("invalid request format: please check your JSON data")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(data); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Process all validation errors
			var errorMessages []string
			for _, fieldError := range validationErrors {
				switch fieldError.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldError.Field()))
				case "email":
					errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address", fieldError.Field()))
				case "min":
					errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param()))
				case "max":
					errorMessages = append(errorMessages, fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param()))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s is invalid", fieldError.Field()))
				}
			}
			return data, fmt.Errorf("validation error: %s", strings.Join(errorMessages, "; "))
		}
		return data, fmt.Errorf("validation error: invalid data provided")
	}

	return data, nil
}
