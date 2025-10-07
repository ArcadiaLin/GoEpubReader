# GoEpubReader

## 简介
GoEpubReader 是一个使用纯 Go 编写的 EPUB 读取工具库，兼容 EPUB2 与 EPUB3 规范。库的设计目标是为电子书阅读与内容分析提供一套稳定的、仅依赖标准库（外加 `x/net/html`）的接口。它目前聚焦于只读场景，未来也可以在此基础上扩展为写入工具。

### 核心特性
- 📦 解析 EPUB 容器，自动定位 OPF、目录与章节内容。
- 🧱 统一的 `Book` 结构体封装，提供元数据、目录、章节访问能力。
- 🔍 支持 Dublin Core 元数据的快捷访问，同时保留自定义 `<meta>` 信息。
- 📚 兼容 EPUB2 的 NCX 目录与 EPUB3 的导航文档。
- 🖼️ 章节解析同时提取文本段落与图片引用路径，适合二次加工。

## Introduction
GoEpubReader is a pure Go toolkit for reading EPUB files (EPUB2 and EPUB3). The project focuses on building a reusable, read-only API surface that can be used in larger systems such as bookshelf services, content analysis pipelines, or document conversion tools. The current version concentrates on reading, while the exposed abstractions allow future write support without breaking changes.

### Highlights
- 📦 Parse the EPUB container to discover OPF packages, table of contents files, and chapters.
- 🧱 A single `Book` abstraction providing access to metadata, TOC, and chapter content.
- 🔍 Convenience helpers for every Dublin Core metadata key with graceful fallbacks when data is missing.
- 📚 Support for both EPUB2 (NCX) and EPUB3 (navigation documents) TOC formats.
- 🖼️ Chapter parsing extracts plain text paragraphs and referenced image paths for downstream processing.

## 快速开始 / Quick Start
```bash
# 获取依赖
go mod tidy

# 运行示例工作流（使用 testEpubs 中的样例文件）
go run ./...
```

示例程序位于 `main.go`，会读取 `testEpubs/testEpub1.epub` 并展示：

- 元数据（标题、作者、语言、扩展信息等）；
- 章节数量与首章文本预览；
- 整体目录结构。

The sample workflow in `main.go` showcases how to:

- Load an EPUB via `epub.ReadBook`.
- Access Dublin Core metadata through dedicated helpers, e.g. `book.Title()` / `book.Creator()`.
- Traverse the TOC with `book.FlattenTOC()` and inspect chapter text using `book.ChapterByIndex`.

## API 概览 / API Overview
```go
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

### TOC & 节点工具
- `book.FlattenTOC()` 返回目录的线性视图，便于构建阅读器界面。
- `TOC.FindByHref(href)` 可根据资源路径反查目录节点。
- `HtmlNode` / `XmlNode` 新增的 `Attr`、`FindAll` / `FindNodes` 辅助方法方便后续拓展。

## 设计说明 / Design Notes
- `Book` 作为统一入口，内部保留了 Container、OPF、TOC、章节等结构，方便上层按需使用。
- 所有读取操作都会尽量保持无副作用，符合“只读工具库”的定位。
- 扩展接口（如 `Chapter.Clone`、`Metadata.GetAll`）便于未来在不破坏 API 的情况下增加写入或缓存功能。

如需在项目中集成，可将 `epub` 包直接引入并调用 `ReadBook`。库会在解析过程中使用通用的 XML/HTML 处理逻辑，确保在不同 EPUB 实现中都能稳定工作。
