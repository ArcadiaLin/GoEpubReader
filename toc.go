package epub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"path"
	"strings"
)

type TOC struct {
	Title    string
	Href     string
	Children []TOC
}

// Walk traverses the table-of-contents tree in depth-first order and invokes fn
// for every entry. Returning a non-nil error from the callback aborts the walk
// and propagates the error to the caller.
func (t TOC) Walk(fn func(TOC) error) error {
	if fn == nil {
		return nil
	}
	if err := fn(t); err != nil {
		return err
	}
	for _, child := range t.Children {
		if err := child.Walk(fn); err != nil {
			return err
		}
	}
	return nil
}

// FindByHref returns the first TOC entry whose href matches the provided value.
// The comparison is performed after cleaning the href to make it resilient to
// small differences in path formatting.
func (t TOC) FindByHref(href string) *TOC {
	cleaned := path.Clean(href)
	var found *TOC
	_ = t.Walk(func(entry TOC) error {
		if found != nil {
			return nil
		}
		if entry.Href != "" && path.Clean(entry.Href) == cleaned {
			copy := entry
			found = &copy
		}
		return nil
	})
	return found
}

// Flatten flattens the table of contents into a slice preserving the natural
// reading order.
func (t TOC) Flatten() []TOC {
	var entries []TOC
	_ = t.Walk(func(entry TOC) error {
		entries = append(entries, entry)
		return nil
	})
	return entries
}

func ParseTOC(tocType, tocFile, opfDir string, files map[string]*zip.File) ([]TOC, error) {
	f, ok := files[path.Join(opfDir, tocFile)]
	if !ok {
		return nil, fmt.Errorf("toc file not found: %s", tocFile)
	}

	content, err := getContent(f)
	if err != nil {
		return nil, err
	}

	switch tocType {
	case TOCTypeEPUB3:
		return parseNavXML(content, opfDir)
	case TOCTypeEPUB2:
		return parseNCX(content, opfDir)
	default:
		return nil, fmt.Errorf("unknown toc type")
	}
}

// parseNavXML 用 XmlNode 解析 EPUB3 nav.xhtml
func parseNavXML(content []byte, basePath string) ([]TOC, error) {
	root, err := ParseXML(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parseNavXML: invalid XHTML: %w", err)
	}

	// 找到 <nav epub:type="toc">
	var tocNav *XmlNode
	var findNav func(*XmlNode)
	findNav = func(n *XmlNode) {
		if n.XMLName.Local == "nav" {
			for _, a := range n.Attrs {
				if (a.Name.Local == "type" && strings.Contains(a.Value, "toc")) ||
					(a.Name.Local == "epub:type" && strings.Contains(a.Value, "toc")) {
					tocNav = n
					return
				}
			}
		}
		for i := range n.XmlNodes {
			findNav(&n.XmlNodes[i])
			if tocNav != nil {
				return
			}
		}
	}
	findNav(root)
	if tocNav == nil {
		return nil, fmt.Errorf("parseNavXML: toc <nav> not found")
	}

	// 找到 <ol>
	ol := tocNav.FindNode("ol")
	if ol == nil {
		return nil, fmt.Errorf("parseNavXML: <ol> not found in toc nav")
	}

	// 递归解析 <li>
	var parseList func(*XmlNode) []TOC
	parseList = func(node *XmlNode) []TOC {
		var entries []TOC
		for _, li := range node.XmlNodes {
			if li.XMLName.Local != "li" {
				continue
			}
			var entry TOC
			for _, child := range li.XmlNodes {
				switch child.XMLName.Local {
				case "a":
					entry.Title = strings.TrimSpace(child.NodeText())
					for _, a := range child.Attrs {
						if a.Name.Local == "href" {
							entry.Href = resolveRelative(basePath, a.Value)
							break
						}
					}
				case "ol":
					entry.Children = parseList(&child)
				}
			}
			if entry.Title != "" {
				entries = append(entries, entry)
			}
		}
		return entries
	}

	return parseList(ol), nil
}

// parseNCX 解析 EPUB2 toc.ncx
func parseNCX(content []byte, basePath string) ([]TOC, error) {
	root, err := ParseXML(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parseNCX: invalid XML: %w", err)
	}

	navMap := root.FindNode("navMap")
	if navMap == nil {
		return nil, fmt.Errorf("parseNCX: no navMap found")
	}

	var parseNavPoints func([]XmlNode) []TOC
	parseNavPoints = func(nodes []XmlNode) []TOC {
		var entries []TOC
		for _, n := range nodes {
			if n.XMLName.Local != "navPoint" {
				continue
			}

			var label, href string
			for _, c := range n.XmlNodes {
				switch c.XMLName.Local {
				case "navLabel":
					label = strings.TrimSpace(c.NodeText())
				case "content":
					for _, a := range c.Attrs {
						if a.Name.Local == "src" {
							href = resolveRelative(basePath, strings.TrimSpace(a.Value))
						}
					}
				}
			}

			children := parseNavPoints(n.XmlNodes)
			if label != "" {
				entries = append(entries, TOC{
					Title:    label,
					Href:     path.Clean(href),
					Children: children,
				})
			}
		}
		return entries
	}

	return parseNavPoints(navMap.XmlNodes), nil
}
