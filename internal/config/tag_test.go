package config

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMatchingTags(t *testing.T) {
	testSource := Source{
		Tags: []string{
			"",
			"foo",
			"bar",
			"1.0.0",
		},
	}

	var tests = []struct {
		tags []string
		want []string
	}{
		{[]string{}, []string{}},
		{[]string{""}, []string{}},
		{[]string{"1.0.0"}, []string{"1.0.0"}},
		{[]string{"foo", "bar"}, []string{"foo", "bar"}},
		{[]string{"latest"}, []string{}},
		{[]string{"latest", "1.0.0"}, []string{"1.0.0"}},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("matchingTags %v", test.tags)
		t.Run(testname, func(t *testing.T) {
			ans := testSource.matchingTags(test.tags)
			if !reflect.DeepEqual(ans, test.want) {
				t.Errorf("got %v, want %v", ans, test.want)
			}
		})
	}
}
