package utils

import (
	"strconv"
	"strings"
)

// JoinWithAND joins a slice of strings with AND operator
func JoinWithAND(clauses []string) string {
	return strings.Join(clauses, " AND ")
}

// SplitCommaString splits a comma-separated string into a slice
func SplitCommaString(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	// Trim spaces from each part
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// ParseIntWithDefault parses a string into an integer with a default value
func ParseIntWithDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
