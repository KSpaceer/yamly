package parser

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// inspired by easyjson (https://github.com/mailru/easyjson/blob/master/parser/pkgpath.go)
func findPkgPath(dirPath string) (string, error) {
	if !filepath.IsAbs(dirPath) {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dirPath = filepath.Join(currentDir, dirPath)
	}

	goModPath, _ := getGoModPath(dirPath)

	if isGoModFile(goModPath) {
		return pkgPathFromGoMod(dirPath, goModPath)
	}
	return pkgPathFromGOPATH(dirPath)
}

func pkgPathFromGOPATH(dirPath string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	paths := strings.Split(gopath, string(filepath.ListSeparator))
	for _, p := range paths {
		prefix := filepath.Join(p, "src") + string(filepath.Separator)
		rel, err := filepath.Rel(prefix, dirPath)
		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return path.Dir(filePathToPackagePath(rel)), nil
		}
	}
	return "", fmt.Errorf("directory '%v' is not in GOPATH '%v'", dirPath, gopath)
}

var pkgPathFromGoModCache = struct {
	paths map[string]string
	sync.RWMutex
}{
	paths: make(map[string]string),
}

func pkgPathFromGoMod(dirPath, goModPath string) (string, error) {
	modulePath := getModulePath(goModPath)
	if modulePath == "" {
		return "", fmt.Errorf("cannot determine module path from %s", goModPath)
	}

	rel := path.Join(modulePath, filePathToPackagePath(strings.TrimPrefix(dirPath, filepath.Dir(goModPath))))

	return path.Clean(rel), nil

}

func getModulePath(goModPath string) string {
	pkgPathFromGoModCache.RLock()
	modPath, ok := pkgPathFromGoModCache.paths[goModPath]
	pkgPathFromGoModCache.RUnlock()
	if ok {
		return modPath
	}

	defer func() {
		pkgPathFromGoModCache.Lock()
		pkgPathFromGoModCache.paths[goModPath] = modPath
		pkgPathFromGoModCache.Unlock()
	}()

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	modPath = modulePath(data)
	return modPath
}

func getGoModPath(dirPath string) (string, error) {
	pkgPathFromGoModCache.RLock()
	goModPath, ok := pkgPathFromGoModCache.paths[dirPath]
	pkgPathFromGoModCache.RUnlock()
	if ok {
		return goModPath, nil
	}

	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Dir = dirPath

	defer func() {
		pkgPathFromGoModCache.Lock()
		pkgPathFromGoModCache.paths[dirPath] = goModPath
		pkgPathFromGoModCache.Unlock()
	}()

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	goModPath = strings.TrimSpace(string(stdout))

	return goModPath, nil
}

func isGoModFile(path string) bool {
	return strings.Contains(path, "go.mod")
}

func filePathToPackagePath(path string) string {
	return filepath.ToSlash(path)
}
