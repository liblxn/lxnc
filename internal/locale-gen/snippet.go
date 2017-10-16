package main

type snippet interface {
	imports() []string
	generate(p *printer)
}

type snippetTest interface {
	testImports() []string
	generateTest(p *printer)
}

func snippetTestOf(s snippet) snippetTest {
	type tester interface {
		test() snippetTest
	}

	if t, ok := s.(snippetTest); ok {
		return t
	}
	if t, ok := s.(tester); ok {
		if ts := t.test(); ts != nil {
			return ts
		}
	}
	return nil
}

type snippets []snippet

func (s snippets) imports() []string {
	return uniqueImports(len(s), func(i int) []string {
		return s[i].imports()
	})
}

func (s snippets) test() snippetTest {
	var res snippetTests
	for i := 0; i < len(s); i++ {
		if t := snippetTestOf(s[i]); t != nil {
			res = append(res, t)
		}
	}
	return res.test()
}

func (s snippets) generate(p *printer) {
	if len(s) == 0 {
		return
	}

	s[0].generate(p)
	for i := 1; i < len(s); i++ {
		p.Println()
		s[i].generate(p)
	}
}

type snippetTests []snippetTest

func (s snippetTests) test() snippetTest {
	if len(s) == 0 {
		return nil
	}
	return s
}

func (s snippetTests) testImports() []string {
	return uniqueImports(len(s), func(i int) []string {
		return s[i].testImports()
	})
}

func (s snippetTests) generateTest(p *printer) {
	if len(s) == 0 {
		return
	}

	s[0].generateTest(p)
	for i := 1; i < len(s); i++ {
		p.Println()
		s[i].generateTest(p)
	}
}

func testSnippetOf(s snippet) snippet {
	if t := snippetTestOf(s); t != nil {
		return testSnippet{s: t}
	}
	return nil
}

type testSnippet struct {
	s snippetTest
}

func (t testSnippet) imports() []string {
	return append(t.s.testImports(), "testing")
}

func (t testSnippet) generate(p *printer) {
	t.s.generateTest(p)
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
