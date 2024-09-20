package resolver

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/jjuliano/runner/pkg/runnerexec"
	"github.com/spf13/afero"
)

// Helper function to initialize a DependencyResolver with sample data.
func setupTestResolver() *DependencyResolver {
	fs := afero.NewMemMapFs() // Using an in-memory filesystem for testing
	logger := log.New(nil)
	session, err := runnerexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	resolver, err := NewGraphResolver(fs, logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	resolver.Resources = []ResourceNodeEntry{
		{Id: "a", Name: "A", Desc: "The first resource in the alphabetical order", Category: "example", Requires: []string{}},
		{Id: "b", Name: "B", Desc: "The second resource, dependent on A", Category: "example", Requires: []string{"a"}},
		{Id: "c", Name: "C", Desc: "The third resource, dependent on B", Category: "example", Requires: []string{"b"}},
		{Id: "d", Name: "D", Desc: "The fourth resource, dependent on C", Category: "example", Requires: []string{"c"}},
		{Id: "e", Name: "E", Desc: "The fifth resource, dependent on D", Category: "example", Requires: []string{"d"}},
		{Id: "f", Name: "F", Desc: "The sixth resource, dependent on E", Category: "example", Requires: []string{"e"}},
		{Id: "g", Name: "G", Desc: "The seventh resource, dependent on F", Category: "example", Requires: []string{"f"}},
		{Id: "h", Name: "H", Desc: "The eighth resource, dependent on G", Category: "example", Requires: []string{"g"}},
		{Id: "i", Name: "I", Desc: "The ninth resource, dependent on H", Category: "example", Requires: []string{"h"}},
		{Id: "j", Name: "J", Desc: "The tenth resource, dependent on I", Category: "example", Requires: []string{"i"}},
		{Id: "k", Name: "K", Desc: "The eleventh resource, dependent on J", Category: "example", Requires: []string{"j"}},
		{Id: "l", Name: "L", Desc: "The twelfth resource, dependent on K", Category: "example", Requires: []string{"k"}},
		{Id: "m", Name: "M", Desc: "The thirteenth resource, dependent on L", Category: "example", Requires: []string{"l"}},
		{Id: "n", Name: "N", Desc: "The fourteenth resource, dependent on M", Category: "example", Requires: []string{"m"}},
		{Id: "o", Name: "O", Desc: "The fifteenth resource, dependent on N", Category: "example", Requires: []string{"n"}},
		{Id: "p", Name: "P", Desc: "The sixteenth resource, dependent on O", Category: "example", Requires: []string{"o"}},
		{Id: "q", Name: "Q", Desc: "The seventeenth resource, dependent on P", Category: "example", Requires: []string{"p"}},
		{Id: "r", Name: "R", Desc: "The eighteenth resource, dependent on Q", Category: "example", Requires: []string{"q"}},
		{Id: "s", Name: "S", Desc: "The nineteenth resource, dependent on R", Category: "example", Requires: []string{"r"}},
		{Id: "t", Name: "T", Desc: "The twentieth resource, dependent on S", Category: "example", Requires: []string{"s"}},
		{Id: "u", Name: "U", Desc: "The twentyfirst resource, dependent on T", Category: "example", Requires: []string{"t"}},
		{Id: "v", Name: "V", Desc: "The twentysecond resource, dependent on U", Category: "example", Requires: []string{"u"}},
		{Id: "w", Name: "W", Desc: "The twentythird resource, dependent on V", Category: "example", Requires: []string{"v"}},
		{Id: "x", Name: "X", Desc: "The twentyfourth resource, dependent on W", Category: "example", Requires: []string{"w"}},
		{Id: "y", Name: "Y", Desc: "The twentyfifth resource, dependent on X", Category: "example", Requires: []string{"x"}},
		{Id: "z", Name: "Z", Desc: "The twentysixth resource, dependent on Y", Category: "example", Requires: []string{"y"}},
	}
	for _, entry := range resolver.Resources {
		resolver.ResourceDependencies[entry.Id] = entry.Requires
	}
	return resolver
}

func TestShowResourceEntry(t *testing.T) {
	resolver := setupTestResolver()

	// Capture the output
	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.ShowResourceEntry("a")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := "📦 Id: a\n📛 Name: A\n📝 Description: The first resource in the alphabetical order\n🏷️  Category: example\n🔗 Requirements: []\n"

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

	resolver.Graph.ListDirectDependencies("z")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := `z
z -> y
z -> y -> x
z -> y -> x -> w
z -> y -> x -> w -> v
z -> y -> x -> w -> v -> u
z -> y -> x -> w -> v -> u -> t
z -> y -> x -> w -> v -> u -> t -> s
z -> y -> x -> w -> v -> u -> t -> s -> r
z -> y -> x -> w -> v -> u -> t -> s -> r -> q
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f -> e
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f -> e -> d
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f -> e -> d -> c
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f -> e -> d -> c -> b
z -> y -> x -> w -> v -> u -> t -> s -> r -> q -> p -> o -> n -> m -> l -> k -> j -> i -> h -> g -> f -> e -> d -> c -> b -> a
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

	resolver.Graph.ListDependencyTree("z")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := "z <- y <- x <- w <- v <- u <- t <- s <- r <- q <- p <- o <- n <- m <- l <- k <- j <- i <- h <- g <- f <- e <- d <- c <- b <- a\n"
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

	resolver.Graph.ListDependencyTreeTopDown("z")

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

func TestListDirectDependencies_CircularDependency(t *testing.T) {
	resolver := setupTestResolver()
	resolver.Resources = []ResourceNodeEntry{
		{Id: "a", Name: "A", Requires: []string{"c"}},
		{Id: "b", Name: "B", Requires: []string{"a"}},
		{Id: "c", Name: "C", Requires: []string{"b"}},
	}
	for _, entry := range resolver.Resources {
		resolver.ResourceDependencies[entry.Id] = entry.Requires
	}

	var output strings.Builder
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	resolver.Graph.ListDirectDependencies("a")

	w.Close()
	os.Stdout = old
	io.Copy(&output, r)

	expectedOutput := `a
a -> c
a -> c -> b
`
	if output.String() != expectedOutput {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}
