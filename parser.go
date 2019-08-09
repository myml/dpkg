package dpkg

import (
	"bufio"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Package control
type Package struct {
	Package       string
	Status        string
	Priority      string
	Architecture  string
	MultiArch     string
	Maintainer    string
	Version       string
	Section       string
	InstalledSize int64
	Depends       string
	PreDepends    string
	Description   string
	Source        string
	Homepage      string
	Raw           map[string]string
}

// FromMap convert map to package
func fromMap(m map[string]string) *Package {
	var pkg Package
	v := reflect.ValueOf(&pkg).Elem()
	t := reflect.TypeOf(pkg)
	for i, n := 0, t.NumField(); i < n; i++ {
		tf := t.Field(i)
		vf := v.Field(i)
		if value, ok := m[tf.Name]; ok {
			switch vf.Kind() {
			case reflect.String:
				vf.SetString(value)
			case reflect.Int64:
				i, _ := strconv.ParseInt(value, 10, 64)
				vf.SetInt(i)
			}
		}
	}
	pkg.Raw = m
	return &pkg
}

// Parse package info
func Parse(r io.Reader) ([]*Package, error) {
	scanner := bufio.NewScanner(r)
	var pkgs []*Package
	pkg := make(map[string]string)
	var buf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == ' ' {
			buf.WriteByte('\n')
			buf.WriteString(line)
			continue
		}
		if buf.Len() > 0 {
			fullline := buf.String()
			index := strings.IndexByte(fullline, ':')
			key, value := strings.TrimSpace(fullline[:index]), strings.TrimSpace(fullline[index+1:])
			key = strings.ReplaceAll(key, "-", "")
			pkg[key] = value
		}
		buf.Reset()
		buf.WriteString(line)
		if len(line) == 0 {
			pkgs = append(pkgs, fromMap(pkg))
			pkg = make(map[string]string)
		}
	}
	if len(pkg) > 0 {
		pkgs = append(pkgs, fromMap(pkg))
	}
	return pkgs, nil
}
