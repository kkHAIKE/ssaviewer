# ssaviewer
A simple golang SSA viewer tool use for code analysis or make a linter

ssa.html generate code modify from `src/cmd/compile/internal/ssa/html.go`, because of `golang.org/x/tools/go/ssa` is so different.

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
