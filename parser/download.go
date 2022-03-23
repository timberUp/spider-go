package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	HTTPSPrefix = "https://"
)

func Download(url, outDir string, body io.ReadCloser) error {
	fileName := filepath.Join(outDir, format(url))
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file, err: %v", err)
		}
		defer f.Close()
		_, err = io.Copy(f, body)
		if err != nil {
			return fmt.Errorf("failed to write html into file, err: %v", err)
		}
	} else {
		return fmt.Errorf("[%s] has been downloaded", fileName)
	}
	return nil
}

func format(s string) string {
	if strings.HasPrefix(s, HTTPSPrefix) {
		s = s[8:]
	} else {
		s = s[7:]
	}
	s = strings.Replace(s, "_", "-", -1)
	s = strings.Replace(s, "/", "_", -1)
	return s
}
