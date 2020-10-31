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

func TestMatchingRegexTags(t *testing.T) {
	tags := []string{
		"",
		"latest",
		"foo",
		"bar",
		"1.0.0",
		"1.1.0",
		"1.2.0",
		"1.2.1",
		"3.2.1",
		"3.2.1-arm64",
		"3.2.1-alpine",
	}

	var tests = []struct {
		source    Source
		want      []string
		wantError bool
	}{
		{Source{RegexTags: []string{}}, []string{}, false},
		{Source{RegexTags: []string{"?)(broken{&invalid\\d[reg|exp"}}, []string{}, true},
		{Source{RegexTags: []string{""}}, []string{}, false},
		{Source{RegexTags: []string{".*"}}, tags, false},
		{Source{RegexTags: []string{"2\\..+"}}, []string{"1.2.0", "1.2.1", "3.2.1", "3.2.1-arm64", "3.2.1-alpine"}, false},
		{Source{RegexTags: []string{"^1\\..+$"}}, []string{"1.0.0", "1.1.0", "1.2.0", "1.2.1"}, false},
		{Source{RegexTags: []string{"^1\\.2\\.[0-9]+"}}, []string{"1.2.0", "1.2.1"}, false},
		{Source{RegexTags: []string{"^3\\.[0-9]+\\.[0-9]+$"}}, []string{"3.2.1"}, false},
		{Source{RegexTags: []string{"notfound"}}, []string{}, false},
		{Source{RegexTags: []string{"^(foo|bar)$"}}, []string{"foo", "bar"}, false},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("matchingRegexTags %v", test.source.RegexTags)
		t.Run(testname, func(t *testing.T) {
			ans, err := test.source.matchingRegexTags(tags)
			if err != nil && !test.wantError {
				t.Errorf("got unexpected error %v", err)
			}
			if !reflect.DeepEqual(ans, test.want) {
				t.Errorf("got %v, want %v", ans, test.want)
			}
		})
	}

}
