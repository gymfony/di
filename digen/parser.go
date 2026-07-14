package digen

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"log/slog"

	"github.com/gymfony/di"
	"golang.org/x/tools/go/packages"
)

type Parser struct {
	log *slog.Logger
}

func NewParser(log *slog.Logger) *Parser {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Parser{log: log}
}

func (p *Parser) Parse(path string) (*di.DependencyGraph, error) {
	graph := di.NewDependencyGraph()

	if path == "" {
		return graph, fmt.Errorf("digen: project path cannot be empty")
	}

	p.log.Debug("digen: project path defined", "path", path)

	pkgs, err := p.getPackages(path)
	if err != nil {
		p.log.Error("digen: package loading failed", "err", err)
		return graph, err
	}
	if len(pkgs) == 0 {
		p.log.Info("digen: list of packages is empty")
		return graph, nil
	}

	p.parseAST(graph, pkgs)

	return graph, nil
}

func (p *Parser) getPackages(path string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  path,
	}

	pkg, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("digen: cannot read packages: %w", err)
	}
	for _, p := range pkg {
		if len(p.Errors) > 0 {
			msgerror := fmt.Sprintf("Errors in package %s:\n", p.ID)
			for _, pkgErr := range p.Errors {
				msgerror = fmt.Sprintf("%s  - %s\n", msgerror, pkgErr.Error())
			}
			return nil, fmt.Errorf("digen: %s", msgerror)
		}
	}

	return pkg, nil
}

func (p *Parser) parseAST(graph *di.DependencyGraph, pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			p.log.Debug("digen: parsing file", "file", file.Name.Name)

			if isGeneratedFile(file) {
				p.log.Debug("digen: skipping generated file", "file", file.Name.Name)
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				// 1. We only need variable declarations (var Services = ...)
				valueSpec, ok := n.(*ast.ValueSpec)
				if !ok {
					return true
				}

				// 2. We loop through the values ​​to the right of the "=" sign
				for _, val := range valueSpec.Values {
					// 3. Let's check that this is a function call.
					callExpr, ok := val.(*ast.CallExpr)
					if !ok {
						continue
					}

					// 4. We are passing the challenge on to rigorous packet and signature analysis
					p.inspectNewSetCall(pkg, callExpr, graph)
				}

				return true
			})
		}
	}
}

// inspectNewSetCall analyzes the di.NewSet call and logs the providers in the graph.
func (p *Parser) inspectNewSetCall(pkg *packages.Package, callExpr *ast.CallExpr, graph *di.DependencyGraph) {
	selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	obj := pkg.TypesInfo.ObjectOf(selectorExpr.Sel)
	if obj == nil {
		return
	}

	pk := obj.Pkg()
	if pk == nil || pk.Path() != "github.com/gymfony/di" || obj.Name() != "NewSet" {
		return
	}

	p.log.Debug("digen: found provider set", "pkg", pk.Path(), "func", obj.Name())

	for _, arg := range callExpr.Args {
		p.processSetArg(pkg, arg, graph)
	}
}

// processSetArg processes one argument from di.NewSet (whether a regular constructor or wrapped in di.Tag).
func (p *Parser) processSetArg(pkg *packages.Package, arg ast.Expr, graph *di.DependencyGraph) {
	currentCtor, boundTags := p.unwrapTagCall(pkg, arg)

	tv, ok := pkg.TypesInfo.Types[currentCtor]
	if !ok {
		p.log.Warn("digen: unable to determine type for argument", "arg", currentCtor)
		return
	}

	sig, ok := tv.Type.(*types.Signature)
	if !ok {
		return
	}
	if sig.Results().Len() == 0 {
		p.log.Error("digen: constructor must return at least one value", "sig", sig.String())
		return
	}

	resultType := sig.Results().At(0).Type().String()
	p.log.Debug("digen: constructor analyzed", "sig", sig.String(), "returns", resultType)

	ctorName := "unknown"
	if ident, ok := currentCtor.(*ast.Ident); ok {
		ctorName = ident.Name
	}

	var deps []string
	paramsLen := sig.Params().Len()
	for j := 0; j < paramsLen; j++ {
		depType := sig.Params().At(j).Type().String()
		p.log.Debug("digen: dependency found", "constructor", ctorName, "param", depType)
		deps = append(deps, depType)
	}

	graph.Nodes[resultType] = &di.ServiceNode{
		Type:     resultType,
		CtorName: ctorName,
		Deps:     deps,
		Tags:     boundTags,
	}
}

// unwrapTagCall checks whether the expression is a call to di.Tag(ctor, key).
// If so, returns the unwrapped constructor and the extracted tag type.
func (p *Parser) unwrapTagCall(pkg *packages.Package, arg ast.Expr) (ast.Expr, []string) {
	innerCall, ok := arg.(*ast.CallExpr)
	if !ok {
		return arg, nil
	}

	selector, ok := innerCall.Fun.(*ast.SelectorExpr)
	if !ok {
		return arg, nil
	}

	innerObj := pkg.TypesInfo.ObjectOf(selector.Sel)
	if innerObj == nil || innerObj.Pkg() == nil {
		return arg, nil
	}

	if innerObj.Pkg().Path() != "github.com/gymfony/di" || innerObj.Name() != "Tag" {
		return arg, nil
	}

	p.log.Debug("digen: found di.Tag wrapper, extracting constructor and tags")
	currentCtor := innerCall.Args[0]
	var boundTags []string

	if len(innerCall.Args) > 1 {
		if targetType := p.extractTagType(pkg, innerCall.Args[1]); targetType != "" {
			boundTags = append(boundTags, targetType)
		}
	}

	return currentCtor, boundTags
}

// extractTagType extracts the string representation of type T from TagKey[T].
func (p *Parser) extractTagType(pkg *packages.Package, tagKeyArg ast.Expr) string {
	tv, ok := pkg.TypesInfo.Types[tagKeyArg]
	if !ok {
		return ""
	}

	named, ok := tv.Type.(*types.Named)
	if !ok {
		return ""
	}

	typeArgs := named.TypeArgs()
	if typeArgs == nil || typeArgs.Len() == 0 {
		return ""
	}

	targetType := typeArgs.At(0).String()
	p.log.Debug("digen: extracted tag type from generic", "type", targetType)
	return targetType
}

// Helper function to check if a file was created by our generator
func isGeneratedFile(file *ast.File) bool {
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if comment.Text == "// Code generated by digen. DO NOT EDIT." {
				return true
			}
		}
	}
	return false
}
