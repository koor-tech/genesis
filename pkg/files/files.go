package files

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(src, dst string) error {
	sources, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create dir %s, error %q", dst, err)
	}

	for _, source := range sources {
		srcPath := filepath.Join(src, source.Name())
		dstPath := filepath.Join(dst, source.Name())

		if source.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SaveInFile(name, content string, permissions int) error {
	file, err := os.OpenFile(name, os.O_CREATE, os.FileMode(permissions))
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("unable to close file")
		}
	}(file)

	if err != nil {
		log.Fatalf("Failed to open public.key for writing: %s", err)
	}
	err = os.WriteFile(name, []byte(content), os.FileMode(permissions))
	if err != nil {
		return err
	}
	return nil
}
