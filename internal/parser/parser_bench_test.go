package parser

import "testing"

func BenchmarkParse(b *testing.B) {
	src := `// header
import somethingVeryLong from "foo";
import b from "b";
import a from "a";

const x = 1;
`
	content := []byte(src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(content)
	}
}
