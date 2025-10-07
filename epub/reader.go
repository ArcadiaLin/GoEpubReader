package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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

// TODO provide a function to new a empty Book type

// TODO provide a ReadBook function to parse a Book from a epub path, which will fill the empty Book type, this function can use other Parse function and all types' Method

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
