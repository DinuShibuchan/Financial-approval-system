package models

import (
	"testing"
)

func TestGetRequiredApprovalLevels(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected int
	}{
		{
			name:     "Under 10000 requires 1 level (Manager)",
			amount:   5000,
			expected: 1,
		},
		{
			name:     "Border under 10000 (9999.99) requires 1 level",
			amount:   9999.99,
			expected: 1,
		},
		{
			name:     "Exact 10000 requires 2 levels (Manager + Finance)",
			amount:   10000,
			expected: 2,
		},
		{
			name:     "Between 10000 and 50000 requires 2 levels",
			amount:   25000,
			expected: 2,
		},
		{
			name:     "Exact 50000 requires 2 levels",
			amount:   50000,
			expected: 2,
		},
		{
			name:     "Above 50000 requires 3 levels (Manager + Finance + Director)",
			amount:   50000.01,
			expected: 3,
		},
		{
			name:     "High value outlay (100000) requires 3 levels",
			amount:   100000,
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetRequiredApprovalLevels(tc.amount)
			if actual != tc.expected {
				t.Errorf("GetRequiredApprovalLevels(%.2f) = %d; want %d", tc.amount, actual, tc.expected)
			}
		})
	}
}

func TestGetApprovalLevelRole(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{level: 1, expected: RoleManager},
		{level: 2, expected: RoleFinance},
		{level: 3, expected: RoleDirector},
		{level: 0, expected: ""},
		{level: 4, expected: ""},
	}

	for _, tc := range tests {
		actual := GetApprovalLevelRole(tc.level)
		if actual != tc.expected {
			t.Errorf("GetApprovalLevelRole(%d) = %s; want %s", tc.level, actual, tc.expected)
		}
	}
}
