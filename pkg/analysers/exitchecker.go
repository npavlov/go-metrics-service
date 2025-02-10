package analysers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	mainPackage   = "main"
	osImport      = "\"os\""
	mainFunction  = "main"
	exitFunction  = "Exit"
	osPackageName = "os"
)

//nolint:exhaustruct,gochecknoglobals
var ExitCheckAnalyser = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Detect calls to os.Exit() and classify them as errors or warnings based on package and function context",
	Run:  run,
}

//nolint:nilnil,gocognit,cyclop
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		inMainPackage := file.Name.Name == mainPackage
		osImported := false

		// Check if "os" is imported
		for _, imp := range file.Imports {
			if imp.Path.Value == osImport {
				osImported = true

				break
			}
		}

		if !osImported {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			funcNode, ok := node.(*ast.FuncDecl)

			//nolint:nestif
			if ok {
				ast.Inspect(funcNode, func(subNode ast.Node) bool {
					if call, ok := subNode.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == osPackageName && sel.Sel.Name == exitFunction {
								if inMainPackage && funcNode.Name.Name == mainFunction {
									pass.Reportf(pkg.NamePos, "error: calling os.Exit in main function of main package")
								} else {
									pass.Reportf(pkg.NamePos, "warning: calling os.Exit")
								}
							}
						}
					}

					return true
				})
			}

			return true
		})
	}

	return nil, nil
}
