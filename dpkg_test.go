package dpkg

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// Please download deb package to tesetdata dir

const ListMaxSize = 100

var list []string

func TestGlobTestData(t *testing.T) {
	var err error
	list, err = filepath.Glob("testdata/*.deb")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal("not found testdata")
	}
	if len(list) > ListMaxSize {
		rand.Seed(time.Now().UnixNano())
		sort.Slice(list, func(i, j int) bool {
			return rand.Intn(2) == 1
		})
		list = list[:ListMaxSize]
	}
}
func TestNewReader(t *testing.T) {
	for i := range list {
		err := newReader(list[i])
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadData(t *testing.T) {
	var fileSize int64
	var dataSize int64
	for i := range list {
		func() {
			f, err := os.Open(list[i])
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			info, err := f.Stat()
			if err != nil {
				t.Fatal(err)
			}
			fileSize += info.Size()
			control, data, err := NewReader(f)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range []*TarReader{control, data} {
				for {
					_, err := r.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					n, err := io.Copy(ioutil.Discard, r)
					if err != nil {
						t.Fatal(err)
					}
					dataSize += n
				}
			}
		}()
	}
	t.Logf("test file size %vM", fileSize/1024/1024)
	t.Logf("test data size %vM", dataSize/1024/1024)
}

func TestParsePackage(t *testing.T) {
	for i := range list {
		func() {
			f, err := os.Open(list[i])
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			control, _, err := NewReader(f)
			if err != nil {
				t.Fatal(err)
			}
			for {
				h, err := control.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if !h.FileInfo().IsDir() && filepath.Base(h.Name) == "control" {
					pkgs, err := Parse(control)
					if err != nil {
						t.Fatal(err)
					}
					if len(pkgs) != 1 {
						t.Error("parse invalid", h.Name, pkgs)
					}
				}
			}
		}()
	}
}

func newReader(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	_, _, err = NewReader(f)
	return err
}
