package epub

import (
	"archive/zip"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type Chapter struct {
	ID         string
	Path       string
	Title      string
	Paragraphs []string
	Images     []string
}

// Text joins all extracted paragraphs into a single string separated by blank
// lines. The returned value is suitable for plain-text readers.
func (c *Chapter) Text() string {
	if c == nil || len(c.Paragraphs) == 0 {
		return ""
	}
	return strings.Join(c.Paragraphs, "\n\n")
}

// HasImages reports whether the chapter contains any referenced images.
func (c *Chapter) HasImages() bool {
	return c != nil && len(c.Images) > 0
}

// Clone returns a deep copy of the chapter. This is handy when callers want to
// modify the returned value without affecting the book cache.
func (c *Chapter) Clone() Chapter {
	if c == nil {
		return Chapter{}
	}
	clone := Chapter{
		ID:    c.ID,
		Path:  c.Path,
		Title: c.Title,
	}
	clone.Paragraphs = append(clone.Paragraphs, c.Paragraphs...)
	clone.Images = append(clone.Images, c.Images...)
	return clone
}

func ParseChapter(id, href string, f *zip.File) (*Chapter, error) {
	if f == nil {
		return nil, fmt.Errorf("nil chapter file reference")
	}

	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rc.Close())
	}()

	root, err := ParseHTML(rc)
	if err != nil {
		return nil, err
	}

	title := findFirstText(root, "title")
	body := root.FindNode("body")
	if body == nil {
		body = root
	}

	paragraphs := extractParagraphs(body)
	images := extractImages(body, filepath.Dir(href))

	return &Chapter{
		ID:         id,
		Path:       href,
		Title:      title,
		Paragraphs: paragraphs,
		Images:     images,
	}, nil
}

// 辅助函数
func findElement(n *HtmlNode, name string) *HtmlNode {
	if n == nil {
		return nil
	}
	if n.Type == ElementNode && n.Name == name {
		return n
	}
	for _, c := range n.Children {
		if r := findElement(c, name); r != nil {
			return r
		}
	}
	return nil
}

func findFirstText(n *HtmlNode, name string) string {
	node := findElement(n, name)
	if node == nil {
		return ""
	}
	return node.NodeText()
}

func extractParagraphs(body *HtmlNode) []string {
	if body == nil {
		return nil
	}
	var res []string
	for _, c := range body.Children {
		if c.Type == ElementNode && (c.Name == "p" || c.Name == "div") {
			txt := c.NodeText()
			if txt != "" {
				res = append(res, txt)
			}
		} else {
			res = append(res, extractParagraphs(c)...)
		}
	}
	return res
}

func extractImages(body *HtmlNode, base string) []string {
	if body == nil {
		return nil
	}
	var res []string
	if body.Type == ElementNode && body.Name == "img" {
		if src, ok := body.Attrs["src"]; ok {
			res = append(res, filepath.Join(base, src))
		}
	}
	for _, c := range body.Children {
		res = append(res, extractImages(c, base)...)
	}
	return res
}
