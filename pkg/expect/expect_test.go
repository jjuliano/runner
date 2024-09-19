package expect

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestProcessExpectations(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []string
	}{
		{"single string", []string{"single string"}},
		{42, []string{"42"}},
		{[]interface{}{"string1", 43, "string2"}, []string{"string1", "43", "string2"}},
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

func TestCheckExpectations(t *testing.T) {
	// Set up test environment variables and files
	os.Setenv("TEST_ENV_VAR", "1")
	defer os.Unsetenv("TEST_ENV_VAR")

	file, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("could not create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Set up a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	tests := []struct {
		output       string
		exitCode     int
		expectations []string
		hasError     bool
	}{
		{"Hello, World!", 0, []string{"Hello", "World", "0"}, false},
		{"Hello, World!", 0, []string{"!Error", "World", "0"}, false},
		{"Hello, World!", 1, []string{"Hello", "World", "0"}, true},
		{"Hello, World!", 0, []string{"Error"}, true},
		{"Error occurred", 1, []string{"!Hello", "!World", "1"}, false},
		{"", 0, []string{"ENV:TEST_ENV_VAR"}, false},
		{"", 0, []string{"!ENV:NON_EXISTENT_ENV_VAR"}, false},
		{"", 0, []string{"FILE:" + file.Name()}, false},
		{"", 0, []string{"!FILE:non_existent_file"}, false},
		{"", 0, []string{"DIR:" + dir}, false},
		{"", 0, []string{"!DIR:non_existent_dir"}, false},
		{"", 0, []string{"URL:" + mockServer.URL + "/valid"}, false},
		{"", 0, []string{"!URL:" + mockServer.URL + "/invalid"}, false},
		{"", 0, []string{"URL:example.com"}, false},
		{"", 0, []string{"!URL:nonexistent.url"}, false},
		{"", 0, []string{"!CMD:go"}, true},               // Test CMD: prefix
		{"", 0, []string{"CMD:go"}, false},               // Test CMD: prefix
		{"", 0, []string{"CMD:invalid_command"}, true},   // Test CMD: prefix with invalid command
		{"", 0, []string{"!CMD:invalid_command"}, false}, // Test CMD: prefix with invalid command
	}

	client := &http.Client{}

	for _, test := range tests {
		err := CheckExpectations(test.output, test.exitCode, test.expectations, client)
		if (err != nil) != test.hasError {
			t.Errorf("expected error %v, got %v", test.hasError, err)
		}
	}
}
