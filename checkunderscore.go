package checkunderscore

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "checkunderscore is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "checkunderscore",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

type funcInfo struct {
	pos          token.Pos
	called       bool
	isRetHandled []bool
}

func newFuncInfo(pos token.Pos, retLen int) *funcInfo {
	return &funcInfo{pos, false, make([]bool, retLen)}
}

type isRetIgnored []bool

func run(pass *analysis.Pass) (interface{}, error) {
	funcInfos := make(map[string]*funcInfo)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if results := fn.Type.Results; results != nil {
			funcInfos[fn.Name.Name] = newFuncInfo(fn.Pos(), len(results.List))
		}
	})

	inspect.Preorder(nil, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			if call, ok := n.Rhs[0].(*ast.CallExpr); ok {
				for i, l := range n.Lhs {
					if !isIgnored(l) {
						funcInfos[funcName(call)].isRetHandled[i] = true
					}
				}
			}
		case *ast.CallExpr:
			funcInfos[funcName(n)].called = true
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				spec, _ := spec.(*ast.ValueSpec)
				if spec == nil {
					continue
				}
				exprs := spec.Values
				if exprs == nil {
					continue
				}
				if call, ok := exprs[0].(*ast.CallExpr); ok {
					for i, id := range spec.Names {
						if !isIgnored(id) {
							funcInfos[funcName(call)].isRetHandled[i] = true
						}
					}
				}
			}
		}
	})

	for funcName, info := range funcInfos {
		if !info.called {
			continue
		}
		isRetHandled := info.isRetHandled
		for i, handled := range isRetHandled {
			if !handled {
				pass.Reportf(info.pos, message(funcName, i, len(isRetHandled) == 1))
				break
			}
		}
	}

	return nil, nil
}

func isIgnored(e ast.Expr) bool {
	if e, ok := e.(*ast.Ident); ok {
		return e.Name == "_"
	}
	return false
}

func funcName(n *ast.CallExpr) string {
	switch fun := n.Fun.(type) {
	case *ast.Ident:
		return fun.Obj.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	}
	return ""
}

func message(funcName string, nth int, singleRet bool) string {
	if singleRet {
		return fmt.Sprintf("%s: returned value is always ignored.\n", funcName)
	}
	return fmt.Sprintf("%s: %s returned value is always ignored.\n", funcName, nthString(nth))
}

func nthString(n int) string {
	switch n % 10 {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
