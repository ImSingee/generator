package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
)

type Field struct {
	Name string // 字段名

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
	IsPresent     bool   // 结构体在包中存在
	PublicFields  Fields // 结构体包含的成员
	PrivateFields Fields
	IgnoreFields  Fields
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

			theField := &Field{
				Name:     name.Name,
				IsPublic: IsPublic(name.Name),
			}

			// TODO: Setter

			if theField.IsPublic {
				theField.WillGenerateGetter = true
				theField.GetterName, _ = ToGetterName(name.Name)

				publicFields[name.Name] = theField
			} else {
				theField.WillGenerateGetter = false

				privateFields[name.Name] = theField
			}
		}
	}

	return
}

func GetStructsFromFile(astFile *ast.File) (Structs, error) {
	structs := make(Structs, 0)

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
				shortName, err := GetShortName(name)

				if err != nil {
					return nil, fmt.Errorf("cannot get shortName for %s: %w", name, err)
				}

				privateFields, publicFields, ignoreFields, err := GetFieldsFromStruct(structType)

				structs[name] = &Struct{
					Name:          name,
					ShortName:     shortName,
					IsPresent:     true,
					PublicFields:  publicFields,
					PrivateFields: privateFields,
					IgnoreFields:  ignoreFields,
				}
			}
		}
	}

	return structs, nil
}

func GetStructsFromPackage() (interface{}, error) {
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

	fileSet := token.NewFileSet()

	// 检查是否传递了要设置 Getter 的列表，未设置则遍历当前文件的结构体定义来设置
	structNames := viper.GetStringSlice("struct")

	if len(structNames) == 0 {
		f, err := parser.ParseFile(fileSet, filename, nil, 0)

		if err != nil {
			return nil, fmt.Errorf("cannot parse file %s: %w", filename, err)
		}

		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)

			if ok && genDecl.Tok == token.TYPE { // type 定义
				for _, spec := range genDecl.Specs {
					// 确保是 type 定义
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					_, ok = typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					// 获取结构体名称
					structNames = append(structNames, typeSpec.Name.Name)
				}
			}
		}
	}

	fmt.Printf("Will process %d structs %+v", len(structNames), structNames)

	structs := make(Structs, len(structNames))

	for _, name := range structNames {
		shortName, err := GetShortName(name)

		if err != nil {
			return nil, fmt.Errorf("cannot get shortName: %w", shortName)
		}

		structs[name] = &Struct{
			Name:      name,
			ShortName: shortName,
		}
	}

	// 遍历整个包，存储结构体

	for _, file := range pkgInfo.GoFiles {
		f, err := parser.ParseFile(fileSet, file, nil, 0)

		if err != nil {
			return nil, err
		}

		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)

			if ok && genDecl.Tok == token.TYPE { // type 定义
				for _, spec := range genDecl.Specs {
					// 确保是 type 定义
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					// 确保是结构体定义
					_, ok = typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					// 获取结构体名称
					structNames = append(structNames, typeSpec.Name.Name)
				}
			}
		}
	}
	//
	//fileSet.Iterate(func(file *token.File) bool {
	//	fmt.Println(file)
	//
	//	fmt.Printf("%#+v\n", file)
	//
	//	return true
	//})

	return nil, nil
}
