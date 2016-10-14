package solparse

import "testing"

func TestIsElementaryTypeName(t *testing.T) {
	var testCases = []struct {
		t string
		v bool
	}{
		{"bool", true},
		{"foo", false},
		{"int", true},
		{"uint256", true},
		{"int57", false},
		{"bytes1", true},
		{"bytes32", true},
		{"bytes33", false},
	}

	for _, tc := range testCases {
		if isElementaryTypeName(tc.t) && !tc.v {
			t.Errorf("Expected %s to elementary type", tc.t)
		} else if !isElementaryTypeName(tc.t) && tc.v {
			t.Errorf("Expected %s to not be elementary type", tc.t)
		}
	}
}
