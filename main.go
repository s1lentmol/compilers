package main

import (
	"compilers/lexer"
	"compilers/parser"
	"compilers/semantic"
	"fmt"
)

func main() {
	codeExample := "var x = 123; \nprint x + 5;"

	l := lexer.NewLexer(codeExample)
	tokens := l.Tokenize()

	p := parser.NewParser(tokens)
	statements := p.Parse()

	errors := semantic.Analyze(statements)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		return
	}
	for _, stmt := range statements {
		fmt.Printf("%#v\n", stmt)
	}
}
