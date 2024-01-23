package generator

type tester interface {
	testSnippet() TestSnippet
}

type Snippet interface {
	Imports() []string
	Generate(p *Printer)
}

var (
	_ Snippet = (Snippets)(nil)
	_ tester  = (Snippets)(nil)
)

type Snippets []Snippet

func (s Snippets) Imports() []string {
	return uniqueImports(len(s), func(i int) []string {
		return s[i].Imports()
	})
}

func (s Snippets) Generate(p *Printer) {
	generateSnippets(s, Snippet.Generate, p)
}

func (s Snippets) testSnippet() TestSnippet {
	var res testSnippets
	for i, n := 0, len(s); i < n; i++ {
		if t := testSnippetOf(s[i]); t != nil {
			res = append(res, t)
		}
	}
	return res.testSnippet()
}

// TestSnippet is an interface for snippets within a test file. Each snippet
// can implement this interface to generate tests.
type TestSnippet interface {
	TestImports() []string
	GenerateTest(p *Printer)
}

func testSnippetOf(s Snippet) TestSnippet {
	type tester interface {
		testSnippet() TestSnippet
	}

	if t, ok := s.(TestSnippet); ok {
		return t
	}
	if t, ok := s.(tester); ok {
		if ts := t.testSnippet(); ts != nil {
			return ts
		}
	}
	return nil
}

var (
	_ TestSnippet = (testSnippets)(nil)
	_ tester      = (testSnippets)(nil)
)

type testSnippets []TestSnippet

func (s testSnippets) TestImports() []string {
	return uniqueImports(len(s), func(i int) []string {
		return s[i].TestImports()
	})
}

func (s testSnippets) GenerateTest(p *Printer) {
	generateSnippets(s, TestSnippet.GenerateTest, p)
}

func (s testSnippets) testSnippet() TestSnippet {
	if len(s) == 0 {
		return nil
	}
	return s
}

func generateSnippets[S any](snippets []S, generate func(S, *Printer), p *Printer) {
	if len(snippets) == 0 {
		return
	}

	generate(snippets[0], p)
	for _, s := range snippets[1:] {
		p.Println()
		generate(s, p)
	}
}

func uniqueImports(n int, f func(int) []string) []string {
	set := make(map[string]struct{})
	res := make([]string, 0, n)
	for i := 0; i < n; i++ {
		for _, s := range f(i) {
			if _, has := set[s]; !has {
				res = append(res, s)
				set[s] = struct{}{}
			}
		}
	}
	return res
}
