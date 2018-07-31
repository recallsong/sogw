package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testSwaggerURLs = map[string]string{
	"/abc/def/abc?defg":     "/abc/def/abc?defg",
	"/{abc}/def":            "/:abc/def",
	"/{abc}/{def}":          "/:abc/:def",
	"/abc/{def}":            "/abc/:def",
	"/abc/def?{abc}={def}":  "/abc/def?:abc=:def",
	"/abc/def?{abc}={def}{": "/abc/def?:abc=:def{",
}

func TestPathToRoute(t *testing.T) {
	for path, expect := range testSwaggerURLs {
		spart, err := pathToRoute(path)
		if !assert.Nil(t, err) {
			return
		}
		if !assert.Equal(t, expect, spart) {
			return
		}
	}
}
