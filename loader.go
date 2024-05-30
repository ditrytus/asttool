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
	absDir, err := filepath.Abs(d.dir)
	if err != nil {
		return nil, nil, err
	}
	var pkgs []*packages.Package
	fileSet := token.NewFileSet()
	err = filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
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
