package main

import (
	"fmt"
	"os"

	"github.com/kkHAIKE/ssaviewer/gen"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func getAllFunction(prog *ssa.Program, pkg *ssa.Package) (ret []*ssa.Function) {
	add := func(f *ssa.Function) {
		ret = append(ret, f)

		for _, af := range f.AnonFuncs {
			ret = append(ret, af)
		}
	}

	for _, v := range pkg.Members {
		switch vv := v.(type) {
		case *ssa.Function:
			add(vv)
		case *ssa.Type:
			if ts := prog.MethodSets.MethodSet(vv.Type()); ts != nil {
				for i := 0; i < ts.Len(); i++ {
					add(prog.MethodValue(ts.At(i)))
				}
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
	prog, ssapkgs := ssautil.Packages(pkgs, 0)
	pkg := ssapkgs[0]

	pkg.SetDebugMode(true)
	pkg.Build()
	fs := getAllFunction(prog, pkg)

	if len(os.Args) == 1 {
		for _, v := range fs {
			fmt.Println(v.RelString(pkg.Pkg))
		}
		return
	}

	var f *ssa.Function
	for _, v := range fs {
		if v.RelString(pkg.Pkg) == os.Args[1] {
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
