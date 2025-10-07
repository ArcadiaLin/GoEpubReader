package epub

import (
	"errors"
	"fmt"
	"strings"
)

type Book struct {
	Container *Container `json:"container,omitempty"`
	Opf       *Opf       `json:"opf,omitempty"`
	TOC       *TOC       `json:"toc,omitempty"`
	Chapters  []Chapter  `json:"chapters,omitempty"`
}

var (
	// ErrMetadataUndefined indicates that a requested metadata field is not
	// present in the document.
	ErrMetadataUndefined = errors.New("metadata not defined")
	// ErrChapterNotFound indicates that the requested chapter could not be
	// located either by ID or by index.
	ErrChapterNotFound = errors.New("chapter not found")
)

// MetadataValues returns all values for the given Dublin Core metadata key.
func (b *Book) MetadataValues(key string) ([]string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return nil, fmt.Errorf("metadata key is empty")
	}
	if b == nil || b.Opf == nil || b.Opf.Metadata == nil {
		return nil, fmt.Errorf("%w: %s", ErrMetadataUndefined, key)
	}
	values := b.Opf.Metadata.Get(key)
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrMetadataUndefined, key)
	}
	return values, nil
}

// MetadataValue returns the first value for the provided Dublin Core key.
func (b *Book) MetadataValue(key string) (string, error) {
	values, err := b.MetadataValues(key)
	if err != nil {
		return "", err
	}
	return values[0], nil
}

func (b *Book) metadataGetter(key string) func() (string, error) {
	return func() (string, error) {
		return b.MetadataValue(key)
	}
}

// The following helpers expose a method for each Dublin Core metadata field.
func (b *Book) Title() (string, error)       { return b.metadataGetter("title")() }
func (b *Book) Creator() (string, error)     { return b.metadataGetter("creator")() }
func (b *Book) Subject() (string, error)     { return b.metadataGetter("subject")() }
func (b *Book) Description() (string, error) { return b.metadataGetter("description")() }
func (b *Book) Publisher() (string, error)   { return b.metadataGetter("publisher")() }
func (b *Book) Contributor() (string, error) { return b.metadataGetter("contributor")() }
func (b *Book) Date() (string, error)        { return b.metadataGetter("date")() }
func (b *Book) Type() (string, error)        { return b.metadataGetter("type")() }
func (b *Book) Format() (string, error)      { return b.metadataGetter("format")() }
func (b *Book) Identifier() (string, error)  { return b.metadataGetter("identifier")() }
func (b *Book) Language() (string, error)    { return b.metadataGetter("language")() }
func (b *Book) Source() (string, error)      { return b.metadataGetter("source")() }
func (b *Book) Relation() (string, error)    { return b.metadataGetter("relation")() }
func (b *Book) Coverage() (string, error)    { return b.metadataGetter("coverage")() }
func (b *Book) Rights() (string, error)      { return b.metadataGetter("rights")() }

// MetadataByKey provides direct access to Dublin Core metadata using a dynamic
// key. It is a convenience wrapper around MetadataValues and should be used by
// callers that need to iterate over keys.
func (b *Book) MetadataByKey(key string) ([]string, error) {
	return b.MetadataValues(key)
}

// AllMetadata returns all metadata entries, including extension fields defined
// in the OPF <meta> tags.
func (b *Book) AllMetadata() map[string][]string {
	if b == nil || b.Opf == nil || b.Opf.Metadata == nil {
		return map[string][]string{}
	}
	result := b.Opf.Metadata.GetAll()
	out := make(map[string][]string, len(result))
	for k, v := range result {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// ChapterCount returns the number of parsed chapters.
func (b *Book) ChapterCount() int {
	if b == nil {
		return 0
	}
	return len(b.Chapters)
}

// ChapterByID returns the chapter that matches the provided ID.
func (b *Book) ChapterByID(id string) (*Chapter, error) {
	if b == nil {
		return nil, ErrChapterNotFound
	}
	for i := range b.Chapters {
		if b.Chapters[i].ID == id {
			return &b.Chapters[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrChapterNotFound, id)
}

// ChapterByIndex returns the chapter by its ordinal index (0-based).
func (b *Book) ChapterByIndex(index int) (*Chapter, error) {
	if b == nil || index < 0 || index >= len(b.Chapters) {
		return nil, fmt.Errorf("%w: index %d", ErrChapterNotFound, index)
	}
	return &b.Chapters[index], nil
}

// ChapterTextByID returns the joined text content of the chapter with the
// provided ID.
func (b *Book) ChapterTextByID(id string) (string, error) {
	chapter, err := b.ChapterByID(id)
	if err != nil {
		return "", err
	}
	return chapter.Text(), nil
}

// ChapterTextByIndex returns the joined text content of the chapter at the
// provided index.
func (b *Book) ChapterTextByIndex(index int) (string, error) {
	chapter, err := b.ChapterByIndex(index)
	if err != nil {
		return "", err
	}
	return chapter.Text(), nil
}

// AllChaptersText concatenates every chapter's text content in reading order.
func (b *Book) AllChaptersText() string {
	if b == nil {
		return ""
	}
	var texts []string
	for i := range b.Chapters {
		if text := b.Chapters[i].Text(); text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "\n\n")
}

// FlattenTOC returns the table of contents entries as a slice, skipping the
// synthetic root node if present.
func (b *Book) FlattenTOC() []TOC {
	if b == nil || b.TOC == nil {
		return nil
	}
	entries := b.TOC.Flatten()
	if len(entries) == 0 {
		return nil
	}
	// Skip the root node which only exists as a container.
	return entries[1:]
}
