package main

import (
	"compilers/interpreter"
	"compilers/lexer"
	"compilers/optimizer"
	"compilers/parser"
	"compilers/semantic"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	showTokens := flag.Bool("tokens", false, "вывести поток токенов и выйти")
	flag.Parse()

	source, err := readSource(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(run(source, *showTokens))
}

func readSource(args []string) (string, error) {
	if len(args) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("не удалось прочитать stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл %q: %w", args[0], err)
	}
	return string(data), nil
}

func run(source string, showTokens bool) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			exitCode = 1
		}
	}()

	tokens := lexer.NewLexer(source).Tokenize()
	if showTokens {
		for _, token := range tokens {
			fmt.Println(token)
		}
		return 0
	}

	statements := parser.NewParser(tokens).Parse()

	result := semantic.Analyze(statements)
	for _, warning := range result.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
	}
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "error: %s\n", e)
		}
		return 1
	}

	statements = optimizer.Optimize(statements)

	if err := interpreter.New().Run(statements); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
