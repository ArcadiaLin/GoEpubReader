package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

var dublinCoreElements = []string{
	"title",
	"creator",
	"subject",
	"description",
	"publisher",
	"contributor",
	"date",
	"type",
	"format",
	"identifier",
	"language",
	"source",
	"relation",
	"coverage",
	"rights",
}

func init() {
	sort.Strings(dublinCoreElements)
}

// NewBook creates an empty book structure that can later be populated by the
// parser. The slices are initialised to avoid nil handling by callers.
func NewBook() *Book {
	return &Book{
		Chapters: make([]Chapter, 0),
	}
}

// ReadBook parses the EPUB file located at epubPath and populates a Book
// structure with metadata, table of contents and chapter information.
func ReadBook(epubPath string) (*Book, error) {
	zr, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, zr.Close())
	}()

	book := NewBook()

	files := collectFiles(zr.File)

	containerFile, ok := files["META-INF/container.xml"]
	if !ok {
		return nil, fmt.Errorf("container.xml not found")
	}
	containerData, err := getContent(containerFile)
	if err != nil {
		return nil, fmt.Errorf("read container: %w", err)
	}
	container, err := ParseContainer(containerData)
	if err != nil {
		return nil, err
	}
	book.Container = container

	opfPath, err := container.FindOpfFile()
	if err != nil {
		return nil, err
	}
	opfPath = path.Clean(opfPath)
	opfFile, ok := files[opfPath]
	if !ok {
		return nil, fmt.Errorf("opf file not found: %s", opfPath)
	}
	opfContent, err := getContent(opfFile)
	if err != nil {
		return nil, fmt.Errorf("read opf: %w", err)
	}
	opf, err := ParseOpf(opfContent)
	if err != nil {
		return nil, err
	}
	book.Opf = opf

	opfDir := path.Dir(opfPath)
	if opfDir == "." {
		opfDir = ""
	}

	tocType, tocFile := opf.FindTOCFile(opfPath)
	if tocType != TOCTypeUnknown && tocFile != "" {
		tocRef := tocFile
		if opfDir != "" && strings.HasPrefix(tocRef, opfDir+"/") {
			tocRef = strings.TrimPrefix(tocRef, opfDir+"/")
		}
		entries, parseErr := ParseTOC(tocType, tocRef, opfDir, files)
		if parseErr != nil {
			return nil, fmt.Errorf("parse toc: %w", parseErr)
		}
		book.TOC = &TOC{Children: entries}
	}

	chapterIDs := opf.Spine.ExtractChapterIDs()
	hrefLookup := opf.Manifest.HrefLookup(opfPath)
	for _, id := range chapterIDs {
		href, ok := hrefLookup[id]
		if !ok {
			continue
		}
		href = path.Clean(href)
		chapFile, ok := files[href]
		if !ok {
			continue
		}
		chapter, parseErr := ParseChapter(id, href, chapFile)
		if parseErr != nil {
			return nil, fmt.Errorf("parse chapter %s: %w", id, parseErr)
		}
		book.Chapters = append(book.Chapters, *chapter)
	}

	return book, nil
}

// getContent 从 zip.File 读取全部内容 / getContent reads the full content from the zip file entry.
func getContent(f *zip.File) ([]byte, error) {
	fileReader, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, fileReader.Close())
	}()
	return io.ReadAll(fileReader)
}

func collectFiles(entries []*zip.File) map[string]*zip.File {
	files := make(map[string]*zip.File, len(entries))
	for _, f := range entries {
		files[f.Name] = f
	}
	return files
}

func xmlNewDecoder(r io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(r)
	decoder.Strict = false
	decoder.CharsetReader = charsetReader
	return decoder
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "", "utf-8", "us-ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
}

func xmlUnmarshal(data []byte, v any) error {
	decoder := xmlNewDecoder(bytes.NewReader(data))
	return decoder.Decode(v)
}
