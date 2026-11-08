package engine

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// camelCase matches a lowercase-first identifier with a capital inside, such as
// toView. Plain lowercase words ("finite", "stop") are not told apart from ordinary
// English, so only this shape is treated as a private name.
var camelCase = regexp.MustCompile(`^[a-z]+[A-Z]\w*$`)

// privateNames lists the unexported camelCase functions, methods, types, constants
// and variables declared in files.
func privateNames(files []*ast.File) []string {
	seen := map[string]bool{}
	add := func(id *ast.Ident) {
		if id != nil && camelCase.MatchString(id.Name) {
			seen[id.Name] = true
		}
	}

	for _, f := range files {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				add(d.Name)
			case *ast.GenDecl:
				for _, s := range d.Specs {
					switch s := s.(type) {
					case *ast.TypeSpec:
						add(s.Name)
					case *ast.ValueSpec:
						for _, n := range s.Names {
							add(n)
						}
					}
				}
			}
		}
	}

	var names []string
	for n := range seen {
		names = append(names, n)
	}
	return names
}

// docProblems lists what is wrong with the documentation of the package made of
// files: an exported name with no comment, or a comment that names a private
// identifier a reader of the documentation cannot look up.
func docProblems(fset *token.FileSet, files []*ast.File) []string {
	// Before doc.New, which cuts unexported declarations out of the files.
	private := privateNames(files)

	pkgFiles := map[string]*ast.File{}
	for _, f := range files {
		pkgFiles[fset.File(f.Pos()).Name()] = f
	}
	pkg := doc.New(&ast.Package{Name: files[0].Name.Name, Files: pkgFiles}, "example.com/"+files[0].Name.Name, 0)

	var problems []string
	check := func(what, text string) {
		if strings.TrimSpace(text) == "" {
			problems = append(problems, fmt.Sprintf("%s has no doc comment", what))
		}
		for _, n := range private {
			if regexp.MustCompile(`\b` + n + `\b`).MatchString(text) {
				problems = append(problems, fmt.Sprintf("the doc comment of %s names the private identifier %s", what, n))
			}
		}
	}
	values := func(vs []*doc.Value) {
		for _, v := range vs {
			check(strings.Join(v.Names, ", "), v.Doc)
		}
	}
	funcs := func(fs []*doc.Func) {
		for _, f := range fs {
			check(f.Name, f.Doc)
		}
	}

	check("package "+pkg.Name, pkg.Doc)
	values(pkg.Consts)
	values(pkg.Vars)
	funcs(pkg.Funcs)
	for _, t := range pkg.Types {
		check("type "+t.Name, t.Doc)
		values(t.Consts)
		values(t.Vars)
		funcs(t.Funcs)
		for _, m := range t.Methods {
			check(t.Name+"."+m.Name, m.Doc)
		}
	}
	return problems
}

func parseSource(t *testing.T, src string) (*token.FileSet, []*ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return fset, []*ast.File{f}
}

// The documentation of every package must describe each exported name and must not
// point at names a reader cannot see.
func TestDocumentation(t *testing.T) {
	for _, dir := range []string{".", "../geometry", "../mesh", "../sprite"} {
		t.Run(dir, func(t *testing.T) {
			fset := token.NewFileSet()
			pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
				return !strings.HasSuffix(fi.Name(), "_test.go")
			}, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range pkgs {
				var files []*ast.File
				for _, f := range p.Files {
					files = append(files, f)
				}
				for _, problem := range docProblems(fset, files) {
					t.Error(problem)
				}
			}
		})
	}
}

func TestDocProblemsFindsWhatItShould(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string // substrings, one per problem; none means clean
	}{
		{"clean", "// Package p is p.\npackage p\n\n// F does f.\nfunc F() {}\n", nil},
		{"function without a comment", "// Package p is p.\npackage p\n\nfunc F() {}\n", []string{"F has no doc comment"}},
		{"type without a comment", "// Package p is p.\npackage p\n\ntype T struct{}\n", []string{"type T has no doc comment"}},
		{"method without a comment", "// Package p is p.\npackage p\n\n// T is a type.\ntype T struct{}\n\nfunc (T) M() {}\n", []string{"T.M has no doc comment"}},
		{"constant group without a comment", "// Package p is p.\npackage p\n\nconst (\n\tA = 1\n\tB = 2\n)\n", []string{"A, B has no doc comment"}},
		{"variable without a comment", "// Package p is p.\npackage p\n\nvar V = 1\n", []string{"V has no doc comment"}},
		{"package without a comment", "package p\n\n// F does f.\nfunc F() {}\n", []string{"package p has no doc comment"}},
		{"unexported names need no comment", "// Package p is p.\npackage p\n\nfunc f() {}\n\ntype t struct{}\n", nil},
		{"names a private function", "// Package p is p.\npackage p\n\n// F is the inverse of toView.\nfunc F() {}\n\nfunc toView() {}\n", []string{"names the private identifier toView"}},
		{"names a private constant", "// Package p is p.\npackage p\n\n// F waits holdTimeout.\nfunc F() {}\n\nconst holdTimeout = 1\n", []string{"names the private identifier holdTimeout"}},
		{"names a private type", "// Package p is p.\npackage p\n\n// F uses a keyParser.\nfunc F() {}\n\ntype keyParser struct{}\n", []string{"names the private identifier keyParser"}},
		{"a plain word is not a private name", "// Package p is p.\npackage p\n\n// F is finite.\nfunc F() {}\n\nfunc finite() {}\n", nil},
		{"a longer word is not the private name", "// Package p is p.\npackage p\n\n// F makes toViews.\nfunc F() {}\n\nfunc toView() {}\n", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset, files := parseSource(t, tt.src)
			got := docProblems(fset, files)

			if len(got) != len(tt.want) {
				t.Fatalf("wanted %d problems %v, got %d: %v", len(tt.want), tt.want, len(got), got)
			}
			for i, w := range tt.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("wanted problem %d to contain %q, got %q", i, w, got[i])
				}
			}
		})
	}
}
