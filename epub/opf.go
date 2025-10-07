package epub

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

const (
	TOCTypeEPUB2   = "EPUB2"
	TOCTypeEPUB3   = "EPUB3"
	TOCTypeUnknown = "UNKNOWN"
)

// Manifest 对应 OPF manifest 区域 / Manifest models the OPF manifest section.
type Manifest struct {
	Items []EmptyXmlNode
}

// Spine 对应 OPF spine 区域 / Spine models the OPF spine section.
type Spine struct {
	Itemrefs []EmptyXmlNode
	Attrs    map[string]string
}

// MetaEntry 存储 metadata 元素 / MetaEntry stores metadata element values and attributes.
type MetaEntry struct {
	Value string            // 文本内容 / Text content
	Attrs map[string]string // 所有属性 / Attributes (refines, property, scheme, id, etc.)
}

// Metadata 统一存储 namespace -> tag -> entries / Metadata maps namespaces and tags to entries.
type Metadata struct {
	Data map[string]map[string][]MetaEntry
}

type Opf struct {
	XmlNode  *XmlNode
	Metadata *Metadata
	Manifest *Manifest
	Spine    *Spine
}

// TODO complement methods for Opf, Metadata, Manifest, Spine types if needed

// GetAll 返回简化后的元数据视图 / GetAll flattens metadata into a simple key/value representation.
func (md *Metadata) GetAll() map[string][]string {
	if md == nil || md.Data == nil {
		return map[string][]string{}
	}
	metadata := make(map[string][]string)
	for _, ns := range md.Data {
		for tag, entries := range ns {
			for _, e := range entries {
				val := e.Value
				if val == "" {
					val = e.Attrs["content"]
				}
				if val == "" {
					continue
				}
				key := tag
				if tag == "meta" {
					if name := e.Attrs["name"]; name != "" {
						key = "meta:" + name
					} else if prop := e.Attrs["property"]; prop != "" {
						key = "meta:" + prop
					}
				}
				metadata[key] = append(metadata[key], val)
			}
		}
	}
	return metadata
}

// Normalize 提供标准化后的元数据视图 / Normalize normalizes metadata keys for compatibility.
func (md *Metadata) Normalize() map[string][]string {
	if md == nil {
		return map[string][]string{}
	}
	flat := make(map[string][]string)
	for _, ns := range md.Data {
		for tag, entries := range ns {
			if in(dublinCoreElements, tag) {
				for _, e := range entries {
					if value := strings.TrimSpace(e.Value); value != "" {
						flat[tag] = append(flat[tag], value)
					}
				}
			}
		}
	}
	return flat
}

// HrefLookup 构建 id->href 映射 / HrefLookup builds an id to href lookup map.
func (mf *Manifest) HrefLookup(opfPath string) map[string]string {
	lookup := make(map[string]string)
	if mf == nil {
		return lookup
	}
	baseDir := path.Dir(opfPath)
	if baseDir == "." {
		baseDir = ""
	}
	for _, item := range mf.Items {
		id := strings.TrimSpace(item.Attrs["id"])
		href := strings.TrimSpace(item.Attrs["href"])
		if id == "" || href == "" {
			continue
		}
		lookup[id] = resolveRelative(baseDir, href)
	}
	return lookup
}

// ExtractChapterIDs 提取 spine 中的章节 ID / extractChapterIDs returns ordered chapter IDs from the spine.
func (spine *Spine) ExtractChapterIDs() []string {
	if spine == nil {
		return nil
	}
	var ids []string
	for _, item := range spine.Itemrefs {
		id := strings.TrimSpace(item.Attrs["idref"])
		if id == "" {
			continue
		}
		if linear, ok := item.Attrs["linear"]; ok && strings.EqualFold(linear, "no") {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

// ParseMetadata 解析 metadata 节点 / ParseMetadata extracts the metadata node from the OPF root.
func (opf *Opf) ParseMetadata() error {
	root := opf.XmlNode
	if root == nil {
		return fmt.Errorf("metadata root is nil")
	}
	md := &Metadata{Data: make(map[string]map[string][]MetaEntry)}

	var mdNode *XmlNode
	if mdNode = root.FindNode("metadata"); mdNode == nil {
		return fmt.Errorf("metadata not found")
	}

	defaultNS := mdNode.XMLName.Space
	for i := range mdNode.XmlNodes {
		child := &mdNode.XmlNodes[i]
		ns := child.XMLName.Space
		tag := child.XMLName.Local
		if ns == "" {
			ns = defaultNS
		}

		entry := MetaEntry{
			Value: strings.TrimSpace(child.NodeText()),
			Attrs: make(map[string]string),
		}
		for _, attr := range child.Attrs {
			entry.Attrs[attr.Name.Local] = attr.Value
		}

		if _, ok := md.Data[ns]; !ok {
			md.Data[ns] = make(map[string][]MetaEntry)
		}
		md.Data[ns][tag] = append(md.Data[ns][tag], entry)
	}

	opf.Metadata = md
	return nil
}

// ParseManifest 解析 manifest 节点 / ParseManifest parses the manifest section of the OPF document.
func (opf *Opf) ParseManifest() error {
	root := opf.XmlNode
	if root == nil {
		return fmt.Errorf("manifest root is nil")
	}

	var manifestNode *XmlNode
	if manifestNode = root.FindNode("manifest"); manifestNode == nil {
		return fmt.Errorf("manifest not found")
	}

	manifest := &Manifest{}
	for i := range manifestNode.XmlNodes {
		child := &manifestNode.XmlNodes[i]
		if child.XMLName.Local != "item" {
			continue
		}
		attrs := make(map[string]string)
		for _, attr := range child.Attrs {
			attrs[attr.Name.Local] = attr.Value
		}
		manifest.Items = append(manifest.Items, EmptyXmlNode{
			Name:  "item",
			Attrs: attrs,
		})
	}

	opf.Manifest = manifest
	return nil
}

// ParseSpine 解析 spine 节点 / ParseSpine parses the spine section of the OPF document.
func (opf *Opf) ParseSpine() error {
	root := opf.XmlNode
	if root == nil {
		return fmt.Errorf("spine root is nil")
	}

	var spineNode *XmlNode
	if spineNode = root.FindNode("spine"); spineNode == nil {
		return fmt.Errorf("spine not found")
	}

	spine := &Spine{Attrs: make(map[string]string)}
	for _, attr := range spineNode.Attrs {
		spine.Attrs[attr.Name.Local] = attr.Value
	}

	for i := range spineNode.XmlNodes {
		child := &spineNode.XmlNodes[i]
		if child.XMLName.Local != "itemref" {
			continue
		}
		attrs := make(map[string]string)
		for _, attr := range child.Attrs {
			attrs[attr.Name.Local] = attr.Value
		}
		spine.Itemrefs = append(spine.Itemrefs, EmptyXmlNode{
			Name:  "itemref",
			Attrs: attrs,
		})
	}

	opf.Spine = spine
	return nil
}

func (opf *Opf) FindTOCFile(opfPath string) (tocType string, tocFile string) {
	baseDir := path.Dir(opfPath)
	if baseDir == "." {
		baseDir = ""
	}
	manifest := opf.Manifest
	spine := opf.Spine

	if manifest == nil {
		manifest = &Manifest{}
	}
	if spine == nil {
		spine = &Spine{}
	}

	// 优先检测 EPUB3 nav 属性 / Prefer EPUB3 nav properties.
	for _, item := range manifest.Items {
		if props, ok := item.Attrs["properties"]; ok {
			for _, p := range strings.Fields(props) {
				if strings.EqualFold(p, "nav") {
					return TOCTypeEPUB3, resolveRelative(baseDir, item.Attrs["href"])
				}
			}
		}
	}

	// Spine toc 属性 (EPUB2) / EPUB2 spine toc attribute.
	if id, ok := spine.Attrs["toc"]; ok {
		for _, item := range manifest.Items {
			if item.Attrs["id"] == id {
				return TOCTypeEPUB2, resolveRelative(baseDir, item.Attrs["href"])
			}
		}
	}

	// 回退到 media-type 判断 (EPUB2) / Fallback to media-type detection.
	for _, item := range manifest.Items {
		if strings.EqualFold(item.Attrs["media-type"], "application/x-dtbncx+xml") {
			return TOCTypeEPUB2, resolveRelative(baseDir, item.Attrs["href"])
		}
	}
	return TOCTypeUnknown, ""
}

// resolveRelative 计算相对路径 / resolveRelative resolves manifest-relative paths.
func resolveRelative(baseDir, href string) string {
	if baseDir == "" {
		return path.Clean(href)
	}
	return path.Clean(path.Join(baseDir, href))
}

// in 判断有序切片中是否包含目标字符串 / in checks if the sorted slice contains the target string.
func in(sortedValues []string, target string) bool {
	index := sort.SearchStrings(sortedValues, target)
	return index < len(sortedValues) && sortedValues[index] == target
}
