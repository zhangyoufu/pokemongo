package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// comment
func main() {
	fmt.Printf("Welcome to the pokemon world. Please tell me your name first ^_^\n")
	fmt.Printf("Size: ")
	var size int
	fmt.Scanf("%d", &size)
	if size == 0 {
		return
	}
	input_base64 := ""
	fmt.Printf("\n\nContent: ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input_base64 += scanner.Text()
		if len(input_base64) >= size {
			break
		}
	}
	// comment
	input_bytes, err := base64.StdEncoding.DecodeString(input_base64)
	if err != nil {
		fmt.Printf("Decode error: %s\n", input_base64)
		return
	}
	output, _ := sanitizeAndRun(string(input_bytes))
	fmt.Printf("Let's go: %s\n", output)
}

func sanitizeAndRun(src string) (string, error) {
	sanitized_src, err := sanitize(src)
	if err != nil {
		return "", err
	}
	return run(sanitized_src)
}

func sanitize(src string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		return "", err
	}
	for _, imp := range f.Imports {
		return "", fmt.Errorf("import %v not allowed", imp.Path.Value)
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func run(src string) (string, error) {
	dir, err := ioutil.TempDir("", "pokemongo")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	src_path := filepath.Join(dir, "main.go")
	if err = ioutil.WriteFile(src_path, []byte(src), 0600); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-buildmode=pie", src_path)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout: %v", err)
		}
		return "", err
	}
	exec_path := filepath.Join(dir, "main")
	output, err := exec.CommandContext(ctx, exec_path).CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout: %v", err)
		}
		return "", err
	}
	return string(output), nil
}
