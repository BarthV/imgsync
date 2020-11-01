package config

import (
	"fmt"
	"testing"
)

func TestGetConfigLocation(t *testing.T) {
	var tests = []struct {
		location string
		want     string
	}{
		{"", ".imgsync.yaml"},
		{"./", ".imgsync.yaml"},
		{"/", "/.imgsync.yaml"},
		{"test.yaml", "test.yaml"},
		{"test.yml", "test.yml"},
		{"test", "test/.imgsync.yaml"},
		{"/../test.yaml", "/../test.yaml"},
		{"foo.bar", "foo.bar/.imgsync.yaml"},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("location \"%s\"", test.location)
		t.Run(testname, func(t *testing.T) {
			ans := getConfigLocation(test.location)
			if ans != test.want {
				t.Errorf("got '%s', want '%s'", ans, test.want)
			}
		})
	}
}

func TestGetRepositoryAddress(t *testing.T) {
	var tests = []struct {
		repo Repo
		want string
	}{
		{
			Repo{Repository: "barthv/imgsync"},
			"index.docker.io/barthv/imgsync",
		},
		{
			Repo{Repository: "busybox"},
			"index.docker.io/busybox",
		},
		{
			Repo{
				Repository: "google-containers/busybox",
				Host:       "gcr.io",
			},
			"gcr.io/google-containers/busybox",
		},
		{
			Repo{
				Repository: "barthv/imgsync",
				Host:       "192.168.42.69:5000",
			},
			"192.168.42.69:5000/barthv/imgsync",
		},
		{
			Repo{
				Repository: "barthv/foo/bar",
				Host:       "127.0.0.1:5000",
			},
			"127.0.0.1:5000/barthv/foo/bar",
		},
	}

	for _, test := range tests {
		testname := fmt.Sprintf("%s-%s", test.repo.Repository, test.repo.Host)
		t.Run(testname, func(t *testing.T) {
			ans := test.repo.GetRepositoryAddress()
			if ans != test.want {
				t.Errorf("got '%s', want '%s'", ans, test.want)
			}
		})
	}
}
