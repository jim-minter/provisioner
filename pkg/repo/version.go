package repo

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// https://www.debian.org/doc/debian-policy/ch-controlfields.html#version

type Version struct {
	Epoch           uint64
	UpstreamVersion string
	DebianRevision  string
}

func NewVersion(s string) (*Version, error) {
	v := &Version{}

	if i := strings.IndexByte(s, ':'); i != -1 {
		var epoch string
		epoch, s = s[:i], s[i+1:]

		var err error
		v.Epoch, err = strconv.ParseUint(epoch, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	if i := strings.LastIndexByte(s, '-'); i != -1 {
		s, v.DebianRevision = s[:i], s[i+1:]

		if strings.ContainsFunc(v.DebianRevision, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '+' || r == '.' || r == '~')
		}) {
			return nil, fmt.Errorf("invalid debian revision %q", v.DebianRevision)
		}
	}

	v.UpstreamVersion = s

	if strings.ContainsFunc(v.UpstreamVersion, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '+' || r == '-' || r == '~')
	}) {
		return nil, fmt.Errorf("invalid upstream version %q", v.UpstreamVersion)
	}

	return v, nil
}

func (v *Version) String() string {
	sb := &strings.Builder{}

	if v.Epoch != 0 {
		fmt.Fprintf(sb, "%d:", v.Epoch)
	}

	sb.WriteString(v.UpstreamVersion)

	if v.DebianRevision != "" {
		fmt.Fprintf(sb, "-%s", v.DebianRevision)
	}

	return sb.String()
}

func (v *Version) Compare(v2 *Version) int {
	if rv := cmp.Compare(v.Epoch, v2.Epoch); rv != 0 {
		return rv
	}

	if rv := compareString(v.UpstreamVersion, v2.UpstreamVersion); rv != 0 {
		return rv
	}

	return compareString(orZero(v.DebianRevision), orZero(v2.DebianRevision))
}

func compareString(a, b string) int {
	for a != "" || b != "" {
		var aprefix, bprefix string
		aprefix, a = cutMatchingPrefix(a, func(r rune) bool { return !unicode.IsNumber(r) })
		bprefix, b = cutMatchingPrefix(b, func(r rune) bool { return !unicode.IsNumber(r) })

		if rv := compareLetters(aprefix, bprefix); rv != 0 {
			return rv
		}
		aprefix, a = cutMatchingPrefix(a, unicode.IsNumber)
		bprefix, b = cutMatchingPrefix(b, unicode.IsNumber)
		aprefix, bprefix = orZero(aprefix), orZero(bprefix)

		ai, err := strconv.ParseUint(aprefix, 10, 64)
		if err != nil {
			panic(err)
		}

		bi, err := strconv.ParseUint(bprefix, 10, 64)
		if err != nil {
			panic(err)
		}

		if rv := cmp.Compare(ai, bi); rv != 0 {
			return rv
		}
	}

	return 0
}

func compareLetters(a, b string) int {
	for i := 0; i < len(a) || i < len(b); i++ {
		var ai, bi byte
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		if rv := cmp.Compare(order(ai), order(bi)); rv != 0 {
			return rv
		}
	}

	return 0
}

func order(b byte) int {
	switch {
	case b == '~':
		return -1
	case b == 0:
		return 0
	case b >= 'A' && b <= 'Z':
		return int(b - 'A' + 1)
	case b >= 'a' && b <= 'z':
		return int(b - 'a' + 27)
	case b == '+':
		return 53
	case b == '-':
		return 54
	case b == '.':
		return 55
	default:
		panic("invalid argument")
	}
}

func cutMatchingPrefix(s string, f func(rune) bool) (string, string) {
	if i := strings.IndexFunc(s, func(r rune) bool { return !f(r) }); i != -1 {
		return s[:i], s[i:]
	}
	return s, ""
}

func orZero(s string) string {
	if s != "" {
		return s
	}

	return "0"
}
