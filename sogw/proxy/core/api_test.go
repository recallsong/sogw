package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRewriteURLs = map[string][]string{
	"/abc/def/abc?defg": []string{"/abc/def/abc?defg"},
	"/{abc}/:def":       []string{"/", ":abc", "/", ":def"},
	"/{abc}/{def}":      []string{"/", ":abc", "/", ":def"},
	"/:abc/{def}":       []string{"/", ":abc", "/", ":def"},
	"/:abc/:def":        []string{"/", ":abc", "/", ":def"},

	"/:abc/:def?:abc=:def":        []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def"},
	"/:abc/:def?{abc}={def}":      []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def"},
	"/:abc/:def?{abc}={def}{":     []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", "{"},
	"/:abc/:def?{abc}={def}:":     []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", ":"},
	"/:abc/:def?{abc}={def}*":     []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", "*"},
	"/:abc/:def?{abc}={def}*xyz":  []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", "*", "xyz"},
	"/:abc/:def?{abc}={def}*?":    []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", "*", "?"},
	"/:abc/:def?{abc}={def}*?xyz": []string{"/", ":abc", "/", ":def", "?", ":abc", "=", ":def", "*", "?xyz"},
	"/:def?{abc}={def}*?xyz":      []string{"/", ":def", "?", ":abc", "=", ":def", "*", "?xyz"},
	"/{def}?{abc}={def}*?xyz":     []string{"/", ":def", "?", ":abc", "=", ":def", "*", "?xyz"},
	"/:def?{abc}={def}$*xyz":      []string{"/", ":def", "?", ":abc", "=", ":def", "$*", "xyz"},
	"/{def}?{abc}={def}$xyz":      []string{"/", ":def", "?", ":abc", "=", ":def", "yz"},
	"/{def}?{abc}=$xyz{def}$":     []string{"/", ":def", "?", ":abc", "=", "yz", ":def"},
	"/{def}$??{abc}=$xyz{def}$":   []string{"/", ":def", "$?", "?", ":abc", "=", "yz", ":def"},
}

func TestURLRewrite(t *testing.T) {
	for path, expect := range testRewriteURLs {
		parts := makeURLRewrite(path)
		if !assert.Equal(t, len(expect), len(parts), fmt.Sprint(path, " ", strings.Join(parts, " "))) {
			return
		}
		for i, p := range parts {
			if !assert.Equal(t, expect[i], p, fmt.Sprint(path, " ", strings.Join(parts, " "))) {
				return
			}
		}
	}
}
