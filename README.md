# go-epub

[English](./README.md) | [简体中文](./README_zh.md)

[![version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/yourname/go-epub)

---

## Introduction
**go-epub** is a pure Go toolkit for reading EPUB files (EPUB2 and EPUB3). The project focuses on building a reusable, read-only API surface that can be used in larger systems such as bookshelf services, content analysis pipelines, or document conversion tools. The current version concentrates on reading, while the exposed abstractions allow future write support without breaking changes.

**Current Version:** `v0.1.0`

### Highlights
- 📦 Parse the EPUB container to discover OPF packages, table of contents files, and chapters.
- 🧱 A single `Book` abstraction providing access to metadata, TOC, and chapter content.
- 🔍 Convenience helpers for every Dublin Core metadata key with graceful fallbacks when data is missing.
- 📚 Support for both EPUB2 (NCX) and EPUB3 (navigation documents) TOC formats.
- 🖼️ Chapter parsing extracts plain text paragraphs and referenced image paths for downstream processing.

---

## Quick Start
```bash
# Fetch dependencies
go mod tidy

# Run the example workflow (using samples in testEpubs)
go run ./...
```

The sample in `cmd/demo/main.go` demonstrates:

- Loading an EPUB via `epub.ReadBook`.
- Accessing Dublin Core metadata (e.g., `book.Title()`, `book.Creator()`).
- Traversing the TOC using `book.FlattenTOC()` and reading chapter text with `book.ChapterByIndex`.

## API Overview

```go
import "github.com/ArcadiaLin/go-epub"

book, err := epub.ReadBook("path/to/book.epub")
if err != nil {
        // handle error
}

// Dublin Core metadata helpers
if title, err := book.Title(); err == nil {
        fmt.Println("Title:", title)
}

// Generic metadata access
values, err := book.MetadataByKey("language")

// Iterate over the full metadata map (includes <meta> extensions)
metadata := book.AllMetadata()

// Work with chapters
fmt.Println("Total chapters:", book.ChapterCount())
firstChapter, _ := book.ChapterByIndex(0)
fmt.Println(firstChapter.Text())

// Concatenate the whole book into a single text blob
fmt.Println(book.AllChaptersText())

```

### TOC and Node Utilities

- `book.FlattenTOC()` returns a linear TOC view for UI rendering.
- `TOC.FindByHref(href)` resolves a node by resource path.
- `HtmlNode` / `XmlNode` include helper methods like `Attr`, `FindAll`, and `FindNodes` for custom extensions.

## Design Notes

- `Book` acts as the unified entry point, internally managing Container, OPF, TOC, and Chapters.
- All operations are side-effect-free, consistent with a read-only design philosophy.
- The extensible API surface (e.g., `Chapter.Clone`, `Metadata.GetAll`) enables caching or write support in the future.

To integrate this library, simply import the `epub` package and call `ReadBook`. The toolkit uses robust XML/HTML parsing logic for stable behavior across various EPUB implementations.
