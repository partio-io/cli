package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// prRef identifies one GitHub pull request.
type prRef struct {
	Owner  string
	Repo   string
	Number int
}

func (r prRef) FullName() string { return r.Owner + "/" + r.Repo }

func (r prRef) String() string { return fmt.Sprintf("%s/%s#%d", r.Owner, r.Repo, r.Number) }

var (
	prNamePattern  = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	prShortPattern = regexp.MustCompile(`^([^/\s#]+)/([^/\s#]+)#([0-9]+)$`)
)

const prRefUsage = "expected a pull request URL (https://github.com/<owner>/<repo>/pull/<n>) or <owner>/<repo>#<n>"

// parsePRRef parses a pull request URL or an owner/repo#number reference.
// Owner, repo and number come only from the URL path; the host is ignored.
func parsePRRef(s string) (prRef, error) {
	s = strings.TrimSpace(s)
	fail := func() (prRef, error) {
		return prRef{}, fmt.Errorf("invalid pull request reference %q: %s", s, prRefUsage)
	}

	var owner, repo, num string
	if m := prShortPattern.FindStringSubmatch(s); m != nil {
		owner, repo, num = m[1], m[2], m[3]
	} else {
		u, err := url.Parse(s)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return fail()
		}
		segs := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(segs) < 4 || segs[2] != "pull" {
			return fail()
		}
		owner, repo, num = segs[0], segs[1], segs[3]
	}

	if !prNamePattern.MatchString(owner) || !prNamePattern.MatchString(repo) {
		return fail()
	}
	n, err := strconv.Atoi(num)
	if err != nil || n <= 0 {
		return fail()
	}
	return prRef{Owner: owner, Repo: repo, Number: n}, nil
}
