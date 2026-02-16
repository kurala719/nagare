package utils

import (
	"fmt"
	"strings"
)

// ParseItemValue formats an item's value with its units
func ParseItemValue(value, units string) string {
	value = strings.TrimSpace(value)
	units = strings.TrimSpace(units)

	if value == "" {
		return ""
	}

	if units == "" {
		return value
	}

	return fmt.Sprintf("%s %s", value, units)
}
