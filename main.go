package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"epub2/epub"
)

// This executable showcases a simple read-only workflow using the library.
func main() {
	example := filepath.Join("testEpubs", "testEpub1.epub")

	book, err := epub.ReadBook(example)
	if err != nil {
		log.Fatalf("read book: %v", err)
	}

	fmt.Printf("Loaded %s\n", example)

	if title, err := book.Title(); err == nil {
		fmt.Printf("Title: %s\n", title)
	} else {
		fmt.Printf("Title: %v\n", err)
	}

	if author, err := book.Creator(); err == nil {
		fmt.Printf("Author: %s\n", author)
	}

	metadata := book.AllMetadata()
	fmt.Printf("Language(s): %s\n", strings.Join(metadata["language"], ", "))
	fmt.Printf("Total chapters: %d\n", book.ChapterCount())

	if first, err := book.ChapterByIndex(0); err == nil {
		fmt.Printf("First chapter title: %s\n", first.Title)
		fmt.Printf("Preview: %s\n", preview(first.Text()))
	}

	fmt.Println("Table of contents:")
	entries := book.FlattenTOC()
	for _, entry := range entries {
		if strings.TrimSpace(entry.Title) == "" {
			continue
		}
		fmt.Printf("  - %s (%s)\n", entry.Title, entry.Href)
	}

	fmt.Println("Highlighted metadata:")
	keys := make([]string, 0, len(metadata))
	for k := range metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if len(metadata[k]) == 0 {
			continue
		}
		fmt.Printf("  %s: %s\n", k, strings.Join(metadata[k], "; "))
	}
}

func preview(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	fields := strings.Fields(text)
	text = strings.Join(fields, " ")
	const limit = 160
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return text
}
