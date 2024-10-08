package validation

import (
	"regexp"
)

// Extracts the domain from an email address.
//
// Ex: "test@gmail" -> "@gmail.com"
func ExtractEmailDomain(email string) (string, error) {
	pattern := `\@[A-Za-z0-9]+\.[A-Za-z]{2,6}`
	input := []byte(email)
	regex, err := regexp.Compile(pattern)
	res := regex.FindString(string(input))
	if err != nil && len(res) > 2 { // sanity check
		return "", err
	}
	return res, nil
}
