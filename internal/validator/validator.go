package validator

import (
	"log"
	"regexp"
	"strings"
	"unicode"
)

var (
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)	
)


type Validator struct {
	Errors map[string]string;
}


func New() *Validator{
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	log.Print("Validator Check: ", ok, " key: ", key)
	if !ok {
		v.AddError(key, message)
	}
}


func Matches(value string, rx *regexp.Regexp) bool { 
	return rx.MatchString(value)
}

func ValidatePassword(pw string) bool {
	if len(pw) < 6 || len(pw) > 12 {
		return false
	}

	var hasUpper bool
	var hasNumber bool
	var hasSpecial bool

	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*", r):
			hasSpecial = true
		}
	}

	return hasUpper && hasNumber && hasSpecial
}
