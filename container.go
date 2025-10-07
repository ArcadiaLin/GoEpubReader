package epub

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Container struct {
	Rootfiles []EmptyXmlNode
}

// ParseContainer parses the META-INF/container.xml document into a Container
// structure.
func ParseContainer(content []byte) (*Container, error) {
	root, err := ParseXML(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse container: %w", err)
	}

	rootfiles := root.FindNode("rootfiles")
	if rootfiles == nil {
		return nil, fmt.Errorf("container rootfiles not found")
	}

	container := &Container{}
	for i := range rootfiles.XmlNodes {
		child := &rootfiles.XmlNodes[i]
		if !strings.EqualFold(child.XMLName.Local, "rootfile") {
			continue
		}
		attrs := make(map[string]string)
		for _, attr := range child.Attrs {
			attrs[attr.Name.Local] = attr.Value
		}
		container.Rootfiles = append(container.Rootfiles, EmptyXmlNode{
			Name:  child.XMLName.Local,
			Attrs: attrs,
		})
	}

	if len(container.Rootfiles) == 0 {
		return nil, fmt.Errorf("no rootfile entries in container")
	}

	return container, nil
}

func (c *Container) FindOpfFile() (string, error) {
	for _, rf := range c.Rootfiles {
		fullPath := strings.TrimSpace(rf.Attrs["full-path"])
		if fullPath == "" {
			continue
		}
		mediaType := strings.TrimSpace(rf.Attrs["media-type"])
		if strings.EqualFold(mediaType, "application/oebps-package+xml") || strings.EqualFold(mediaType, "application/epub+zip") || mediaType == "" {
			return fullPath, nil
		}
	}
	return "", errors.New("no root file found")
}
