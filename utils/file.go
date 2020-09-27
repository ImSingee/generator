package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"go/token"
	"golang.org/x/tools/imports"
	"io"
	"os"
)

func SaveToFile(filename string, content []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)

	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", filename, err)
	}
	defer f.Close()

	writer := io.Writer(f)

	_, err = writer.Write(content)

	if viper.GetBool("debug") {
		fmt.Printf("Save %s:\n%s", filename, content)
	}

	return err
}

func SaveGoCodeToFile(filename string, content []byte) error {
	// imports.Process 包含了 format.Source 所做的内容
	content, err := imports.Process(filename, content, nil)

	if err != nil {
		return fmt.Errorf("cannot format generated code: %w", err)
	}

	return SaveToFile(filename, content)
}

func GetStringFromPosition(start token.Pos, end token.Pos) (string, error) {
	startPos := fileSet.Position(start)
	endPos := fileSet.Position(end)

	f, err := os.Open(startPos.Filename)

	if err != nil {
		return "", fmt.Errorf("cannot read file %s: %w", startPos.Filename, err)
	}

	defer f.Close()

	text := make([]byte, endPos.Offset-startPos.Offset)

	_, _ = f.Seek(0, 0)
	_, err = f.ReadAt(text, int64(startPos.Offset))

	if err != nil {
		return "", fmt.Errorf("cannot read file %s in [%d, %d): %w", startPos.Filename, startPos.Offset, endPos.Offset, err)
	}

	return ToString(text), nil
}
