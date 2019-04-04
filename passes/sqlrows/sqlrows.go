// Copyright 2019 The Go Authors. All rights reserved.

// Package sqlrows defines an Analyzer that checks for mistakes using sql.Rows.
package sqlrows

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/sqlrows/sqlrowsutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/ssa"
)

const Doc = `check for mistakes using Rows iterator of database/sql 
A common mistake when using the database/sql package is to defer a function
call to close the Rows before checking the error that
determines whether the returned records are valid:
	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age=?", age)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	// (defer statement belongs here)
This checker helps uncover latent nil dereference bugs by reporting a
diagnostic for such mistakes.`

var Analyzer = &analysis.Analyzer{
	Name: "sqlrows",
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		buildssa.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Fast path: if the package doesn't import database/sql,
	// skip the traversal.
	if !imports(pass.Pkg, "database/sql") {
		return nil, nil
	}

	rowsType := analysisutil.TypeOf(pass, "database/sql", "*Rows")
	if rowsType == nil {
		// skip checking
		return nil, nil
	}

	var methods []*types.Func
	if m := analysisutil.MethodOf(rowsType, "Close"); m != nil {
		methods = append(methods, m)
	}

	for _, f := range funcs {
		for _, b := range f.Blocks {
			for i := range b.Instrs {
				var pos token.Pos
				switch instr := b.Instrs[i].(type) {
				case *ssa.Extract:
					fmt.Println(instr.Type())
					pos = instr.Tuple.Pos()
				default:
					pos = instr.Pos()
				}
				called, ok := sqlrowsutil.CalledFrom(b, i, rowsType, methods...)
				if ok && !called {
					pass.Reportf(pos, "rows.Close must be called")
				}
			}
		}
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		call := n.(*ast.CallExpr)
		if !hasRowsSignature(pass.TypesInfo, call) {
			return true // the function call is not related to this check.
		}

		// Find the innermost containing block, and get the list
		// of statements starting with the one containing call.
		stmts := restOfBlock(stack)
		if len(stmts) < 2 {
			return true // the call to the http function is the last statement of the block.
		}

		asg, ok := stmts[0].(*ast.AssignStmt)
		if !ok {
			return true // the first statement is not assignment.
		}
		resp := rootIdent(asg.Lhs[0])
		if resp == nil {
			return true // could not find the sql.Rows in the assignment.
		}

		def, ok := stmts[1].(*ast.DeferStmt)
		if !ok {
			return true // the following statement is not a defer.
		}
		root := rootIdent(def.Call.Fun)
		if root == nil {
			return true // could not find the receiver of the defer call.
		}

		if resp.Obj == root.Obj {
			pass.Reportf(root.Pos(), "using %s before checking for errors", resp.Name)
		}
		return true
	})
	return nil, nil
}

// hasRowsSignature checks whether the given call expression is on
// either a function of the database/sql package that returns (*sql.Rows, error).
func hasRowsSignature(info *types.Info, expr *ast.CallExpr) bool {
	fun, _ := expr.Fun.(*ast.SelectorExpr)
	sig, _ := info.Types[fun].Type.(*types.Signature)
	if sig == nil {
		return false // the call is not of the form x.f()
	}

	res := sig.Results()
	if res.Len() != 2 {
		return false // the function called does not return two values.
	}
	if ptr, ok := res.At(0).Type().(*types.Pointer); !ok || !isNamedType(ptr.Elem(), "database/sql", "Rows") {
		return false // the first return type is not *sql.Rows.
	}

	errorType := types.Universe.Lookup("error").Type()
	if !types.Identical(res.At(1).Type(), errorType) {
		return false // the second return type is not error
	}

	return true
}

// restOfBlock, given a traversal stack, finds the innermost containing
// block and returns the suffix of its statements starting with the
// current node (the last element of stack).
func restOfBlock(stack []ast.Node) []ast.Stmt {
	for i := len(stack) - 1; i >= 0; i-- {
		if b, ok := stack[i].(*ast.BlockStmt); ok {
			for j, v := range b.List {
				if v == stack[i+1] {
					return b.List[j:]
				}
			}
			break
		}
	}
	return nil
}

// rootIdent finds the root identifier x in a chain of selections x.y.z, or nil if not found.
func rootIdent(n ast.Node) *ast.Ident {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return rootIdent(n.X)
	case *ast.Ident:
		return n
	default:
		return nil
	}
}

// isNamedType reports whether t is the named type path.name.
func isNamedType(t types.Type, path, name string) bool {
	n, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := n.Obj()
	return obj.Name() == name && obj.Pkg() != nil && obj.Pkg().Path() == path
}

func imports(pkg *types.Package, path string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return true
		}
	}
	return false
}
