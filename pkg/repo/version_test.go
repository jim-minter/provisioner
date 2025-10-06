package repo

import (
	"fmt"
	"strconv"
	"testing"
	"unicode"

	"github.com/onsi/gomega"
)

func TestNewVersion(t *testing.T) {
	t.Parallel()

	for i, tt := range []struct {
		s     string
		check func(*gomega.WithT, *Version, error)
	}{
		{
			s: "v1.0.0",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(v).To(gomega.Equal(&Version{UpstreamVersion: "v1.0.0"}))
			},
		},
		{
			s: "v1.0.0-debian1",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(v).To(gomega.Equal(&Version{UpstreamVersion: "v1.0.0", DebianRevision: "debian1"}))
			},
		},
		{
			s: "v1.0.0-alpha1-debian1",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(v).To(gomega.Equal(&Version{UpstreamVersion: "v1.0.0-alpha1", DebianRevision: "debian1"}))
			},
		},
		{
			s: "1:v1.0.0-alpha1-debian1",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(v).To(gomega.Equal(&Version{Epoch: 1, UpstreamVersion: "v1.0.0-alpha1", DebianRevision: "debian1"}))
			},
		},
		{
			s: "bad:v1.0.0-alpha1-debian1",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).To(gomega.BeAssignableToTypeOf(&strconv.NumError{}))
				g.Expect(v).To(gomega.BeNil())
			},
		},
		{
			s: "1:v1.0.0-alpha1-debian1!",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).To(gomega.MatchError(`invalid debian revision "debian1!"`))
				g.Expect(v).To(gomega.BeNil())
			},
		},
		{
			s: "1:v1.0.0-alpha1!-debian1",
			check: func(g *gomega.WithT, v *Version, err error) {
				g.Expect(err).To(gomega.MatchError(`invalid upstream version "v1.0.0-alpha1!"`))
				g.Expect(v).To(gomega.BeNil())
			},
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			v, err := NewVersion(tt.s)
			tt.check(g, v, err)

			if v != nil {
				g.Expect(v.String()).To(gomega.Equal(tt.s))
			}
		})
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	g.Expect((&Version{}).Compare(&Version{Epoch: 1})).To(gomega.Equal(-1))
	g.Expect((&Version{}).Compare(&Version{UpstreamVersion: "1.0"})).To(gomega.Equal(-1))
	g.Expect((&Version{}).Compare(&Version{DebianRevision: "1.0"})).To(gomega.Equal(-1))
}

func TestCompareString(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	g.Expect(compareString("1.0~beta1~svn1245", "1.0~beta1")).To(gomega.Equal(-1))
	g.Expect(compareString("1.0~beta1", "1.0")).To(gomega.Equal(-1))

	g.Expect(compareString("1.0", "1.1")).To(gomega.Equal(-1))
	g.Expect(compareString("1.0", "1.0")).To(gomega.Equal(0))

	g.Expect(func() { compareString("18446744073709551616", "") }).To(gomega.Panic())
	g.Expect(func() { compareString("", "18446744073709551616") }).To(gomega.Panic())
}

func TestCompareLetters(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	g.Expect(compareLetters("~~", "~~a")).To(gomega.Equal(-1))
	g.Expect(compareLetters("~~a", "~")).To(gomega.Equal(-1))
	g.Expect(compareLetters("~", "")).To(gomega.Equal(-1))
	g.Expect(compareLetters("", "a")).To(gomega.Equal(-1))
}

func TestOrder(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	g.Expect(order('~')).To(gomega.Equal(-1))
	g.Expect(order(0)).To(gomega.Equal(0))
	g.Expect(order('B')).To(gomega.Equal(2))
	g.Expect(order('b')).To(gomega.Equal(28))
	g.Expect(order('+')).To(gomega.Equal(53))
	g.Expect(order('-')).To(gomega.Equal(54))
	g.Expect(order('.')).To(gomega.Equal(55))
	g.Expect(func() { order('$') }).To(gomega.PanicWith("invalid argument"))
}

func TestCutMatchingPrefix(t *testing.T) {
	t.Parallel()

	for i, tt := range []struct {
		s     string
		f     func(rune) bool
		check func(*gomega.WithT, string, string)
	}{
		{
			s: "123",
			f: unicode.IsNumber,
			check: func(g *gomega.WithT, prefix, s string) {
				g.Expect(prefix).To(gomega.Equal("123"))
				g.Expect(s).To(gomega.Equal(""))
			},
		},
		{
			s: "123abc",
			f: unicode.IsNumber,
			check: func(g *gomega.WithT, prefix, s string) {
				g.Expect(prefix).To(gomega.Equal("123"))
				g.Expect(s).To(gomega.Equal("abc"))
			},
		},
		{
			s: "123abc",
			f: func(r rune) bool { return !unicode.IsNumber(r) },
			check: func(g *gomega.WithT, prefix, s string) {
				g.Expect(prefix).To(gomega.Equal(""))
				g.Expect(s).To(gomega.Equal("123abc"))
			},
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			g := gomega.NewWithT(t)

			prefix, s := cutMatchingPrefix(tt.s, tt.f)
			tt.check(g, prefix, s)
		})
	}
}

func TestOrZero(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	g.Expect(orZero("s")).To(gomega.Equal("s"))
	g.Expect(orZero("")).To(gomega.Equal("0"))
}
