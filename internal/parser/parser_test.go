package parser

import (
	"testing"
)

func TestParseBasic(t *testing.T) {
	src := `// header

import b from "b";
// attached
import a from "a";

const x = 1;
`
	res, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Imports) != 2 {
		t.Fatalf("Expected 2 imports, got %d", len(res.Imports))
	}
	if res.Imports[0].Original != "import b from \"b\";" {
		t.Errorf("Unexpected first import: %q", res.Imports[0].Original)
	}
	if res.Imports[1].Original != "\n// attached\nimport a from \"a\";" {
		t.Errorf("Unexpected second import: %q", res.Imports[1].Original)
	}
}

func TestParseCommonJS(t *testing.T) {
	src := `
const foo = require("foo");
let bar = require("bar");

function test() {
    const baz = require("baz");
}
`
	res, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Imports) != 2 {
		t.Fatalf("Expected 2 imports, got %d", len(res.Imports))
	}
	if res.Imports[0].Original != "const foo = require(\"foo\");" {
		t.Errorf("Unexpected first import: %q", res.Imports[0].Original)
	}
	if res.Imports[1].Original != "\nlet bar = require(\"bar\");" {
		t.Errorf("Unexpected second import: %q", res.Imports[1].Original)
	}
}
