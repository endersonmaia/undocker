package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"git.sr.ht/~motiejus/undocker/rootfs"
)

func TestExecute(t *testing.T) {
	var _foo = []byte("foo foo")

	tests := []struct {
		name      string
		fixture   func(*testing.T, string)
		flattener func(io.ReadSeeker, io.Writer, ...rootfs.Option) error
		infile    string
		outfile   string
		wantErr   string
		assertion func(*testing.T, string)
	}{
		{
			name:   "ok passthrough via stdout",
			infile: "t10-in.txt",
			fixture: func(t *testing.T, dir string) {
				fname := filepath.Join(dir, "t10-in.txt")
				if err := ioutil.WriteFile(fname, _foo, 0644); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			outfile: "-",
		},
		{
			name:   "ok passthrough via file",
			infile: "t20-in.txt",
			fixture: func(t *testing.T, dir string) {
				fname := filepath.Join(dir, "t20-in.txt")
				if err := ioutil.WriteFile(fname, _foo, 0644); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			outfile: "t20-out.txt",
		},
		{
			name:   "bad flattener should remove the file",
			infile: "t30-in.txt",
			fixture: func(t *testing.T, dir string) {
				fname := filepath.Join(dir, "t30-in.txt")
				if err := ioutil.WriteFile(fname, _foo, 0644); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
			flattener: flattenBad,
			outfile:   "t30-out.txt",
			wantErr:   "some error",
			assertion: func(t *testing.T, dir string) {
				d, err := os.ReadDir(dir)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(d) != 1 {
					t.Fatalf("expected 1 entry, got %d", len(d))
				}
				if d[0].Name() != "t30-in.txt" {
					t.Fatalf("expected to find only t30-in.txt, got %s", d[0].Name())
				}
			},
		},
		{
			name:    "infile does not exist",
			infile:  "t3-does-not-exist.txt",
			wantErr: "^open .*not-exist.txt: no such file or directory$",
		},
		{
			name:    "outpath dir not writable",
			outfile: filepath.Join("t4", "does", "not", "exist"),
			wantErr: "^create: open .*not/exist: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var stdout bytes.Buffer
			if tt.flattener == nil {
				tt.flattener = flattenPassthrough
			}

			if tt.fixture != nil {
				tt.fixture(t, dir)
			}
			if tt.outfile != "-" {
				tt.outfile = filepath.Join(dir, tt.outfile)
			}
			inf := filepath.Join(dir, tt.infile)

			c := &command{Stdout: &stdout, flattener: tt.flattener}
			err := c.execute(inf, tt.outfile, "")

			if tt.assertion != nil {
				tt.assertion(t, dir)
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				r := regexp.MustCompile(tt.wantErr)
				if r.FindStringIndex(err.Error()) == nil {
					t.Errorf("%s not found in %s", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var out []byte
			if tt.outfile == "-" {
				out = stdout.Bytes()
			} else {
				out, err = ioutil.ReadFile(tt.outfile)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if !bytes.Equal([]byte("foo foo"), out) {
				t.Errorf("out != foo foo: %q", string(out))
			}

		})
	}
}

func flattenPassthrough(r io.ReadSeeker, w io.Writer, _ ...rootfs.Option) error {
	_, err := io.Copy(w, r)
	return err
}

func flattenBad(_ io.ReadSeeker, _ io.Writer, _ ...rootfs.Option) error {
	return errors.New("some error")
}
