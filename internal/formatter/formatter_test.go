package formatter

import (
	"testing"

	"github.com/0necontroller/jsimportfmt/internal/parser"
)

func TestFormatBasic(t *testing.T) {
	src := `// header

import somethingVeryLong from "foo";
import b from "b";
import a from "a";

const x = 1;
`
	res, err := parser.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	formatted := Format(res, false)
	expected := `// header

import a from "a";
import b from "b";
import somethingVeryLong from "foo";

const x = 1;
`
	if formatted != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q\n", expected, formatted)
	}
}

func TestFormatSeparateTypes(t *testing.T) {
	src := `
import type { Foo } from "./types";
import React from "react";
import type { Bar } from "./bar";
import { Button } from "./button";
`
	res, err := parser.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	formatted := Format(res, true)
	expected := `
import React from "react";
import { Button } from "./button";

import type { Bar } from "./bar";
import type { Foo } from "./types";
`
	if formatted != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q\n", expected, formatted)
	}
}
