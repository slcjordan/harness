package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <file.go>")
	}

	filename := os.Args[1]
	outfile := os.Args[2]

	// Set up the parser
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filename, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("working!", filename)

	// Walk the AST
	ast.Inspect(node, func(n ast.Node) bool {
		pk, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		if strings.Contains(pk.Value, "PRIVATE KEY") {
			os.WriteFile(outfile, []byte(strings.Trim(pk.Value, "`")), 600)
		}
		return true
	})
}
