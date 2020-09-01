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
	pos     token.Pos
	called  bool
	realRet []bool
}

func run(pass *analysis.Pass) (interface{}, error) {
	infos := make(map[string]*funcInfo)
	var calledFuncNames []string
	lhsNames := make(map[string][][]string)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder(nil, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			if call, ok := n.Rhs[0].(*ast.CallExpr); ok {
				funcName := call.Fun.(*ast.Ident).Obj.Name
				lhsName := make([]string, 0)
				for _, lhs := range n.Lhs {
					lhsName = append(lhsName, lhs.(*ast.Ident).Name)
				}
				lhsNames[funcName] = append(lhsNames[funcName], lhsName)
			}
		case *ast.CallExpr:
			calledFuncNames = append(calledFuncNames, n.Fun.(*ast.Ident).Obj.Name)
		case *ast.FuncDecl:
			if n.Type.Results == nil {
				return
			}
			returnValues := n.Type.Results.List
			infos[n.Name.Name] = &funcInfo{n.Pos(), false, make([]bool, len(returnValues))}
		case *ast.GenDecl:
			specs := n.Specs
			for _, spec := range specs {
				if spec, ok := spec.(*ast.ValueSpec); ok {
					exprs := spec.Values
					if exprs == nil {
						continue
					}
					if call, ok := exprs[0].(*ast.CallExpr); ok {
						funcName := call.Fun.(*ast.Ident).Obj.Name
						lhsName := make([]string, 0)
						for _, id := range spec.Names {
							lhsName = append(lhsName, id.Name)
						}
						lhsNames[funcName] = append(lhsNames[funcName], lhsName)
					}
				}
			}
		}
	})

	for _, name := range calledFuncNames {
		infos[name].called = true
	}

	for fnName, lhsName := range lhsNames {
		info := infos[fnName]
		for _, lhs := range lhsName {
			for i, l := range lhs {
				if l != "_" {
					info.realRet[i] = true
				}
			}
		}
	}

	for fn, info := range infos {
		if !info.called {
			continue
		}
		for i, isReal := range info.realRet {
			if !isReal {
				var msg string
				if len(info.realRet) == 1 {
					msg = fmt.Sprintf("%s: returned value is always unhandled.\n", fn)
				} else {
					msg = fmt.Sprintf("%s: %s returned value is always unhandled.\n", fn, nthString(i))
				}
				pass.Reportf(info.pos, msg)
			}
		}
	}

	return nil, nil
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
