package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func parseURL(t testing.TB, s string) *url.URL {
	url, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return url
}

func ptrTime(v time.Time) *time.Time {
	return &v
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

` + "```bash\necho 'test code block with type'\n```" + `

https://github.com/restic/restic/issues/12345
https://github.com/restic/rest-server/issues/232323
https://github.com/restic/restic/pull/666666
https://forum.restic.net/t/getting-last-successful-backup-time/531
`,
			Entry{
				Title:      "Foo bar subject",
				Paragraphs: []string{"```bash\necho 'test code block with type'\n```"},
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
func TestReadReleases(t *testing.T) {
	type testData struct {
		Date       *time.Time
		FolderName string
		Version    string
	}
	dir := t.TempDir()
	releases := []testData{
		{Date: nil, FolderName: "unreleased", Version: "unreleased"},
		{Date: ptrTime(time.Date(2023, time.November, 12, 0, 0, 0, 0, time.UTC)), FolderName: "2.0.0-rc.1+build.12345_2023-11-12", Version: "2.0.0-rc.1+build.12345"},
		{Date: ptrTime(time.Date(2023, time.November, 10, 0, 0, 0, 0, time.UTC)), FolderName: "0.0.1-rc.1_2023-11-10", Version: "0.0.1-rc.1"},
		{Date: ptrTime(time.Date(2023, time.November, 10, 0, 0, 0, 0, time.UTC)), FolderName: "1.0.1_2023-11-10", Version: "1.0.1"},
		{Date: ptrTime(time.Date(2023, time.November, 9, 0, 0, 0, 0, time.UTC)), FolderName: "4.0.0_2023-11-09", Version: "4.0.0"},
		{Date: ptrTime(time.Date(2023, time.November, 8, 0, 0, 0, 0, time.UTC)), FolderName: "1.0.2-alpha.10_2023-11-08", Version: "1.0.2-alpha.10"},
		{Date: ptrTime(time.Date(2023, time.September, 7, 0, 0, 0, 0, time.UTC)), FolderName: "1.0.0_2023-09-07", Version: "1.0.0"},
		{Date: ptrTime(time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC)), FolderName: "12.10.21_2023-05-01", Version: "12.10.21"},
	}
	for _, release := range releases {
		err := os.Mkdir(filepath.Join(dir, release.FolderName), 0750)
		if err != nil {
			t.Fatal(err)
		}
	}
	parsedReleases := readReleases(dir)
	// test the sorting and the parsing of the folder names
	for i, parsedRelease := range parsedReleases {
		if ((releases[i].Date == nil || parsedRelease.Date == nil) && releases[i].Date != parsedRelease.Date) || (releases[i].Date != nil && !releases[i].Date.Equal(*parsedRelease.Date)) {
			t.Fatalf("date mismatch, expected %v, got %v", releases[i].Date, parsedRelease.Date)
		}
		if releases[i].Version != parsedRelease.Version {
			t.Fatalf("version mismatch, expected %v, got %v", releases[i].Version, parsedRelease.Version)
		}
	}
}

func TestWrapIndent(t *testing.T) {
	var tests = []struct {
		In     string
		Width  int
		Indent int
		Out    string
	}{
		{"Example string", 80, 4, "Example string"},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 70, 3, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do\n   eiusmod tempor incididunt ut labore et dolore magna aliqua."},
		{"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", 55, 2, "Lorem ipsum dolor sit amet, consectetur adipiscing\n  elit, sed do eiusmod tempor incididunt ut labore et\n  dolore magna aliqua."},
		{"```\nexample\n   with\n       random spaces\n```", 10, 3, "```\n   example\n      with\n          random spaces\n   ```"},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			res, err := wrapIndent(test.In, test.Width, test.Indent)
			if err != nil {
				t.Fatal(err)
			}

			if diff := deep.Equal(res, test.Out); diff != nil {
				t.Error(diff)
			}
		})
	}
}
