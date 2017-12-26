package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/go-test/deep"
)

func parseURL(t testing.TB, s string) *url.URL {
	url, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return url
}

func TestReadFile(t *testing.T) {
	var tests = []struct {
		Data string
		Entry
	}{
		{
			"Bugfix: subject line",
			Entry{
				Title:     "Subject line",
				Type:      "Bugfix",
				TypeShort: "Fix",
			},
		},
		{
			`Security: short and terse summary

A block of text. Lorem ipsum or so. May
wrap around,
arbitrarily.

second block of text.
may also contain
many different lines.

Last block contains just
a few
links.

https://github.com/restic/restic/issues/12345
https://github.com/restic/restic/pull/666666
`,
			Entry{
				Title:     "Short and terse summary",
				Type:      "Security",
				TypeShort: "Sec",
				Paragraphs: []string{
					"A block of text. Lorem ipsum or so. May wrap around, arbitrarily.",
					"Second block of text. may also contain many different lines.",
					"Last block contains just a few links.",
				},
				URLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
					parseURL(t, "https://github.com/restic/restic/pull/666666"),
				},
				PrimaryID: "12345",
				Issues:    []string{"12345"},
				PRs:       []string{"666666"},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			f, err := ioutil.TempFile("", "calens-test-")
			if err != nil {
				t.Fatal(err)
			}

			defer os.Remove(f.Name())

			_, err = f.Write([]byte(test.Data))
			if err != nil {
				t.Fatal(err)
			}

			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}

			entry := readFile(f.Name())
			if diff := deep.Equal(test.Entry, entry); diff != nil {
				t.Error(diff)
			}
		})
	}

}
