package repo

import (
	"fmt"
	"regexp"
	"strings"
)

type Expression interface {
	FindSatisfyingPackage(map[string]*Package) *Package
	String() string
}

func parseExpressions(expr string) (exprs []Expression, _ error) {
	if expr == "" {
		return nil, nil
	}

	for s := range strings.SplitSeq(expr, ", ") {
		expr, err := parseOrExpression(s)
		if err != nil {
			return nil, err
		}

		exprs = append(exprs, expr)
	}

	return exprs, nil
}

type OrExpression []Expression

func parseOrExpression(expr string) (Expression, error) {
	var orExpr OrExpression
	for s := range strings.SplitSeq(expr, " | ") {
		expr, err := parsePackageExpression(s)
		if err != nil {
			return nil, err
		}

		orExpr = append(orExpr, expr)
	}

	if len(orExpr) == 1 {
		return orExpr[0], nil
	}

	return orExpr, nil
}

func (orExpr OrExpression) FindSatisfyingPackage(installed map[string]*Package) *Package {
	for _, expr := range orExpr {
		if pkg := expr.FindSatisfyingPackage(installed); pkg != nil {
			return pkg
		}
	}

	return nil
}

func (orExpr OrExpression) String() string {
	var exprs []string

	for _, expr := range orExpr {
		exprs = append(exprs, expr.String())
	}

	return strings.Join(exprs, " | ")
}

type PackageExpression struct {
	Package string
	Op      string
	Version *Version
}

var rxPackageExpression = regexp.MustCompile(`^([^()]+)(?: \((<<|<=|=|>=|>>) (.+)\))?$`)

func parsePackageExpression(expr string) (*PackageExpression, error) {
	m := rxPackageExpression.FindStringSubmatch(expr)
	if m == nil {
		return nil, fmt.Errorf("failed to match %q", expr)
	}

	var version *Version
	if m[2] != "" {
		var err error
		version, err = NewVersion(m[3])
		if err != nil {
			return nil, err
		}
	}

	return &PackageExpression{
		Package: strings.TrimSuffix(m[1], ":any"),
		Op:      m[2],
		Version: version,
	}, nil
}

func (pe *PackageExpression) FindSatisfyingPackage(installed map[string]*Package) *Package {
	if pkg, found := installed[pe.Package]; found && pe.isSatisfiedBy(&PackageExpression{Package: pkg.Package, Op: "=", Version: pkg.Version}) {
		return pkg
	}

	for _, pkg := range installed { // TODO: O(N)
		for _, expr := range pkg.Provides {
			if pe.isSatisfiedBy(expr.(*PackageExpression)) {
				return pkg
			}
		}
	}

	return nil
}

func (pe *PackageExpression) isSatisfiedBy(pe2 *PackageExpression) bool {
	if pe.Package != pe2.Package {
		return false
	}

	switch pe.Op {
	case "":
		return true
	case "<<":
		return pe.Version.Compare(pe2.Version) == 1
	case "<=":
		return pe.Version.Compare(pe2.Version) > -1
	case "=":
		return pe.Version.Compare(pe2.Version) == 0
	case ">=":
		return pe.Version.Compare(pe2.Version) < 1
	case ">>":
		return pe.Version.Compare(pe2.Version) == -1
	default:
		panic(fmt.Sprintf("invalid op %q", pe.Op))
	}
}

func (pe *PackageExpression) String() string {
	sb := &strings.Builder{}

	sb.WriteString(pe.Package)

	if pe.Op != "" {
		fmt.Fprintf(sb, " (%s %s)", pe.Op, pe.Version)
	}

	return sb.String()
}
