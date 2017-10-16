package main

type errorString struct{}

func (e errorString) imports() []string {
	return []string{"fmt"}
}

func (e errorString) generate(p *printer) {
	p.Println(`type errorString string`)
	p.Println()
	p.Println(`func errorf(msg string, args ...interface{}) error {`)
	p.Println(`	return errorString(fmt.Sprintf(msg, args...))`)
	p.Println(`}`)
	p.Println()
	p.Println(`func (e errorString) Error() string {`)
	p.Println(`	return string(e)`)
	p.Println(`}`)
}
