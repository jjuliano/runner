package resolver

import (
	"testing"
)

func TestDependencyResolver_ListDependencyTree(t *testing.T) {
	dr := &DependencyResolver{
		Packages: []PackageEntry{
			{Package: "a", Requires: []string{"b"}},
			{Package: "b", Requires: []string{"c"}},
			{Package: "c", Requires: []string{}},
		},
		packageDependencies: map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {},
		},
	}

	output := captureOutput(func() {
		dr.ListDependencyTree("a")
	})

	expected := "a > b > c\n"
	if output != expected {
		t.Errorf("Expected %s, but got %s", expected, output)
	}
}

func TestDependencyResolver_ListDependencyTreeTopDown(t *testing.T) {
	dr := &DependencyResolver{
		Packages: []PackageEntry{
			{Package: "a", Requires: []string{"b"}},
			{Package: "b", Requires: []string{"c"}},
			{Package: "c", Requires: []string{}},
		},
		packageDependencies: map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {},
		},
	}

	output := captureOutput(func() {
		dr.ListDependencyTreeTopDown("a")
	})

	expected := "c\nb\na\n"
	if output != expected {
		t.Errorf("Expected %s, but got %s", expected, output)
	}
}
