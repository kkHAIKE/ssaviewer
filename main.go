package main

import (
	"fmt"
	"os"

	"github.com/kkHAIKE/ssaviewer/gen"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadSyntax,
		Fset: gen.Fset,
	}, ".")
	if err != nil {
		panic(err)
	}
	_, ssapkgs := ssautil.Packages(pkgs, 0)
	pkg := ssapkgs[0]

	if len(os.Args) == 1 {
		for k, v := range pkg.Members {
			if _, ok := v.(*ssa.Function); ok {
				fmt.Println(k)
			}
		}
		return
	}

	pkg.SetDebugMode(true)
	pkg.Build()
	f := pkg.Func(os.Args[1])
	if f == nil {
		panic("not found")
	}

	w := gen.NewHTMLWriter("ssa.html", f)
	w.WriteFunc()
	w.Close()
}
