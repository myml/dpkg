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
	Package       string `json:"package"`
	Status        string `json:"status"`
	Priority      string `json:"priority"`
	Architecture  string `json:"architecture"`
	MultiArch     string `json:"multi_arch"`
	Maintainer    string `json:"maintainer"`
	Version       string `json:"version"`
	Section       string `json:"section"`
	InstalledSize int64  `json:"installed_size"`
	Depends       string `json:"depends"`
	PreDepends    string `json:"pre_depends"`
	Description   string `json:"description"`
	Source        string `json:"source"`
	Homepage      string `json:"homepage"`

	Raw map[string]string `json:"raw"`
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

func ParseLine(fullline string) (string, string) {
	index := strings.IndexByte(fullline, ':')
	key, value := strings.TrimSpace(fullline[:index]), strings.TrimSpace(fullline[index+1:])
	key = strings.ReplaceAll(key, "-", "")
	return key, value
}

// Parse package info
func Parse(r io.Reader) ([]*Package, error) {
	scanner := bufio.NewScanner(r)
	var pkgs []*Package
	pkg := make(map[string]string)
	var buf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			buf.WriteByte('\n')
			buf.WriteString(line)
			continue
		}
		if buf.Len() > 0 {
			key, value := ParseLine(buf.String())
			pkg[key] = value
		}
		buf.Reset()
		buf.WriteString(line)
		if len(line) == 0 {
			pkgs = append(pkgs, fromMap(pkg))
			pkg = make(map[string]string)
		}
	}
	key, value := ParseLine(buf.String())
	pkg[key] = value
	if len(pkg) > 0 {
		pkgs = append(pkgs, fromMap(pkg))
	}
	return pkgs, nil
}
