package domain

import "strings"

func NormalizeIdentity(identity string) string {
	identity = strings.TrimSpace(identity)
	if strings.Contains(identity, "@") {
		return strings.ToLower(identity)
	}

	return identity
}
