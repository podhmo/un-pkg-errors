package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

// go run github.com/podhmo/un-pkg-errors@latest -debug -replace $(git grep -l -P 'errors\.(Wrapf?|WithMessage|WithStack|New|Errorf)\(' | grep -v vendor | grep "\.go$")

func main() {
	config := Config{}

	log.SetFlags(0)
	log.SetPrefix("log::")

	flag.BoolVar(&config.debug, "debug", config.debug, "show debug log")
	flag.BoolVar(&config.replace, "replace", config.replace, "write file instead of output to stdout")
	flag.BoolVar(&config.quiet, "quiet", config.quiet, "don't printer.Fprint(os.Stdout, fset, tree)")
	flag.StringVar(&config.tmpdir, "tmpdir", ".", "tmpdir")
	flag.VisitAll(func(f *flag.Flag) {
		if v := os.Getenv(strings.ToUpper(f.Name)); v != "" {
			f.Value.Set(v)
		}
	})
	flag.Parse()
	if err := run(config, flag.Args()); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

type Config struct {
	debug   bool
	replace bool
	quiet   bool
	tmpdir  string
}

func run(config Config, files []string) error {
	fset := token.NewFileSet()
	scanner := &Scanner{fset: fset, Config: config}
	fixer := &Fixer{fset: fset, Config: config}

	for _, filename := range files {
		if strings.HasPrefix(filename, "-") {
			continue
		}

		f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		scanner.Scan(f)
	}

	for _, t := range scanner.targets {
		filename := fset.File(t.syntax.Pos()).Name()
		if !t.needFix {
			if config.debug {
				log.Printf("%-6s %s", "skip:", filename)
			}
			continue
		}
		fixer.Fix(t)

		if config.replace {
			log.Printf("%-6s %s", "write:", filename)
			wf, err := os.CreateTemp(config.tmpdir, "*.go")
			if err != nil {
				return fmt.Errorf("create file: %s, -- %w", filename, err)
			}
			if err := FprintTree(wf, fset, filename, t.syntax); err != nil {
				return err
			}
			if err := os.Rename(wf.Name(), filename); err != nil {
				return fmt.Errorf("mv file: %s -> %s, -- %w", wf.Name(), filename, err)
			}
		} else if !config.quiet {
			if err := FprintTree(os.Stdout, fset, filename, t.syntax); err != nil {
				return err
			}
		}
	}
	return nil
}

func FprintTree(w io.Writer, fset *token.FileSet, filename string, tree ast.Node) (retErr error) {
	if t, ok := w.(io.Closer); ok {
		defer func() {
			if err := t.Close(); err != nil {
				retErr = fmt.Errorf("close file....: %s, -- %w", filename, err)
			}
		}()
	}

	buf := new(bytes.Buffer)
	if err := printer.Fprint(buf, fset, tree); err != nil {
		return fmt.Errorf("format file: %s, -- %w", filename, err)
	}
	code, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format file..: %s, -- %w", filename, err)
	}
	if _, err := w.Write(code); err != nil {
		return fmt.Errorf("emit file..: %s, -- %w", filename, err)
	}
	return nil
}

type Scanner struct {
	Config

	fset    *token.FileSet
	targets []*Target
}

func (s *Scanner) Scan(f *ast.File) {
	fset := s.fset
	if s.debug {
		log.Printf("%-6s %s", "scan:", fset.File(f.Pos()).Name())
	}
	// ast.Fprint(os.Stdout, fset, f, nil)

	target := &Target{syntax: f}
	var stack []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1] // pop
		} else {
			stack = append(stack, n) // push
		}

		switch n := n.(type) {
		case *ast.CallExpr:
			// errors.Wrap() or errors.Wrapf()
			if fun, ok := n.Fun.(*ast.SelectorExpr); ok {
				if prefix, ok := fun.X.(*ast.Ident); ok && prefix.Name == "errors" {
					if s.debug {
						log.Printf("\t%-6s %10s():%d", "scan:", fun.Sel.Name, fset.File(f.Pos()).Line(n.Pos()))
					}

					switch fun.Sel.Name {
					case "Is":
						target.needFixImport = true
					case "New":
						target.needFixImport = true
						target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: nil}) // xxx
					case "Errorf":
						target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: nil}) // xxx
					case "Wrap", "Wrapf", "WithMessage", "WithMessagef", "WithStack":
						switch parent := stack[len(stack)-2].(type) {
						case *ast.ReturnStmt:
							set := func(new ast.Expr) {
								for i, x := range parent.Results {
									if x == n {
										parent.Results[i] = new
									}
								}
							}
							target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: set})
						case *ast.CallExpr:
							set := func(new ast.Expr) {
								for i, x := range parent.Args {
									if x == n {
										parent.Args[i] = new
									}
								}
							}
							target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: set})
						case *ast.ParenExpr:
							set := func(new ast.Expr) {
								parent.X = new
							}
							target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: set})
						case *ast.AssignStmt:
							set := func(new ast.Expr) {
								for i, x := range parent.Rhs {
									if x == n {
										parent.Rhs[i] = new
									}
								}
							}
							target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: set})
						case *ast.SelectorExpr:
							set := func(new ast.Expr) {
								parent.X = new
							}
							target.calls = append(target.calls, &call{name: fun.Sel.Name, expr: n, stack: stack[:], set: set})
						default:
							log.Printf("\tWARNING: unexpected type: %T in %s", parent, fun.Sel.Name)
						}
					}
				}
			}
		}
		return true
	})
	target.needFix = len(target.calls) > 0 || target.needFixImport
	s.targets = append(s.targets, target)
}

type Fixer struct {
	Config

	fset *token.FileSet
}

func (f *Fixer) Fix(target *Target) {
	fset := f.fset
	syntax := target.syntax

	log.Printf("%-6s %s", "fix:", fset.File(syntax.Pos()).Name())
	if target.needFixImport {
		for _, im := range syntax.Imports {
			if im.Path.Value == `"github.com/pkg/errors"` {
				im.Path.Value = `"errors"`
			}
		}
	}
	for _, call := range target.calls {
		pos := call.expr.Pos()
		if f.debug {
			log.Printf("\t%-6s %10s():%d", "fix:", call.name, fset.File(syntax.Pos()).Line(pos))
		}

		switch call.name {
		case "New", "Errorf":
			// errors.New(...) -> fmt.Errorf(...) // prevent using pkg/errors.New() with goimports
			// errors.Errorf(...) -> fmt.Errorf(...)
			fn := call.expr.Fun.(*ast.SelectorExpr)
			fn.X.(*ast.Ident).Name = "fmt"
			fn.Sel.Name = "Errorf"
		case "WithStack":
			call.set(call.expr.Args[0])
		case "Wrap", "Wrapf", "WithMessage", "WithMessagef": // errors.Wrap(err, "<...>") -> fmt.Errorf("<...> -- %w", err)
			errArg := call.expr.Args[0]
			fmtArg := call.expr.Args[1]

			if v, ok := fmtArg.(*ast.BasicLit); ok && v.Kind == token.STRING {
				fmtArg = &ast.BasicLit{ValuePos: v.Pos(), Kind: v.Kind, Value: strconv.Quote(strings.Trim(v.Value, "- \"") + " -- %w")}
			} else {
				v := fmtArg
				fmtArg = &ast.BinaryExpr{X: v, OpPos: v.Pos(), Op: token.ADD, Y: &ast.BasicLit{ValuePos: v.Pos(), Kind: token.STRING, Value: strconv.Quote(" -- %w")}}
			}

			args := make([]ast.Expr, 0, len(call.expr.Args)+1)
			args = append(args, fmtArg)
			args = append(args, call.expr.Args[2:]...)
			args = append(args, errArg)
			call.set(&ast.CallExpr{
				Fun:    &ast.SelectorExpr{X: &ast.Ident{NamePos: pos, Name: "fmt"}, Sel: &ast.Ident{NamePos: pos, Name: "Errorf"}},
				Lparen: pos,
				Args:   args,
				Rparen: pos,
			})
		}
	}
}

type Target struct {
	syntax *ast.File
	calls  []*call

	needFix       bool
	needFixImport bool
}
type call struct {
	name string

	expr  *ast.CallExpr
	stack []ast.Node
	set   func(ast.Expr)
}
