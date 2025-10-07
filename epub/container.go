package epub

import (
	"errors"
	"strings"
)

type Container struct {
	Rootfiles []EmptyXmlNode
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
