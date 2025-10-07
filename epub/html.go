package epub

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
)

// TODO complement methods for these Node types if needed

// HtmlNode 是 HTML 版本的通用节点结构
type HtmlNode struct {
	Type     NodeType
	Name     string
	Attrs    map[string]string
	Content  string
	Children []*HtmlNode
}

// ParseHTML 解析 HTML/XHTML，返回 HtmlNode 树
func ParseHTML(r io.Reader) (*HtmlNode, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return convertHTMLNode(root), nil
}

// convertHTMLNode 将 html.Node 转为 HtmlNode
func convertHTMLNode(n *html.Node) *HtmlNode {
	switch n.Type {
	case html.ElementNode:
		node := &HtmlNode{
			Type:     ElementNode,
			Name:     n.Data,
			Attrs:    map[string]string{},
			Children: []*HtmlNode{},
		}
		for _, attr := range n.Attr {
			node.Attrs[attr.Key] = attr.Val
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			child := convertHTMLNode(c)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		}
		return node

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text == "" {
			return nil
		}
		return &HtmlNode{Type: TextNode, Content: text}
	}
	return nil
}

// NodeText 递归提取所有文本（去除标签与多余空白）
func (hn *HtmlNode) NodeText() string {
	if hn == nil {
		return ""
	}
	var sb strings.Builder
	stack := []*HtmlNode{hn}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node.Type == TextNode {
			sb.WriteString(node.Content)
			sb.WriteString(" ")
		}
		for i := len(node.Children) - 1; i >= 0; i-- {
			stack = append(stack, node.Children[i])
		}
	}
	return strings.TrimSpace(sb.String())
}

func (hn *HtmlNode) FindNode(name string) *HtmlNode {
	if hn == nil {
		return nil
	}
	if hn.Type == ElementNode && hn.Name == name {
		return hn
	}
	for _, c := range hn.Children {
		if r := findElement(c, name); r != nil {
			return r
		}
	}
	return nil
}
