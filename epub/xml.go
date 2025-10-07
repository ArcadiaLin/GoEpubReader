package epub

import (
	"encoding/xml"
	"io"
	"strings"
)

type XmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	XmlNodes []XmlNode  `xml:",any"`
}

type EmptyXmlNode struct {
	Name  string
	Attrs map[string]string
}

// TODO complement methods for these Node types if needed

// ParseXML 解析 XML，返回 XmlNode 树 / ParseXML decodes the XML stream into a XmlNode tree.
func ParseXML(r io.Reader) (*XmlNode, error) {
	decoder := xml.NewDecoder(r)
	var root XmlNode
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	return &root, nil
}

func (xn *XmlNode) NodeText() string {
	if xn == nil {
		return ""
	}
	var parts []string
	if strings.TrimSpace(xn.Content) != "" {
		parts = append(parts, strings.TrimSpace(xn.Content))
	}
	for i := range xn.XmlNodes {
		if text := xn.XmlNodes[i].NodeText(); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func (xn *XmlNode) FindNode(name string) *XmlNode {
	if xn == nil {
		return nil
	}
	if strings.EqualFold(xn.XMLName.Local, name) {
		return xn
	}
	for i := range xn.XmlNodes {
		if res := xn.XmlNodes[i].FindNode(name); res != nil {
			return res
		}
	}
	return nil
}
