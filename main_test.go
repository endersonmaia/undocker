package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"
)

func TestExecute(t *testing.T) {
	var _foo = []byte("foo foo")

	tests := []struct {
		name    string
		fixture func(*testing.T, string)
		infile  string
		outfile string
		wantErr string
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
			name:    "infile does not exist",
			infile:  "t3-does-not-exist.txt",
			wantErr: "^open .*t3-does-not-exist.txt: no such file or directory$",
		},
		{
			name:    "outpath dir not writable",
			outfile: filepath.Join("t4", "does", "not", "exist"),
			wantErr: "^create: open .*/t4/does/not/exist: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var stdout bytes.Buffer
			c := &command{Stdout: &stdout}
			if tt.fixture != nil {
				tt.fixture(t, dir)
			}
			if tt.outfile != "-" {
				tt.outfile = filepath.Join(dir, tt.outfile)
			}
			inf := filepath.Join(dir, tt.infile)
			c.flattener = flattenPassthrough

			err := c.execute(inf, tt.outfile)
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
			var out []byte
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.outfile == "-" {
				out = stdout.Bytes()
			} else {
				out, err = ioutil.ReadFile(tt.outfile)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if !bytes.Equal([]byte("foo foo"), out) {
				t.Errorf("out != foo foo: %s", string(out))
			}
		})
	}
}

func flattenPassthrough(r io.ReadSeeker, w io.Writer) error {
	_, err := io.Copy(w, r)
	return err
}
