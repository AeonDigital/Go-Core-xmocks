package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var osStat = os.Stat

type GeneratorOptions struct {
	InputFile     string
	InterfaceName string
	Alias         string
	OutputPath    string
}

type MethodInfo struct {
	Name             string
	SignatureParams  string
	SignatureResults string
	CallArguments    string
	FuncFieldType    string
	SetReturnParams  string
	ReturnArguments  string
}

func GenerateMockFile(opts GeneratorOptions) (string, error) {
	if opts.InputFile == "" {
		return "", fmt.Errorf("input file is required")
	}
	if opts.InterfaceName == "" {
		return "", fmt.Errorf("interface name is required")
	}
	if opts.Alias == "" {
		opts.Alias = opts.InterfaceName
	}

	src, err := os.ReadFile(opts.InputFile)
	if err != nil {
		return "", err
	}

	srcFile, imports, err := parseSourceFile(opts.InputFile, src)
	if err != nil {
		return "", err
	}

	iface, err := findInterface(srcFile, opts.InterfaceName)
	if err != nil {
		return "", err
	}

	methods, usedImports, err := buildMethods(iface, srcFile.Name.Name, imports)
	if err != nil {
		return "", err
	}

	outputPath, err := resolveOutputPath(opts.OutputPath, opts.Alias)
	if err != nil {
		return "", err
	}

	dirName := filepath.Base(filepath.Dir(outputPath))
	packageName := "mocks"
	if dirName != "." && dirName != "" {
		packageName = dirName
	}

	code, err := renderMock(packageName, srcFile.Name.Name, opts.Alias, methods, usedImports, opts.InputFile)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}

	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		return "", err
	}

	return outputPath, nil
}

func parseSourceFile(filename string, src []byte) (*ast.File, map[string]string, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	imports := map[string]string{}
	for _, spec := range parsed.Imports {
		importPath, err := strconvUnquote(spec.Path.Value)
		if err != nil {
			return nil, nil, err
		}
		alias := ""
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		if alias == "" {
			alias = path.Base(importPath)
		}
		imports[alias] = importPath
	}

	return parsed, imports, nil
}

func strconvUnquote(value string) (string, error) {
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1], nil
	}
	return "", fmt.Errorf("invalid import path %s", value)
}

func findInterface(file *ast.File, name string) (*ast.InterfaceType, error) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			if typeSpec.Name.Name != name {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("type %s is not an interface", name)
			}
			return iface, nil
		}
	}
	return nil, fmt.Errorf("interface %s not found", name)
}

func buildMethods(iface *ast.InterfaceType, sourcePackage string, imports map[string]string) ([]MethodInfo, map[string]string, error) {
	methods := []MethodInfo{}
	usedImports := map[string]string{}

	for _, field := range iface.Methods.List {
		if len(field.Names) == 0 {
			return nil, nil, fmt.Errorf("embedded interfaces are not supported")
		}
		methodName := field.Names[0].Name
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			return nil, nil, fmt.Errorf("unsupported interface member %s", methodName)
		}

		if err := collectImports(funcType, imports, usedImports); err != nil {
			return nil, nil, err
		}

		params, argNames := formatParams(funcType.Params, sourcePackage)
		results, resultTypes := formatResults(funcType.Results, sourcePackage)
		setReturnParams, returnArgs := deriveSetReturn(resultTypes)

		funcTypeString := renderFuncType(funcType, sourcePackage)
		methods = append(methods, MethodInfo{
			Name:             methodName,
			SignatureParams:  params,
			SignatureResults: results,
			CallArguments:    strings.Join(argNames, ", "),
			FuncFieldType:    funcTypeString,
			SetReturnParams:  setReturnParams,
			ReturnArguments:  returnArgs,
		})
	}

	return methods, usedImports, nil
}

func collectImports(expr ast.Node, imports map[string]string, used map[string]string) error {
	var foundErr error
	ast.Inspect(expr, func(node ast.Node) bool {
		if foundErr != nil {
			return false
		}
		switch t := node.(type) {
		case *ast.SelectorExpr:
			if pkgIdent, ok := t.X.(*ast.Ident); ok {
				if importPath, found := imports[pkgIdent.Name]; found {
					used[pkgIdent.Name] = importPath
				} else {
					foundErr = fmt.Errorf("unknown selector package %s", pkgIdent.Name)
					return false
				}
			}
		}
		return true
	})
	return foundErr
}

func formatParams(list *ast.FieldList, sourcePackage string) (string, []string) {
	if list == nil {
		return "", nil
	}

	parts := []string{}
	args := []string{}
	index := 0
	for _, field := range list.List {
		typeStr := nodeString(field.Type, sourcePackage)

		// Verifica se o parâmetro original é variádico (ex: ...string)
		isVariadic := strings.HasPrefix(typeStr, "...")

		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
			argName := fmt.Sprintf("arg%d", index)
			if isVariadic {
				argName += "..."
			}
			args = append(args, argName)
			index++
			continue
		}
		names := []string{}
		for _, name := range field.Names {
			names = append(names, name.Name)
			argName := name.Name
			if isVariadic {
				argName += "..."
			}
			args = append(args, argName)
			index++
		}
		parts = append(parts, fmt.Sprintf("%s %s", strings.Join(names, ", "), typeStr))
	}
	return strings.Join(parts, ", "), args
}

func formatResults(list *ast.FieldList, sourcePackage string) (string, []string) {
	if list == nil {
		return "", nil
	}

	parts := []string{}
	resultTypes := []string{}
	for _, field := range list.List {
		typeStr := nodeString(field.Type, sourcePackage)
		resultTypes = append(resultTypes, typeStr)
		if len(list.List) == 1 {
			parts = append(parts, typeStr)
			break
		}
		parts = append(parts, typeStr)
	}

	if len(parts) == 1 {
		return parts[0], resultTypes
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, ", ")), resultTypes
}

func deriveSetReturn(resultTypes []string) (string, string) {
	if len(resultTypes) == 0 {
		return "", ""
	}

	parts := []string{}
	vars := []string{}
	for index, typ := range resultTypes {
		name := fmt.Sprintf("result%d", index)
		if len(resultTypes) == 1 {
			if typ == "error" {
				name = "err"
			} else {
				name = "result"
			}
		} else if index == len(resultTypes)-1 && typ == "error" {
			name = "err"
		}
		parts = append(parts, fmt.Sprintf("%s %s", name, typ))
		vars = append(vars, name)
	}

	return strings.Join(parts, ", "), strings.Join(vars, ", ")
}

func renderFuncType(funcType *ast.FuncType, sourcePackage string) string {
	params, _ := formatParams(funcType.Params, sourcePackage)
	results, _ := formatResults(funcType.Results, sourcePackage)
	if results == "" {
		return fmt.Sprintf("func(%s)", params)
	}
	return fmt.Sprintf("func(%s) %s", params, results)
}

func nodeString(node ast.Node, sourcePackage string) string {
	finalNode := node

	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if !isBuiltinType(ident.Name) {
				if node == ident {
					finalNode = &ast.SelectorExpr{
						X:   ast.NewIdent(sourcePackage),
						Sel: ident,
					}
				} else {
					_ = ident.Name
				}
			}
		}
		return true
	})

	if ident, ok := node.(*ast.Ident); ok {
		if !isBuiltinType(ident.Name) {
			finalNode = &ast.SelectorExpr{
				X:   ast.NewIdent(sourcePackage),
				Sel: ident,
			}
		}
	}

	var buf bytes.Buffer
	_ = printer.Fprint(&buf, token.NewFileSet(), finalNode)
	return buf.String()
}

func isBuiltinType(name string) bool {
	builtins := map[string]bool{
		"string": true, "error": true, "int": true, "int64": true, "int32": true,
		"uint": true, "uint64": true, "uint32": true, "bool": true, "byte": true,
		"rune": true, "float64": true, "float32": true, "uintptr": true, "any": true,
	}
	return builtins[name]
}

func getModulePath() string {
	file, err := os.Open("go.mod")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	scanner.Err()

	return ""
}

func renderMock(
	packageName string,
	sourcePackage string,
	alias string,
	methods []MethodInfo,
	usedImports map[string]string,
	inputFile string,
) ([]byte, error) {
	imports := map[string]string{}
	imports["fmt"] = "fmt"
	imports["runtime"] = "runtime"
	for aliasName, importPath := range usedImports {
		imports[aliasName] = importPath
	}

	if packageName != sourcePackage {
		modPath := getModulePath()
		if modPath != "" {
			subFolder := filepath.Dir(inputFile)
			subFolder = filepath.ToSlash(subFolder)
			subFolder = strings.TrimPrefix(subFolder, "./")

			needParentImport := false
			for _, method := range methods {
				if strings.Contains(method.SignatureParams, sourcePackage+".") ||
					strings.Contains(method.SignatureResults, sourcePackage+".") ||
					strings.Contains(method.FuncFieldType, sourcePackage+".") {
					needParentImport = true
					break
				}
			}

			if needParentImport {
				imports[sourcePackage] = modPath + "/" + subFolder
			}
		}
	}

	importLines := []string{}
	for aliasName, importPath := range imports {
		if aliasName == path.Base(importPath) {
			importLines = append(importLines, fmt.Sprintf("\t\"%s\"", importPath))
		} else {
			importLines = append(importLines, fmt.Sprintf("\t%s \"%s\"", aliasName, importPath))
		}
	}
	sort.Strings(importLines)

	var buf strings.Builder
	fmt.Fprintf(&buf, "// Code generated by xmocks; DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	fmt.Fprintf(&buf, "import (\n")
	for _, line := range importLines {
		fmt.Fprintf(&buf, "%s", line)
		fmt.Fprintf(&buf, "\n")
	}
	fmt.Fprintf(&buf, ")\n\n")

	fmt.Fprintf(&buf, "type TestCase%s struct {\n", alias)
	fmt.Fprintf(&buf, "\tName string\n")
	fmt.Fprintf(&buf, "\tEnv map[string]string\n")
	fmt.Fprintf(&buf, "\tInput any\n")
	fmt.Fprintf(&buf, "\tWant any\n")
	fmt.Fprintf(&buf, "\tWantErr bool\n")
	fmt.Fprintf(&buf, "\tErrContains string\n")
	fmt.Fprintf(&buf, "\tMockFn func(m *Mock%s)\n", alias)
	fmt.Fprintf(&buf, "}\n\n")

	fmt.Fprintf(&buf, "func NewMock%s() *Mock%s {\n", alias, alias)
	fmt.Fprintf(&buf, "\tmock := &Mock%s{}\n", alias)
	fmt.Fprintf(&buf, "\tmock.OnCall = &mockOnCall%s{mock: mock}\n", alias)
	fmt.Fprintf(&buf, "\tmock.SetReturn = &mockSetReturn%s{mock: mock}\n", alias)
	fmt.Fprintf(&buf, "\treturn mock\n")
	fmt.Fprintf(&buf, "}\n\n")

	fmt.Fprintf(&buf, "type Mock%s struct {\n", alias)
	for _, method := range methods {
		fmt.Fprintf(&buf, "\t%sFunc %s\n", method.Name, method.FuncFieldType)
	}
	fmt.Fprintf(&buf, "\tOnCall *mockOnCall%s\n", alias)
	fmt.Fprintf(&buf, "\tSetReturn *mockSetReturn%s\n", alias)
	fmt.Fprintf(&buf, "}\n\n")

	fmt.Fprintf(&buf, "func (m *Mock%s) panicIfNotConfigured() {\n", alias)
	fmt.Fprintf(&buf, "\tpc, _, _, ok := runtime.Caller(2)\n")
	fmt.Fprintf(&buf, "\tmethodName := \"UnknownMethod\"\n")
	fmt.Fprintf(&buf, "\tif ok {\n")
	fmt.Fprintf(&buf, "\t\tfullFnName := runtime.FuncForPC(pc).Name()\n")
	fmt.Fprintf(&buf, "\t\tshortName := fullFnName\n")
	fmt.Fprintf(&buf, "\t\tfor i := len(fullFnName) - 1; i >= 0; i-- {\n")
	fmt.Fprintf(&buf, "\t\t\tif fullFnName[i] == '/' {\n")
	fmt.Fprintf(&buf, "\t\t\t\tshortName = fullFnName[i+1:]\n")
	fmt.Fprintf(&buf, "\t\t\t\tbreak\n")
	fmt.Fprintf(&buf, "\t\t\t}\n")
	fmt.Fprintf(&buf, "\t\t}\n")
	fmt.Fprintf(&buf, "\t\tmethodName = shortName\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tpanic(fmt.Sprintf(\"CRITICAL: Mock for %%s not configured.\", methodName))\n")
	fmt.Fprintf(&buf, "}\n\n")

	for _, method := range methods {
		fmt.Fprintf(&buf, "func (oMock *Mock%s) %s(%s) %s {\n", alias, method.Name, method.SignatureParams, method.SignatureResults)
		fmt.Fprintf(&buf, "\tif oMock.%sFunc == nil {\n", method.Name)
		fmt.Fprintf(&buf, "\t\toMock.panicIfNotConfigured()\n")
		fmt.Fprintf(&buf, "}\n")
		if method.SignatureResults == "" {
			fmt.Fprintf(&buf, "\t oMock.%sFunc(%s)\n", method.Name, method.CallArguments)
		} else {
			fmt.Fprintf(&buf, "\treturn oMock.%sFunc(%s)\n", method.Name, method.CallArguments)
		}
		fmt.Fprintf(&buf, "}\n\n")
	}

	fmt.Fprintf(&buf, "type mockOnCall%s struct {\n", alias)
	fmt.Fprintf(&buf, "\tmock *Mock%s\n", alias)
	fmt.Fprintf(&buf, "}\n\n")

	for _, method := range methods {
		fmt.Fprintf(&buf, "func (o *mockOnCall%s) %s(fn %s) {\n", alias, method.Name, method.FuncFieldType)
		fmt.Fprintf(&buf, "\to.mock.%sFunc = fn\n", method.Name)
		fmt.Fprintf(&buf, "}\n\n")
	}

	fmt.Fprintf(&buf, "type mockSetReturn%s struct {\n", alias)
	fmt.Fprintf(&buf, "\tmock *Mock%s\n", alias)
	fmt.Fprintf(&buf, "}\n\n")

	for _, method := range methods {
		fmt.Fprintf(&buf, "func (r *mockSetReturn%s) %s(%s) {\n", alias, method.Name, method.SetReturnParams)
		fmt.Fprintf(&buf, "\tr.mock.OnCall.%s(func(%s) %s {\n", method.Name, method.SignatureParams, method.SignatureResults)
		if method.ReturnArguments != "" {
			fmt.Fprintf(&buf, "\t\treturn %s\n", method.ReturnArguments)
		}
		fmt.Fprintf(&buf, "\t})\n")
		fmt.Fprintf(&buf, "}\n\n")
	}

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func resolveOutputPath(output string, alias string) (string, error) {
	if output == "" {
		return filepath.Join("internal", "pkgxmock", strings.ToLower(alias)+".go"), nil
	}

	exists, err := pathExists(output)
	if err != nil {
		return "", err
	}

	if exists {
		info, err := osStat(output)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return filepath.Join(output, strings.ToLower(alias)+".go"), nil
		}
	}

	if strings.HasSuffix(output, ".go") {
		return output, nil
	}

	return filepath.Join(output, strings.ToLower(alias)+".go"), nil
}

func pathExists(filePath string) (bool, error) {
	_, err := osStat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
