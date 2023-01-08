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
			"Bugfix: subject line\n\nhttps://github.com/restic/restic/issues/12345",
			Entry{
				Title:      "Subject line",
				Type:       "Bugfix",
				TypeShort:  "Fix",
				PrimaryID:  12345,
				PrimaryURL: parseURL(t, "https://github.com/restic/restic/issues/12345"),
				URLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
				},
				Issues: []string{"12345"},
				IssueURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
				},
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
				PrimaryID:  12345,
				PrimaryURL: parseURL(t, "https://github.com/restic/restic/issues/12345"),
				Issues:     []string{"12345"},
				IssueURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
				},
				PRs: []string{"666666"},
				PRURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/pull/666666"),
				},
			},
		},
		{
			`Enhancement: foo bar subject

https://github.com/restic/restic/issues/12345
https://github.com/restic/rest-server/issues/232323
https://github.com/restic/restic/pull/666666
https://forum.restic.net/t/getting-last-successful-backup-time/531
`,
			Entry{
				Title:      "Foo bar subject",
				Type:       "Enhancement",
				TypeShort:  "Enh",
				PrimaryID:  12345,
				PrimaryURL: parseURL(t, "https://github.com/restic/restic/issues/12345"),
				Issues:     []string{"12345", "232323"},
				IssueURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
					parseURL(t, "https://github.com/restic/rest-server/issues/232323"),
				},
				PRs: []string{"666666"},
				PRURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/pull/666666"),
				},
				URLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
					parseURL(t, "https://github.com/restic/rest-server/issues/232323"),
					parseURL(t, "https://github.com/restic/restic/pull/666666"),
					parseURL(t, "https://forum.restic.net/t/getting-last-successful-backup-time/531"),
				},
				OtherURLs: []*url.URL{
					parseURL(t, "https://forum.restic.net/t/getting-last-successful-backup-time/531"),
				},
			},
		},
		{
			"Security: short and terse summary\n\n```\nexample\n   with\n       random spaces\n```\n\nLast block contains just\na few\nlinks.\n\nhttps://github.com/restic/restic/issues/12345",
			Entry{
				Title:     "Short and terse summary",
				Type:      "Security",
				TypeShort: "Sec",
				Paragraphs: []string{
					"```\nexample\n   with\n       random spaces\n```",
					"Last block contains just a few links.",
				},
				URLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
				},
				PrimaryID:  12345,
				PrimaryURL: parseURL(t, "https://github.com/restic/restic/issues/12345"),
				Issues:     []string{"12345"},
				IssueURLs: []*url.URL{
					parseURL(t, "https://github.com/restic/restic/issues/12345"),
				},
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
