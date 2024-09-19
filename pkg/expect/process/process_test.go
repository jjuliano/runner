package process

import (
	"os"
	"testing"
)

func TestProcessExpectations(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		input    interface{}
		expected []string
	}{
		{"single string", []string{"single string"}},
		{42, []string{"42"}},
		{[]interface{}{"string1", 43, "string2"}, []string{"string1", "43", "string2"}},
		{"a ${TEST_VAR} variable", []string{"a test_value variable"}},
	}

	for _, test := range tests {
		result := ProcessExpectations(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("expected length %d, got %d", len(test.expected), len(result))
		}
		for i, exp := range test.expected {
			if result[i] != exp {
				t.Errorf("expected %q, got %q", exp, result[i])
			}
		}
	}
}
