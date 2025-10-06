package repo

import (
	"fmt"
)

type RequiredPackages map[string]*Package

func (requiredPackages RequiredPackages) Add(availablePackages AvailablePackages, pkgnames ...string) error {
	for _, pkgname := range pkgnames {
		pkg := availablePackages[pkgname]
		if pkg == nil {
			return fmt.Errorf("package %q not found", pkgname)
		}

		requiredPackages[pkgname] = pkg
	}

	return nil
}

func (requiredPackages RequiredPackages) AddDependencies(availablePackages AvailablePackages) error {
	for {
		var changed bool

		for _, pkg := range requiredPackages {
			var deps []Expression
			deps = append(deps, pkg.PreDepends...)
			deps = append(deps, pkg.Depends...)

			for _, dep := range deps {
				changed_, err := requiredPackages.addDependency(availablePackages, dep, false)
				if err != nil {
					return err
				}

				changed = changed || changed_
			}

			for _, dep := range pkg.Recommends {
				changed_, err := requiredPackages.addDependency(availablePackages, dep, true)
				if err != nil {
					return err
				}

				changed = changed || changed_
			}
		}

		if !changed {
			return nil
		}
	}
}

func (requiredPackages RequiredPackages) addDependency(availablePackages AvailablePackages, dep Expression, isRecommend bool) (bool, error) {
	pkg := dep.FindSatisfyingPackage(requiredPackages)
	if pkg != nil {
		return false, nil
	}

	pe, ok := dep.(*PackageExpression)
	if !ok {
		return false, fmt.Errorf("manual choice needed from %s", dep)
	}

	pkg = pe.FindSatisfyingPackage(availablePackages)
	if pkg == nil {
		if isRecommend {
			return false, nil
		}
		return false, fmt.Errorf("package matching %q not found", pe)
	}

	if err := requiredPackages.Add(availablePackages, pkg.Package); err != nil {
		return false, err
	}

	return true, nil
}
