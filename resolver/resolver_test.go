package resolver

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// Helper function to initialize a DependencyResolver with sample data.
func setupTestResolver() *DependencyResolver {
	fs := afero.NewMemMapFs() // Using an in-memory filesystem for testing
	resolver := NewDependencyResolver(fs)
	resolver.Packages = []PackageEntry{
		{Package: "a", Name: "A", Sdesc: "Package A", Ldesc: "The first package in the alphabetical order", Category: "example", Requires: []string{}},
		{Package: "b", Name: "B", Sdesc: "Package B", Ldesc: "The second package, dependent on A", Category: "example", Requires: []string{"a"}},
		{Package: "c", Name: "C", Sdesc: "Package C", Ldesc: "The third package, dependent on B", Category: "example", Requires: []string{"b"}},
		{Package: "d", Name: "D", Sdesc: "Package D", Ldesc: "The fourth package, dependent on C", Category: "example", Requires: []string{"c"}},
		{Package: "e", Name: "E", Sdesc: "Package E", Ldesc: "The fifth package, dependent on D", Category: "example", Requires: []string{"d"}},
		{Package: "f", Name: "F", Sdesc: "Package F", Ldesc: "The sixth package, dependent on E", Category: "example", Requires: []string{"e"}},
		{Package: "g", Name: "G", Sdesc: "Package G", Ldesc: "The seventh package, dependent on F", Category: "example", Requires: []string{"f"}},
		{Package: "h", Name: "H", Sdesc: "Package H", Ldesc: "The eighth package, dependent on G", Category: "example", Requires: []string{"g"}},
		{Package: "i", Name: "I", Sdesc: "Package I", Ldesc: "The ninth package, dependent on H", Category: "example", Requires: []string{"h"}},
		{Package: "j", Name: "J", Sdesc: "Package J", Ldesc: "The tenth package, dependent on I", Category: "example", Requires: []string{"i"}},
		{Package: "k", Name: "K", Sdesc: "Package K", Ldesc: "The eleventh package, dependent on J", Category: "example", Requires: []string{"j"}},
		{Package: "l", Name: "L", Sdesc: "Package L", Ldesc: "The twelfth package, dependent on K", Category: "example", Requires: []string{"k"}},
		{Package: "m", Name: "M", Sdesc: "Package M", Ldesc: "The thirteenth package, dependent on L", Category: "example", Requires: []string{"l"}},
		{Package: "n", Name: "N", Sdesc: "Package N", Ldesc: "The fourteenth package, dependent on M", Category: "example", Requires: []string{"m"}},
		{Package: "o", Name: "O", Sdesc: "Package O", Ldesc: "The fifteenth package, dependent on N", Category: "example", Requires: []string{"n"}},
		{Package: "p", Name: "P", Sdesc: "Package P", Ldesc: "The sixteenth package, dependent on O", Category: "example", Requires: []string{"o"}},
		{Package: "q", Name: "Q", Sdesc: "Package Q", Ldesc: "The seventeenth package, dependent on P", Category: "example", Requires: []string{"p"}},
		{Package: "r", Name: "R", Sdesc: "Package R", Ldesc: "The eighteenth package, dependent on Q", Category: "example", Requires: []string{"q"}},
		{Package: "s", Name: "S", Sdesc: "Package S", Ldesc: "The nineteenth package, dependent on R", Category: "example", Requires: []string{"r"}},
		{Package: "t", Name: "T", Sdesc: "Package T", Ldesc: "The twentieth package, dependent on S", Category: "example", Requires: []string{"s"}},
		{Package: "u", Name: "U", Sdesc: "Package U", Ldesc: "The twenty-first package, dependent on T", Category: "example", Requires: []string{"t"}},
		{Package: "v", Name: "V", Sdesc: "Package V", Ldesc: "The twenty-second package, dependent on U", Category: "example", Requires: []string{"u"}},
		{Package: "w", Name: "W", Sdesc: "Package W", Ldesc: "The twenty-third package, dependent on V", Category: "example", Requires: []string{"v"}},
		{Package: "x", Name: "X", Sdesc: "Package X", Ldesc: "The twenty-fourth package, dependent on W", Category: "example", Requires: []string{"w"}},
		{Package: "y", Name: "Y", Sdesc: "Package Y", Ldesc: "The twenty-fifth package, dependent on X", Category: "example", Requires: []string{"x"}},
		{Package: "z", Name: "Z", Sdesc: "Package Z", Ldesc: "The twenty-sixth package, dependent on Y", Category: "example", Requires: []string{"y"}},
	}
	for _, entry := range resolver.Packages {
		resolver.packageDependencies[entry.Package] = entry.Requires
	}
	return resolver
}

func TestShowPackageEntry(t *testing.T) {
	resolver := setupTestResolver()

	// Capture the output
	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.ShowPackageEntry("a")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := "Package: a\nName: A\nShort Description: Package A\nLong Description: The first package in the alphabetical order\nCategory: example\nRequirements: []\n"
	if output.String() != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}

func TestListDirectDependencies(t *testing.T) {
	resolver := setupTestResolver()

	// Capture the output
	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.ListDirectDependencies("z")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := `z
z > y
z > y > x
z > y > x > w
z > y > x > w > v
z > y > x > w > v > u
z > y > x > w > v > u > t
z > y > x > w > v > u > t > s
z > y > x > w > v > u > t > s > r
z > y > x > w > v > u > t > s > r > q
z > y > x > w > v > u > t > s > r > q > p
z > y > x > w > v > u > t > s > r > q > p > o
z > y > x > w > v > u > t > s > r > q > p > o > n
z > y > x > w > v > u > t > s > r > q > p > o > n > m
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e > d
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e > d > c
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e > d > c > b
z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e > d > c > b > a
`
	if output.String() != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}

func TestListDependencyTree(t *testing.T) {
	resolver := setupTestResolver()

	// Capture the output
	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.ListDependencyTree("z")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := "z > y > x > w > v > u > t > s > r > q > p > o > n > m > l > k > j > i > h > g > f > e > d > c > b > a\n"
	if output.String() != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}

func TestListDependencyTreeList(t *testing.T) {
	resolver := setupTestResolver()

	// Capture the output
	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.ListDependencyTreeTopDown("z")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput :=
		`a
b
c
d
e
f
g
h
i
j
k
l
m
n
o
p
q
r
s
t
u
v
w
x
y
z
`

	if output.String() != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}
