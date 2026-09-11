package handlers

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseSportIDs(value string) ([]int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []int{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		id, err := strconv.Atoi(part)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid integer value %q", part)
		}

		result = append(result, id)
	}

	return result, nil
}

func ParseLeagueRanks(value string) ([]int, error) {
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || v < 0 {
			return nil, fmt.Errorf("invalid league rank: %q", part)
		}

		result = append(result, v)
	}

	return result, nil
}

func SportIDsToString(ids []int) string {
	if len(ids) == 0 {
		return ""
	}

	parts := make([]string, len(ids))

	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}

	return strings.Join(parts, ",")
}
