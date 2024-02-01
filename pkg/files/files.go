package files

import (
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

	os.MkdirAll(dst, os.ModePerm)

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
	defer file.Close()

	if err != nil {
		log.Fatalf("Failed to open public.key for writing: %s", err)
	}
	err = os.WriteFile(name, []byte(content), os.FileMode(permissions))
	if err != nil {
		return err
	}
	return nil
}
