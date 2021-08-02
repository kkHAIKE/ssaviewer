package gen

import (
	"bufio"
	"bytes"
	"go/ast"
	"os"
)

func readLines(fpath string, start, end int) (ret *FuncLines, err error) {
	fp, err := os.Open(fpath)
	if err != nil {
		return
	}
	defer fp.Close()

	ret = &FuncLines{
		Filename:    fpath,
		StartLineno: uint(start),
	}

	s := bufio.NewScanner(fp)
	var i int
	for s.Scan() {
		i++
		if i >= start && i <= end {
			ret.Lines = append(ret.Lines, s.Text())
		} else if i > end {
			break
		}
	}
	return
}

func astBuffer(n ast.Node) (ret *bytes.Buffer) {
	ret = &bytes.Buffer{}
	ast.Fprint(ret, Fset, n, nil)
	return
}

func (w *HTMLWriter) WriteFunc() (err error) {
	f := Fset.File(w.Func.Pos())
	lns, err := readLines(f.Name(), posToLine(w.Func.Syntax().Pos()), posToLine(w.Func.Syntax().End()))
	if err != nil {
		return
	}

	w.WriteSources("sources", []*FuncLines{lns})
	w.WriteAST("AST", astBuffer(w.Func.Syntax()))
	w.WriteColumn("start", "start", "", w.Func.HTML("start"))
	return
}
