package main

import (
	"strconv"
	s "strings"
)

type Range struct {
	From int64
	To   int64
}

func ParseRangeRequest(request string) ([]Range, error) {
	a := s.Replace(request, " ", "", -1)
	a = s.Replace(a, "bytes=", "", -1)
	ranges := []Range{}
	for _, part := range s.Split(a, ",") {
		parts := s.Split(part, "-")
		resultRange := Range{}
		from, err := convertToInt(parts[0])
		if err != nil {
			return nil, err
		}

		resultRange.From = from
		to, err := convertToInt(parts[1])

		if err != nil {
			return nil, err
		}
		resultRange.To = to
		ranges = append(ranges, resultRange)
	}
	return ranges, nil
}

func convertToInt(part string) (int64, error) {
	if part != "" {
		result, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, err
		}
		return result, nil
	}
	return -1, nil
}
