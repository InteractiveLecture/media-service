package main

import (
	"strconv"
	s "strings"
)

type Range struct {
	From int64
	To   int64
}

func (r Range) ContentLength() int64 {
	return (r.To - r.From) + 1
}

func NewRange(from, to, fileSize int64) Range {
	r := Range{from, to}
	if to < 0 {
		r.To = fileSize - 1
	}
	if from < 0 {
		r.From = 0
	}
	return r
}

func ParseRangeRequest(request string, fileSize int64) ([]Range, error) {
	a := s.Replace(request, " ", "", -1)
	a = s.Replace(a, "bytes=", "", -1)
	ranges := []Range{}
	for _, part := range s.Split(a, ",") {
		parts := s.Split(part, "-")
		from, err := convertToInt(parts[0])
		if err != nil {
			return nil, err
		}
		to, err := convertToInt(parts[1])
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, NewRange(from, to, fileSize))
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
