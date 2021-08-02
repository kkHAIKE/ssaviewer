package main

import (
	"fmt"
	"os"

	"github.com/kkHAIKE/ssaviewer/gen"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func getAllFunction(pkg *ssa.Package) (ret []*ssa.Function) {
	for _, v := range pkg.Members {
		if vv, ok := v.(*ssa.Function); ok {
			ret = append(ret, vv)

			for _, af := range vv.AnonFuncs {
				ret = append(ret, af)
			}
		}
	}
	return
}

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

	pkg.SetDebugMode(true)
	pkg.Build()
	fs := getAllFunction(pkg)

	if len(os.Args) == 1 {
		for _, v := range fs {
			fmt.Println(v.Name())
		}
		return
	}

	var f *ssa.Function
	for _, v := range fs {
		if v.Name() == os.Args[1] {
			f = v
			break
		}
	}
	if f == nil {
		panic("not found")
	}

	w := gen.NewHTMLWriter("ssa.html", f)
	w.WriteFunc()
	w.Close()
}
