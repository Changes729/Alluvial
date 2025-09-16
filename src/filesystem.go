package main

import (
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func md_content_process(content string) string {
	re := regexp.MustCompile(`(!)?\[(.*?)\]\(https://s2.eslite.com(.*?)\)`)
	return re.ReplaceAllString(content, "${1}[${2}](https://example.com${3})")
}

func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

type AlluvialFile struct {
	http.File

	md_buf string
}

type AlluvialFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	sys     any
}

func (fs *AlluvialFileInfo) Name() string       { return fs.name }
func (fs *AlluvialFileInfo) Size() int64        { return fs.size }
func (fs *AlluvialFileInfo) Mode() fs.FileMode  { return fs.mode }
func (fs *AlluvialFileInfo) ModTime() time.Time { return fs.modTime }
func (fs *AlluvialFileInfo) Sys() any           { return &fs.sys }
func (fs *AlluvialFileInfo) IsDir() bool        { return fs.mode.IsDir() }

func makeAlluvialFile(f http.File) (af AlluvialFile) {
	af = AlluvialFile{File: f, md_buf: ""}

	if af.IsMarkdown() {
		data, err := io.ReadAll(f)
		if err == nil {
			af.md_buf = strings.TrimSpace(md_content_process(string(data)))
			af.File.Seek(0, io.SeekStart) // Reset file offset
		}
	}

	return
}

func (f AlluvialFile) IsMarkdown() bool {
	finfo, _ := f.File.Stat()
	return strings.HasSuffix(finfo.Name(), ".md")
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f AlluvialFile) Readdir(n int) (fis []fs.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files { // Filters out the dot files
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}
	if err == nil && n > 0 && len(fis) == 0 {
		err = io.EOF
	}
	return
}

func (f AlluvialFile) Read(p []byte) (n int, err error) {
	offset, _ := f.File.Seek(0, io.SeekCurrent)
	content_size := int64(len(f.md_buf))
	end := offset + int64(len(p))
	err = nil

	if offset >= content_size {
		return 0, io.EOF
	}

	if f.IsMarkdown() {
		if end > content_size {
			end = content_size
		}

		n = copy(p, f.md_buf[offset:end])
		f.File.Seek(int64(n), io.SeekCurrent)
		return
	}
	return f.File.Read(p)
}

func (f AlluvialFile) Stat() (fs.FileInfo, error) {
	fi, e := f.File.Stat()

	var fs AlluvialFileInfo
	fs.name = fi.Name()
	fs.size = fi.Size()
	fs.mode = fi.Mode()
	fs.modTime = fi.ModTime()
	fs.sys = fi.Sys()

	if f.IsMarkdown() {
		fs.size = int64(len(f.md_buf))
	}

	return &fs, e
}

// AlluvialFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type AlluvialFileSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fsys AlluvialFileSystem) Open(name string) (http.File, error) {
	if containsDotFile(name) { // If dot file, return 403 response
		return nil, fs.ErrPermission
	}

	file, err := fsys.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return makeAlluvialFile(file), nil
}
