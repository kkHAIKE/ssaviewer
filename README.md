# ssaviewer
A simple Golang SSA viewer tool that can be used for code analysis or to create a linter.

ssa.html generate code modify from `src/cmd/compile/internal/ssa/html.go`

because `golang.org/x/tools/go/ssa` and `src/cmd/compile/internal/ssa` are both SSA implementations in Go, but they are very different from each other.

## screenshot
![screenshot](/screenshot.png)

## install
`go get github.com/kkHAIKE/ssaviewer`

## usage
1. cd into your package folder
2. `ssaviewer` show all ssa.Function
3. `ssaviewer function_name` generate the ssa.html

## TODOs
- [ ] AST location binding

## search key
1. GOSSAFUNC
