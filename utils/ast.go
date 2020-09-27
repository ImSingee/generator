package utils

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var fileSet *token.FileSet

type Field struct {
	Name string // 字段名
	Type string // 字段类型

	GetterName         string // Getter 的名称
	GetterAlreadyExist bool   // Getter 是否在原本的代码中就存在，含同名 field 已经存在的情况
	WillGenerateGetter bool   // 是否要生成 Getter

	SetterName         string
	SetterAlreadyExist bool
	WillGenerateSetter bool

	IsPublic  bool
	HasGetter bool

	IgnoreReason string
}

type Fields map[string]*Field

type Struct struct {
	Name          string // 结构体名称
	ShortName     string // 生成的函数中用于引用结构体的名称
	LowerName     string
	IsPresent     bool   // 结构体在包中存在
	PublicFields  Fields // 结构体包含的成员
	PrivateFields Fields
	IgnoreFields  Fields

	ImportedStatements string // 这个 struct 定义可能需要依赖的导入语句
}

type Structs map[string]*Struct

func GetFieldsFromStruct(structType *ast.StructType) (privateFields Fields, publicFields Fields, ignoreFields Fields, err error) {
	privateFields = make(Fields, 0)
	publicFields = make(Fields, 0)
	ignoreFields = make(Fields, 0)

	for _, field := range structType.Fields.List {

		for _, name := range field.Names {
			if ShouldIgnore(name.Name) {
				ignoreFields[name.Name] = &Field{
					Name:         name.Name,
					IgnoreReason: "name is invalid",
				}

				continue
			}

			// type 的内容
			startPosition := fileSet.Position(field.Type.Pos())
			endPosition := fileSet.Position(field.Type.End())

			file, err := os.Open(startPosition.Filename)

			if err != nil {
				return nil, nil, nil, fmt.Errorf("cannot read file %s: %w", startPosition.Filename, err)
			}

			defer file.Close()

			b := make([]byte, endPosition.Offset-startPosition.Offset)
			_, err = file.ReadAt(b, int64(startPosition.Offset))

			if err != nil {
				return nil, nil, nil, fmt.Errorf(
					"cannot read file %s in position [%d, %d): %w",
					startPosition.Filename, startPosition.Offset, endPosition.Offset, err,
				)
			}

			theField := &Field{
				Name:     name.Name,
				Type:     string(b),
				IsPublic: IsPublic(name.Name),
			}

			theField.WillGenerateGetter = true
			theField.SetterName = "" // TODO

			if theField.IsPublic {
				theField.WillGenerateGetter = false

				publicFields[name.Name] = theField
			} else {
				theField.WillGenerateGetter = true
				theField.GetterName = toGetterName(name.Name)

				privateFields[name.Name] = theField
			}
		}
	}

	return
}

func GetStructsFromFile(astFile *ast.File) (Structs, error) {
	structs := make(Structs, 0)

	// 依赖的导入的内容
	var f *os.File
	var err error
	importedStatementsBuilder := bytes.NewBuffer(make([]byte, 0, 128))
	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		startPos := fileSet.Position(genDecl.Pos())
		endPos := fileSet.Position(genDecl.End())

		if f == nil {
			f, err = os.Open(startPos.Filename)
			if err != nil {
				return nil, fmt.Errorf("cannot read file %s: %w", startPos.Filename, err)
			}
			defer f.Close()
		}

		line := make([]byte, endPos.Offset-startPos.Offset)

		_, _ = f.Seek(0, 0)
		_, err := f.ReadAt(line, int64(startPos.Offset))

		if err != nil {
			return nil, fmt.Errorf("cannot read file %s in [%d, %d): %w", startPos.Filename, startPos.Offset, endPos.Offset, err)
		}

		importedStatementsBuilder.Write(line)
		importedStatementsBuilder.WriteByte('\n')
	}

	importedStatements := importedStatementsBuilder.String()

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)

		if ok && genDecl.Tok == token.TYPE { // type 定义
			for _, spec := range genDecl.Specs {
				// 确保是 type 定义
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				// 确保是 struct 定义
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				name := typeSpec.Name.Name

				if ShouldIgnore(name) {
					return nil, fmt.Errorf("struct name %s is invalid", name)
				}

				shortName, err := GetShortName(name)

				if err != nil {
					return nil, fmt.Errorf("cannot get shortName for %s: %w", name, err)
				}

				privateFields, publicFields, ignoreFields, err := GetFieldsFromStruct(structType)

				structs[name] = &Struct{
					Name:               name,
					ShortName:          shortName,
					LowerName:          strings.ToLower(name),
					IsPresent:          true,
					PublicFields:       publicFields,
					PrivateFields:      privateFields,
					IgnoreFields:       ignoreFields,
					ImportedStatements: importedStatements,
				}
			}
		}
	}

	return structs, nil
}

func GetStructsFromPackage() (Structs, error) {
	pkgName := viper.GetString("gopackage")
	if pkgName == "" {
		return nil, fmt.Errorf("missing package name (gopackage config)")
	}

	// 获取包中所有的 Go 文件
	pkgInfo, err := build.ImportDir(".", 0)
	if err != nil {
		return nil, fmt.Errorf("cannot build from package: %w", err)
	}

	// 检查传递的 gofile 是否在包中，不在则报错

	filename := viper.GetString("gofile")
	found := false

	for _, existFilename := range pkgInfo.GoFiles {
		if filename == existFilename {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("cannot find provided filename %s", filename)
	}

	fileSet = token.NewFileSet()

	// 检查是否传递了要设置 Getter 的列表，未设置则遍历当前文件的结构体定义来设置
	structNames := viper.GetStringSlice("struct")
	var structs Structs

	if len(structNames) == 0 { // 获取当前文件中的所有 struct
		f, err := parser.ParseFile(fileSet, filename, nil, 0)

		if err != nil {
			return nil, fmt.Errorf("cannot parse file %s: %w", filename, err)
		}

		structs, err = GetStructsFromFile(f)

		if err != nil {
			return nil, fmt.Errorf("cannot get structs from file %s: %w", filename, err)
		}
	} else {
		structs = make(Structs, len(structNames))
		for _, structName := range structNames {
			structs[structName] = nil
		}

		for _, file := range pkgInfo.GoFiles {
			f, err := parser.ParseFile(fileSet, file, nil, 0)

			if err != nil {
				return nil, fmt.Errorf("cannot parse file %s: %w", filename, err)
			}

			tempStructs, err := GetStructsFromFile(f)
			if err != nil {
				return nil, fmt.Errorf("cannot get structs from file %s: %w", file, err)
			}

			for tempStructName, tempStruct := range tempStructs {
				if _, ok := structs[tempStructName]; ok {
					structs[tempStructName] = tempStruct
				}
			}
		}

		for name, s := range structs {
			if s == nil {
				return nil, fmt.Errorf("cannot get struct %s from package", name)
			}
		}
	}

	return structs, nil
}
