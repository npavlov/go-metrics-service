package analysers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	specPackage      = "main"
	specImport       = "\"os\""
	specDeclaration  = "main"
	specFunction     = "Exit"
	specFunctionPrfx = "os"
)

//nolint:exhaustruct,gochecknoglobals
var ExitCheckAnalyser = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for calling os.Exit()",
	Run:  run,
}

//nolint:gocognit,cyclop,nonamedreturns
func run(pass *analysis.Pass) (res any, err error) {
	for _, file := range pass.Files {
		haveSpecImport := false
		ast.Inspect(file, func(node ast.Node) bool {
			switch nodeType := node.(type) {
			case *ast.File:
				if nodeType.Name.Name != specPackage {
					return false
				}
			case *ast.ImportSpec:
				if nodeType.Path.Value == specImport {
					haveSpecImport = true
				}
			case *ast.FuncDecl:
				if !haveSpecImport {
					return false
				}
				if nodeType.Name.Name != specDeclaration {
					return false
				}

				return true
			case *ast.CallExpr:
				if f, ok := nodeType.Fun.(*ast.SelectorExpr); ok {
					if pkg, ok := f.X.(*ast.Ident); ok {
						if pkg.Name == specFunctionPrfx && f.Sel.Name == specFunction {
							pass.Reportf(pkg.NamePos, "calling os.Exit in function main")

							return false
						}
					}
				}
			}

			return true
		})
	}
	//nolint:nilnil
	return nil, nil
}
