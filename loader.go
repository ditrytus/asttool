package asttool

import (
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
)

type PackageLoader interface {
	Load() ([]*packages.Package, *token.FileSet, error)
}

type dirPackageLoader struct {
	dir string
}

func NewDirPackageLoader(dir string) PackageLoader {
	return &dirPackageLoader{dir: dir}
}

func (d *dirPackageLoader) Load() ([]*packages.Package, *token.FileSet, error) {
	var pkgs []*packages.Package
	fileSet := token.NewFileSet()
	err := filepath.Walk(d.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		conf := &packages.Config{Mode: packages.LoadSyntax, Fset: fileSet, Dir: path}
		dirPkgs, err := packages.Load(conf, path)
		if err != nil {
			return err
		}
		for _, pkg := range dirPkgs {
			if len(pkg.Errors) == 1 && pkg.Errors[0].Kind == packages.ListError {
				continue
			}
			pkgs = append(pkgs, pkg)
		}
		return nil
	})
	return pkgs, fileSet, err
}
