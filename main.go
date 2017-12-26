package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"text/template"

	"github.com/spf13/pflag"
)

var opts struct {
	Output       string
	InputDir     string
	TemplateFile string
}

func init() {
	pflag.StringVarP(&opts.InputDir, "input", "i", "changelog", "read input files from `dir`")
	pflag.StringVarP(&opts.Output, "output", "o", "", "write generated changelog to this `file` (default: print to stdout)")
	pflag.StringVarP(&opts.TemplateFile, "template", "t", "CHANGELOG.tmpl", "read template from `file` (relative to input directory)")
	pflag.Parse()
}

func die(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\\n") {
		msg += "\n"
	}
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

// V prints a debug message to stderr.
func V(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\\n") {
		msg += "\n"
	}
	fmt.Fprintf(os.Stderr, msg, args...)
}

// files lists all file names in dir. The file name is split by _, and the first component is used as the key in the resulting map.
func files(dir string) []string {
	d, err := os.Open(dir)
	if err != nil {
		die("error opening dir: %v", err)
	}

	names, err := d.Readdirnames(-1)
	if err != nil {
		_ = d.Close()
		die("error listing dir: %v", err)
	}

	err = d.Close()
	if err != nil {
		die("error closing dir: %v", err)
	}

	sort.Strings(names)

	var files []string
	for _, name := range names {
		// skip the template and versions file
		if name == "TEMPLATE" || name == "versions" {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}

	return files
}

// Release is one release, with an optional release date.
type Release struct {
	Version string
	Date    *time.Time
}

var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// readVersions reads the "versions" file in dir and returns the contents.
func readVersions(dir string) (result []Release) {
	data, err := ioutil.ReadFile(filepath.Join(dir, "versions"))
	if err != nil {
		die("unable to read file 'versions': %v", err)
	}

	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		// ignore comments
		if strings.HasPrefix(strings.TrimSpace(sc.Text()), "#") {
			continue
		}

		data := strings.SplitN(sc.Text(), " ", 2)
		ver := data[0]

		if !versionRegex.MatchString(ver) {
			die("version %q has wrong format", ver)
		}

		rel := Release{
			Version: ver,
		}

		if len(data) == 2 {
			t, err := time.Parse("2006-01-02", data[1])
			if err != nil {
				die("unable to parse date %q: %v", data[1], err)
			}
			rel.Date = &t
		}

		result = append(result, rel)
	}

	return result
}

// Entry describes a change.
type Entry struct {
	Type      string
	TypeShort string
	Title     string
	Text      string
	URLs      []*url.URL
	Issues    []string
	PRs       []string
	PrimaryID string
}

// EntryTypePriority contains the list of valid types, order is priority in the changelog.
var EntryTypePriority = map[string]int{
	"Security":    1,
	"Bugfix":      2,
	"Change":      3,
	"Enhancement": 4,
}

// EntryTypeAbbreviation contains the shortened entry types for the overview.
var EntryTypeAbbreviation = map[string]string{
	"Security":    "Sec",
	"Bugfix":      "Fix",
	"Change":      "Chg",
	"Enhancement": "Enh",
}

// Valid returns an error if the entry is invalid in any way.
func (e Entry) Valid() error {
	if e.Type == "" {
		return errors.New("entry title does not have a prefix, example: Bugfix: restore old behavior")
	}

	if _, ok := EntryTypePriority[e.Type]; !ok {
		return fmt.Errorf("entry type %q is invalid, valid types: %v", e.Type, EntryTypePriority)
	}

	if len(e.Type)+len(e.Title)+1 > 80 {
		return errors.New("title is too long")
	}

	return nil
}

func readFile(filename string) (e Entry) {
	f, err := os.Open(filename)
	if err != nil {
		die("unable to open %v: %v", filename, err)
	}

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		die("unable to read first line from %v", filename)
	}

	title := sc.Text()
	data := strings.SplitN(title, ": ", 2)
	if len(data) == 2 {
		e.Type = capitalize(data[0])
		e.TypeShort = EntryTypeAbbreviation[e.Type]
		data = data[1:]
	}
	e.Title = capitalize(data[0])

	var text []string
	var sect string
	for sc.Scan() {
		if sc.Err() != nil {
			die("unable to read lines from %v: %v", filename, sc.Err())
		}

		if strings.TrimSpace(sc.Text()) == "" {
			if sect != "" {
				text = append(text, sect)
			}

			sect = ""
			continue
		}

		if sect != "" {
			sect += " "
		}
		sect += strings.TrimSpace(sc.Text())
	}

	err = f.Close()
	if err != nil {
		die("error closing %v: %v", filename, err)
	}

	if sect != "" {
		text = append(text, sect)
	}

	if len(text) > 0 {
		links := text[len(text)-1]
		text = text[:len(text)-1]

		sc = bufio.NewScanner(strings.NewReader(links))
		sc.Split(bufio.ScanWords)
		for sc.Scan() {
			url, err := url.Parse(sc.Text())
			if err != nil {
				die("file %v: unable to parse url %q: %v", filename, sc.Text(), err)
			}
			e.URLs = append(e.URLs, url)
		}
	}

	e.Text = capitalize(strings.TrimSpace(strings.Join(text, "\n\n")))

	e.Issues, e.PRs = githubIDs(e.URLs)

	if len(e.Issues) > 0 {
		e.PrimaryID = e.Issues[0]
	} else if len(e.PRs) > 0 {
		e.PrimaryID = e.PRs[0]
	}

	err = e.Valid()
	if err != nil {
		die("file %v: %v", filename, err)
	}

	return e
}

const issuePath = "/restic/restic/issues/"
const pullRequestPath = "/restic/restic/pull/"

// githubIDs extracts all issue and pull request IDs from the urls.
func githubIDs(urls []*url.URL) (issues, prs []string) {
	for _, url := range urls {
		if url.Host != "github.com" {
			continue
		}

		if strings.HasPrefix(url.Path, issuePath) {
			issues = append(issues, url.Path[len(issuePath):])
		}

		if strings.HasPrefix(url.Path, pullRequestPath) {
			prs = append(prs, url.Path[len(pullRequestPath):])
		}
	}

	return issues, prs
}

func readEntries(dir string, versions []Release) (entries map[string][]Entry) {
	entries = make(map[string][]Entry)

	for _, ver := range versions {
		for _, file := range files(filepath.Join(dir, ver.Version)) {
			entries[ver.Version] = append(entries[ver.Version], readFile(file))
		}
	}

	// sort all entries according to priority, otherwise leave the original ordering
	for ver, list := range entries {
		sort.SliceStable(list, func(i, j int) bool {
			return EntryTypePriority[list[i].Type] < EntryTypePriority[list[j].Type]
		})
		entries[ver] = list
	}

	return entries
}

// wrapText formats the text in a column smaller than width characters,
// indenting each new line with indent spaces.
func wrapText(text string, width, indent int) (result string, err error) {
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Split(bufio.ScanWords)
	cl := 0
	for sc.Scan() {
		if sc.Err() != nil {
			return "", sc.Err()
		}

		if cl+len(sc.Text()) > width {
			result += "\n"
			result += strings.Repeat(" ", indent)
			cl = 0
		}

		if cl > 0 {
			result += " "
		}
		result += sc.Text()
		cl += len(sc.Text())
	}

	return result, nil
}

// capitalize returns a string with the first letter in upper case.
func capitalize(text string) string {
	if text == "" {
		return text
	}

	first, rest := text[0:1], text[1:]
	return strings.ToUpper(first) + rest
}

var helperFuncs = template.FuncMap{
	"wrap":       wrapText,
	"capitalize": capitalize,
}

func main() {
	if !filepath.IsAbs(opts.TemplateFile) {
		opts.TemplateFile = filepath.Join(opts.InputDir, opts.TemplateFile)
	}

	buf, err := ioutil.ReadFile(opts.TemplateFile)
	if err != nil {
		die("unable to read template from %v: %v", opts.TemplateFile, err)
	}

	templ, err := template.New("").Funcs(helperFuncs).Parse(string(buf))
	if err != nil {
		die("unable to compile template: %v", err)
	}

	type VersionChanges struct {
		Version string
		Date    string
		Entries []Entry
	}

	versions := readVersions(opts.InputDir)

	var changes []VersionChanges

	all := readEntries(opts.InputDir, versions)
	for _, ver := range versions {
		vc := VersionChanges{
			Version: ver.Version,
			Entries: all[ver.Version],
		}

		if ver.Date != nil {
			vc.Date = ver.Date.Format("2006-01-02")
		} else {
			vc.Date = "UNRELEASED"
		}

		changes = append(changes, vc)
	}

	wr := os.Stdout

	if opts.Output != "" {
		wr, err = os.Create(opts.Output)
		if err != nil {
			die("unable to create file %v: %v", opts.Output, err)
		}
	}

	err = templ.Execute(wr, changes)
	if err != nil {
		die("error executing template: %v", err)
	}

	if opts.Output != "" {
		err = wr.Close()
		if err != nil {
			die("error closing file %v: %v", opts.Output, err)
		}
	}
}
