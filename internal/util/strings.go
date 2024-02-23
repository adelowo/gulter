package util

import "strings"

func IsStringEmpty(s string) bool { return len(strings.TrimSpace(s)) == 0 }
