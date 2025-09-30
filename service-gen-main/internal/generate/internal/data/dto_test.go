package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Name  string
		Input []interface{}
		Want  []enumNormalized
		Err   error
	}{
		{
			Name:  "basic",
			Input: []interface{}{"ENUM_ONE", "ENUM_TWO"},
			Want: []enumNormalized{
				{
					Value:          "ENUM_ONE",
					NameNormalized: "ENUM_ONE",
				},
				{
					Value:          "ENUM_TWO",
					NameNormalized: "ENUM_TWO",
				},
			},
		},
		{
			Name:  "starts with number",
			Input: []interface{}{"1_SOMETHING", "2_OTHER"},
			Want: []enumNormalized{
				{
					Value:          "1_SOMETHING",
					NameNormalized: "ENUM_1_SOMETHING",
				},
				{
					Value:          "2_OTHER",
					NameNormalized: "ENUM_2_OTHER",
				},
			},
		},
		{
			Name:  "contains dots",
			Input: []interface{}{"MY.SOMETHING", "HIS.OTHER"},
			Want: []enumNormalized{
				{
					Value:          "MY.SOMETHING",
					NameNormalized: "MY_SOMETHING",
				},
				{
					Value:          "HIS.OTHER",
					NameNormalized: "HIS_OTHER",
				},
			},
		},
		{
			Name:  "numbers and dots",
			Input: []interface{}{"1.3.158.00165387.100.40.50", "1.3.158.00165387.100.40.60"},
			Want: []enumNormalized{
				{
					Value:          "1.3.158.00165387.100.40.50",
					NameNormalized: "ENUM_1_3_158_00165387_100_40_50",
				},
				{
					Value:          "1.3.158.00165387.100.40.60",
					NameNormalized: "ENUM_1_3_158_00165387_100_40_60",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			result, err := processEnums(test.Input)
			assert.Equal(t, test.Want, result)
			assert.Equal(t, test.Err, err)
		})
	}
}
