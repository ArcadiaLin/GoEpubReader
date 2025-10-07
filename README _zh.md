# go-epub

[English](./README.md) | [简体中文](./README_zh.md)

[![版本](https://img.shields.io/badge/版本-0.1.0-blue.svg)](https://github.com/yourname/go-epub)

---

## 简介
**go-epub** 是一个使用纯 Go 编写的 EPUB 阅读工具库，兼容 EPUB2 与 EPUB3 规范。其目标是提供稳定、可复用的只读 API，方便集成到电子书阅读器、内容分析管线或文档转换系统中。当前版本聚焦于阅读功能，API 设计支持未来的写入扩展而无需破坏兼容性。

**当前版本：** `v0.1.0`

### 核心特性
- 📦 解析 EPUB 容器，自动定位 OPF、目录与章节内容。
- 🧱 统一的 `Book` 结构体封装，提供元数据、目录与章节访问。
- 🔍 提供 Dublin Core 元数据的便捷访问方法，并保留自定义 `<meta>` 信息。
- 📚 同时兼容 EPUB2 的 NCX 目录与 EPUB3 的导航文档格式。
- 🖼️ 章节解析可提取纯文本段落与图片引用路径，适合二次分析处理。

---

## 快速开始
```bash
# 获取依赖
go mod tidy

# 运行示例（使用 testEpubs 下的样例文件）
go run ./...
```

示例程序位于 `cmd/demo/main.go`，会读取 `testEpubs/testEpub1.epub` 并展示：

- 元数据（标题、作者、语言、扩展信息等）；
- 章节数量与首章内容预览；
- 整体目录结构。

## API 概览

```go
import "github.com/ArcadiaLin/go-epub"

book, err := epub.ReadBook("path/to/book.epub")
if err != nil {
        // 处理错误
}

// Dublin Core 元数据访问
if title, err := book.Title(); err == nil {
        fmt.Println("Title:", title)
}

// 通用元数据接口
values, err := book.MetadataByKey("language")

// 遍历完整元数据（包含 <meta> 扩展）
metadata := book.AllMetadata()

// 章节操作
fmt.Println("章节总数:", book.ChapterCount())
firstChapter, _ := book.ChapterByIndex(0)
fmt.Println(firstChapter.Text())

// 拼接整本书文本
fmt.Println(book.AllChaptersText())

```

### 目录与节点工具

- `book.FlattenTOC()` 返回线性目录视图，方便构建阅读器界面。
- `TOC.FindByHref(href)` 可根据资源路径查找对应节点。
- `HtmlNode` / `XmlNode` 提供 `Attr`、`FindAll`、`FindNodes` 等辅助方法。

## 设计说明

- `Book` 是统一入口，内部封装 Container、OPF、TOC 与章节结构。
- 所有读取操作保持无副作用，符合“只读工具库”的定位。
- 扩展接口（如 `Chapter.Clone`、`Metadata.GetAll`）便于未来增加缓存或写入功能。

如需集成，可直接引入 `epub` 包并调用 `ReadBook`。解析逻辑基于通用 XML/HTML 处理，保证在不同 EPUB 实现中稳定运行。
