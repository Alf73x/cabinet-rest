package handlers

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseSportIDs(value string) ([]int, error) {
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty sport id")
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid sport id: %s", part)
		}
		if id <= 0 {
			return nil, fmt.Errorf("sport id must be positive: %d", id)
		}
		ids = append(ids, id)
	}
	return ids, nil
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
