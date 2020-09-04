package checkunderscore

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "checkunderscore checks for returned value which is always ignored."

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

func run(pass *analysis.Pass) (interface{}, error) {
	funcInfos := make(map[string]*funcInfo)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder(nil, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for i, rhs := range n.Rhs {
				rhs, _ := rhs.(*ast.FuncLit)
				if rhs == nil {
					continue
				}
				switch lhs := n.Lhs[i].(type) {
				case *ast.Ident:
					if results := rhs.Type.Results; results != nil {
						funcInfos[lhs.Name] = newFuncInfo(n.Pos(), len(results.List))
					}
				}
			}
		case *ast.FuncDecl:
			if results := n.Type.Results; results != nil {
				funcInfos[n.Name.Name] = newFuncInfo(n.Pos(), len(results.List))
			}
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
				for i, ex := range exprs {
					ex, _ := ex.(*ast.FuncLit)
					if ex == nil {
						continue
					}
					if results := ex.Type.Results; results != nil {
						funcInfos[spec.Names[i].Name] = newFuncInfo(n.Pos(), len(results.List))
					}
				}
			}
		}
	})

	inspect.Preorder(nil, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			if call, ok := n.Rhs[0].(*ast.CallExpr); ok {
				info, _ := funcInfos[funcName(call)]
				if info == nil {
					return
				}
				info.called = true
				for i, l := range n.Lhs {
					if isNotIgnored(l) {
						info.isRetHandled[i] = true
					}
				}
			}
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
						if isNotIgnored(id) {
							info, _ := funcInfos[funcName(call)]
							if info == nil {
								continue
							}
							info.isRetHandled[i] = true
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

func isNotIgnored(e ast.Expr) bool {
	if e, ok := e.(*ast.Ident); ok {
		return e.Name != "_"
	}
	return true
}

func funcName(n *ast.CallExpr) string {
	switch fun := n.Fun.(type) {
	case *ast.Ident:
		if fun.Obj != nil {
			return fun.Obj.Name
		}
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
