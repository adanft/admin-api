package domain

import (
	"net/mail"
	"strings"
)

type Email string

func NewEmail(raw string) (Email, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if !isEmailValid(value) {
		return "", ErrInvalidEmail
	}

	return Email(value), nil
}

func (e Email) String() string {
	return string(e)
}

func isEmailValid(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return false
	}

	localPart, domainPart, found := strings.Cut(email, "@")
	if !found || localPart == "" || domainPart == "" {
		return false
	}
	if len(localPart) > 64 || len(domainPart) > 253 {
		return false
	}
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") || strings.Contains(domainPart, "..") {
		return false
	}

	labels := strings.Split(domainPart, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, char := range label {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' {
				continue
			}

			return false
		}
	}

	return len(labels[len(labels)-1]) >= 2
}
