package main

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateMockFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")

	// Atualizamos o conteúdo da interface do teste.
	// Adicionamos o tipo customizado 'MyCustomType' e fazemos a interface usá-lo.
	content := `package sample

import (
	"io/fs"
	"os"
)

type MyCustomType struct {
	Name string
}

type IExample interface {
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
	NoOp()
	Multi(a int, b string) (bool, error)
	DirEntries(f *os.File, n int) ([]fs.DirEntry, error)

	// PONTO 1: 'param MyCustomType' fará o primeiro 'if !isBuiltinType' ser TRUE
	CustomParam(param MyCustomType)

	// PONTO 2: O retorno 'MyCustomType' fará o segundo 'if !isBuiltinType' ser TRUE no finalNode
	CustomReturn() MyCustomType

	// COBERTURA DO PONTO 1: Parâmetro variádico NOMEADO
	VariadicNamed(prefix string, args ...string)

	// COBERTURA DO PONTO 2: Parâmetro variádico ANÔNIMO (Sem nome de variável)
	VariadicAnonymous(...int)
}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Cria um go.mod fake temporário no diretório de execução do teste para o getModulePath não quebrar
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		if err := os.WriteFile("go.mod", []byte("module ://github.com"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove("go.mod")
	}

	outputPath, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
		Alias:         "Example",
		OutputPath:    filepath.Join(tmpDir, "outdir"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(outputPath, "/example.go") && !strings.HasSuffix(outputPath, "\\example.go") {
		t.Fatalf("expected generated file name to end with example.go, got %s", outputPath)
	}

	generated, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	text := string(generated)
	if !strings.Contains(text, "type MockExample struct") {
		t.Fatal("generated code missing MockExample definition")
	}

	// VALIDAÇÃO EXTRA: Garante que o gerador aplicou corretamente o prefixo 'sample.' nos tipos locais
	if !strings.Contains(text, "CustomParam(param sample.MyCustomType)") {
		t.Fatal("generated code missing sample. prefix on custom parameter type")
	}
	if !strings.Contains(text, "CustomReturn() sample.MyCustomType") {
		t.Fatal("generated code missing sample. prefix on custom return type")
	}
}

func TestParseSourceFile_ParsesImports(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	content := `package sample

import (
	ioAlias "io"
	"os"
)
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	file, imports, err := parseSourceFile(inputPath, []byte(content))
	if err != nil {
		t.Fatal(err)
	}
	if file.Name.Name != "sample" {
		t.Fatalf("expected package name sample, got %s", file.Name.Name)
	}
	if got, ok := imports["ioAlias"]; !ok || got != "io" {
		t.Fatalf("expected import alias ioAlias => io, got %v", imports)
	}
	if got, ok := imports["os"]; !ok || got != "os" {
		t.Fatalf("expected import alias os => os, got %v", imports)
	}
}

func TestParseSourceFile_ParseFailure(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("package sample\nimport (unquoted)")
	_, _, err := parseSourceFile(filepath.Join(tmpDir, "bad.go"), content)
	if err == nil {
		t.Fatal("expected parse error for invalid import syntax")
	}
}

func TestParseSourceFile_InvalidRawStringImportPath(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("package sample\nimport `io`\n")
	_, _, err := parseSourceFile(filepath.Join(tmpDir, "raw.go"), content)
	if err == nil || !strings.Contains(err.Error(), "invalid import path") {
		t.Fatalf("expected invalid import path error for raw string import, got %v", err)
	}
}

func TestStrconvUnquote_Error(t *testing.T) {
	_, err := strconvUnquote("notquoted")
	if err == nil || !strings.Contains(err.Error(), "invalid import path") {
		t.Fatalf("expected invalid import path error, got %v", err)
	}
}

func TestGenerateMockFile_MissingInputFile(t *testing.T) {
	_, err := GenerateMockFile(GeneratorOptions{
		InterfaceName: "IExample",
	})
	if err == nil || !strings.Contains(err.Error(), "input file is required") {
		t.Fatalf("expected input file required error, got %v", err)
	}
}

func TestGenerateMockFile_MissingInterfaceName(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	if err := os.WriteFile(inputPath, []byte("package sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile: inputPath,
	})
	if err == nil || !strings.Contains(err.Error(), "interface name is required") {
		t.Fatalf("expected interface name required error, got %v", err)
	}
}

func TestGenerateMockFile_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     filepath.Join(tmpDir, "doesnotexist.go"),
		InterfaceName: "IExample",
	})
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected read file error, got %v", err)
	}
}

func TestGenerateMockFile_ParseError(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bad.go")
	if err := os.WriteFile(inputPath, []byte("package sample\ntype IExample interface {"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
	})
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestGenerateMockFile_InterfaceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	if err := os.WriteFile(inputPath, []byte("package sample\n\ntype IOther interface { NoOp() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
	})
	if err == nil || !strings.Contains(err.Error(), "interface IExample not found") {
		t.Fatalf("expected interface not found error, got %v", err)
	}
}

func TestFindInterface_TypeIsNotInterface(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	content := []byte("package sample\n\ntype IExample struct { Name string }\n")
	if err := os.WriteFile(inputPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	file, _, err := parseSourceFile(inputPath, content)
	if err != nil {
		t.Fatal(err)
	}

	_, err = findInterface(file, "IExample")
	if err == nil || !strings.Contains(err.Error(), "type IExample is not an interface") {
		t.Fatalf("expected type is not an interface error, got %v", err)
	}
}

func TestBuildMethods_UnsupportedInterfaceMember(t *testing.T) {
	iface := &ast.InterfaceType{
		Methods: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{{Name: "Method"}},
			Type:  ast.NewIdent("int"),
		}}},
	}

	// Adicionado o parâmetro "sample" exigido pelo buildMethods atualizado
	_, _, err := buildMethods(iface, "sample", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "unsupported interface member Method") {
		t.Fatalf("expected unsupported interface member error, got %v", err)
	}
}

func TestBuildMethods_CollectImportsError(t *testing.T) {
	iface := &ast.InterfaceType{
		Methods: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{{Name: "Method"}},
			Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{
				Type: &ast.SelectorExpr{X: ast.NewIdent("pkg"), Sel: ast.NewIdent("Type")},
			}}}},
		}}},
	}

	// Adicionado o parâmetro "sample" exigido pelo buildMethods atualizado
	_, _, err := buildMethods(iface, "sample", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "unknown selector package pkg") {
		t.Fatalf("expected unknown selector package error, got %v", err)
	}
}

func TestFormatParams_NilList(t *testing.T) {
	// Adicionado o parâmetro "sample"
	params, args := formatParams(nil, "sample")
	if params != "" || args != nil {
		t.Fatalf("expected empty params and nil args for nil list, got %q %v", params, args)
	}
}

func TestFormatParams_AnonymousParameter(t *testing.T) {
	list := &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("string")}}}
	params, args := formatParams(list, "sample")
	if params != "string" {
		t.Fatalf("expected params string, got %q", params)
	}
	if len(args) != 1 || args[0] != "arg0" {
		t.Fatalf("expected anonymous arg arg0, got %v", args)
	}
}

func TestFormatParams_AnonymousVariadic(t *testing.T) {
	// Cria um nó representando ...int na AST manualmente
	list := &ast.FieldList{List: []*ast.Field{{
		Type: &ast.Ellipsis{Elt: ast.NewIdent("int")},
	}}}

	params, args := formatParams(list, "sample")
	if !strings.Contains(params, "int") {
		t.Fatalf("expected params to contain int, got %q", params)
	}
	if len(args) != 1 || args[0] != "arg0..." {
		t.Fatalf("expected arg0... for anonymous variadic parameter, got %v", args)
	}
}

func TestFormatResults_SingleResult(t *testing.T) {
	list := &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("error")}}}
	// Adicionado o parâmetro "sample"
	results, types := formatResults(list, "sample")
	if results != "error" {
		t.Fatalf("expected single result string error, got %q", results)
	}
	if len(types) != 1 || types[0] != "error" {
		t.Fatalf("expected result types [error], got %v", types)
	}
}

func TestDeriveSetReturn_SingleResultName(t *testing.T) {
	signature, values := deriveSetReturn([]string{"string"})
	if signature != "result string" || values != "result" {
		t.Fatalf("expected result string / result, got %q / %q", signature, values)
	}

	signature, values = deriveSetReturn([]string{"error"})
	if signature != "err error" || values != "err" {
		t.Fatalf("expected err error / err, got %q / %q", signature, values)
	}
}

func TestRenderMock_ImportsWithAlias(t *testing.T) {
	// Adicionado o parâmetro final "sample/file.go" na chamada da renderMock
	code, err := renderMock("mocks", "sample", "Example", nil, map[string]string{"ioAlias": "io"}, "sample/file.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(code), "ioAlias \"io\"") {
		t.Fatalf("expected aliased import line, got: %s", string(code))
	}
}

func TestResolveOutputPath_DefaultOutput(t *testing.T) {
	output, err := resolveOutputPath("", "Example")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join("internal", "pkgxmock", "example.go")
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestResolveOutputPath_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	output, err := resolveOutputPath(tmpDir, "Example")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(tmpDir, "example.go")
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestResolveOutputPath_StatErrorAfterExists(t *testing.T) {
	originalStat := osStat
	defer func() { osStat = originalStat }()

	callCount := 0
	osStat = func(name string) (os.FileInfo, error) {
		callCount++
		if callCount == 1 {
			return originalStat(name)
		}
		return nil, fmt.Errorf("stat failed")
	}

	tmpDir := t.TempDir()
	_, err := resolveOutputPath(tmpDir, "Example")
	if err == nil || !strings.Contains(err.Error(), "stat failed") {
		t.Fatalf("expected stat failed error, got %v", err)
	}
}

func TestPathExists_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	exists, err := pathExists(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("expected existing path to return true")
	}
}

func TestGenerateMockFile_EmbeddedInterfaceNotSupported(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	content := `package sample

import "io"

type IExample interface {
	io.Closer
}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
	})
	if err == nil || !strings.Contains(err.Error(), "embedded interfaces are not supported") {
		t.Fatalf("expected embedded interface error, got %v", err)
	}
}

func TestGenerateMockFile_RenderMockError(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "sample.go")
	content := `package sample

type IExample interface {
	NoOp()
}
`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
		Alias:         "Bad-Alias",
	})
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("expected render error, got %v", err)
	}
}

func TestGenerateMockFile_ResolveOutputPathError(t *testing.T) {
	tmpDir := t.TempDir()
	symlinkPath := filepath.Join(tmpDir, "broken")
	if err := os.Symlink(filepath.Join(tmpDir, "missing"), symlinkPath); err != nil {
		t.Fatal(err)
	}

	inputPath := filepath.Join(tmpDir, "sample.go")
	if err := os.WriteFile(inputPath, []byte("package sample\ntype IExample interface { NoOp() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
		OutputPath:    symlinkPath,
	})
	if err == nil {
		t.Fatal("expected resolve output path error")
	}
}

func TestGenerateMockFile_MkdirAllError(t *testing.T) {
	tmpDir := t.TempDir()
	badParent := filepath.Join(tmpDir, "badparent")
	if err := os.WriteFile(badParent, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	inputPath := filepath.Join(tmpDir, "sample.go")
	if err := os.WriteFile(inputPath, []byte("package sample\ntype IExample interface { NoOp() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(badParent, "example.go")
	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
		OutputPath:    outputPath,
	})
	if err == nil {
		t.Fatal("expected mkdirall error")
	}
}

func TestGenerateMockFile_WriteFileError(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Prepara o arquivo de interface de entrada válido
	inputPath := filepath.Join(tmpDir, "sample.go")
	if err := os.WriteFile(inputPath, []byte("package sample\ntype IExample interface { NoOp() }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Criamos um nome de arquivo gigante (maior que 255 caracteres)
	// strings.Repeat cria uma string longa de letras 'a' terminando em .go
	longFileName := strings.Repeat("a", 260) + ".go"

	// O diretório pai é válido (.../tmpDir/outdir), então o MkdirAll vai passar com sucesso!
	badOutputPath := filepath.Join(tmpDir, "outdir", longFileName)

	// 3. Executa o gerador
	_, err := GenerateMockFile(GeneratorOptions{
		InputFile:     inputPath,
		InterfaceName: "IExample",
		Alias:         "Example",
		OutputPath:    badOutputPath,
	})

	// O MkdirAll cria a pasta 'outdir' normalmente (OK).
	// Mas o WriteFile tenta criar o arquivo com 260 caracteres e o Linux barra com "file name too long"!
	if err == nil {
		t.Fatal("expected write file error because the file name is too long for the OS")
	}
}

func TestGetModulePath_NoModuleDeclaration(t *testing.T) {
	// 1. Criamos um arquivo go.mod temporário inválido (sem a linha do module)
	tmpDir := t.TempDir()
	fakeGoMod := filepath.Join(tmpDir, "go.mod")
	content := []byte("// This file does not contain a module declaration\ngo 1.26\n")
	if err := os.WriteFile(fakeGoMod, content, 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Como a função 'getModulePath' lê o go.mod do diretório atual de execução,
	// nós mudamos temporariamente o diretório de trabalho do teste para a pasta temporária.
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	// Garante que o teste vai restaurar a pasta original quando terminar
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// 3. Executa a função interna
	path := getModulePath()

	// Como o arquivo foi lido até o final sem achar "module ",
	// o fluxo vai bater exatamente no return "" final que faltava cobrir!
	if path != "" {
		t.Fatalf("expected empty module path, got %q", path)
	}
}
