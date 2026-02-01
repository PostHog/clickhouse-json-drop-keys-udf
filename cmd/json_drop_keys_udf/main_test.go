package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessLineErrorsOnMalformedJSON(t *testing.T) {
	var buf bytes.Buffer
	err := processLine(nil, []byte("{\"a\":"), &buf)
	assert.Error(t, err, "expected error for malformed JSON, got nil")
}

func TestDropKeysJSON(t *testing.T) {
	cases := []struct {
		name, input, want string
		keys              []string
	}{
		{
			"empty",
			"{}",
			"{}",
			nil,
		},
		{
			"empty2",
			"{}",
			"{}",
			[]string{"jeden"},
		},
		{
			"one one key to be dropped",
			`{"jeden": 1}`,
			`{}`,
			[]string{"jeden"},
		},
		{
			name:  "one key to be dropped, one to be kept",
			input: `{"jeden": 1, "dwa": 2}`,
			want:  `{"dwa":2}`,
			keys:  []string{"jeden"},
		},
		{
			name:  "one key to be dropped one to be kept (order doesnt matter)",
			input: `{"dwa": 2, "jeden": 1}`,
			want:  `{"dwa":2}`,
			keys:  []string{"jeden"},
		},
		{
			name:  "multiple keys to be dropped one to be kept (order doesnt matter)",
			input: `{"dwa": 2, "jeden": 1, "trzy": 3, "cztery": 4, "piec": {"dwa": 1}}`,
			want:  `{"jeden":1,"cztery":4,"piec":{"dwa":1}}`,
			keys:  []string{"dwa", "trzy"},
		},
		{
			name:  "drop nested key with dot notation",
			input: `{"id":1,"props":{"secret":"xxx","public":"yyy"}}`,
			want:  `{"id":1,"props":{"public":"yyy"}}`,
			keys:  []string{"props.secret"},
		},
		{
			name:  "drop deeply nested key",
			input: `{"a":{"b":{"c":1,"d":2}}}`,
			want:  `{"a":{"b":{"d":2}}}`,
			keys:  []string{"a.b.c"},
		},
		{
			name:  "drop entire nested object",
			input: `{"a":{"b":1},"c":2}`,
			want:  `{"c":2}`,
			keys:  []string{"a"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := processLine(makeKeyDict(c.keys), []byte(c.input), &buf)
			assert.NoError(t, err, "unexpected error processing line")
			assert.Equal(t, c.want, buf.String(), "unexpected output")
		})
	}
}

func TestParseSingleQuotedArray(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"empty array", "[]", nil, false},
		{"single element", "['foo']", []string{"foo"}, false},
		{"two elements", "['foo', 'bar']", []string{"foo", "bar"}, false},
		{"escaped single quote", `['some other \'string']`, []string{"some other 'string"}, false},
		{"mixed", `['some string', 'some other \'string']`, []string{"some string", "some other 'string"}, false},
		{"with spaces", "[ 'a' , 'b' ]", []string{"a", "b"}, false},
		{"no brackets", "foo", nil, true},
		{"unterminated string", "['foo", nil, true},
		{"missing quote", "[foo]", nil, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseSingleQuotedArray(c.input)
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want, got)
			}
		})
	}
}

func TestMakeKeyDict(t *testing.T) {
	cases := []struct {
		name string
		keys []string
		want jsonKey
	}{
		{
			name: "nil input",
			keys: nil,
			want: jsonKey{},
		},
		{
			name: "empty input",
			keys: []string{},
			want: jsonKey{},
		},
		{
			name: "single top-level key",
			keys: []string{"a"},
			want: jsonKey{"a": nil},
		},
		{
			name: "multiple top-level keys",
			keys: []string{"a", "b", "c"},
			want: jsonKey{"a": nil, "b": nil, "c": nil},
		},
		{
			name: "single nested key",
			keys: []string{"a.b"},
			want: jsonKey{"a": jsonKey{"b": nil}},
		},
		{
			name: "deeply nested key",
			keys: []string{"a.b.c.d"},
			want: jsonKey{"a": jsonKey{"b": jsonKey{"c": jsonKey{"d": nil}}}},
		},
		{
			name: "mixed top-level and nested keys",
			keys: []string{"x", "a.b"},
			want: jsonKey{"x": nil, "a": jsonKey{"b": nil}},
		},
		{
			name: "multiple nested keys under same parent",
			keys: []string{"a.b", "a.c"},
			want: jsonKey{"a": jsonKey{"b": nil, "c": nil}},
		},
		{
			name: "nested key and parent key both specified",
			keys: []string{"a.b", "a"},
			want: jsonKey{"a": nil},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := makeKeyDict(c.keys)
			assert.Equal(t, c.want, got)
		})
	}
}
