package main

import "testing"

func TestParseRange(t *testing.T) {
	testcase := "bytes = 0-100,800-,-900"
	ranges, err := ParseRangeRequest(testcase)
	if err != nil {
		t.Error(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("expected 3 ranges, but got %d", len(ranges))
	}
	expectedResults := []Range{Range{0, 100}, Range{800, -1}, Range{-1, 900}}
	if !compareAll(expectedResults, ranges) {
		t.Fatalf("expected %v, got %v", expectedResults, ranges)
	}

	testcase = "bytes=asdf"
	ranges, err = ParseRangeRequest(testcase)
	if err == nil {
		t.Fatal("expected an error, because range request is malformed")
	}

}

func compare(a Range, b Range) bool {
	return a.From == b.From && a.To == b.To
}

func compareAll(a []Range, b []Range) bool {
	for index, item := range a {
		if !compare(item, b[index]) {
			return false
		}
	}
	return true
}
