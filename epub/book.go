package epub

type Book struct {
	Container *Container `json:"container,omitempty"`
	Opf       *Opf       `json:"opf,omitempty"`
	TOC       *TOC       `json:"toc,omitempty"`
	Chapters  []Chapter  `json:"chapters,omitempty"`
}

// TODO design some Methods fo Book Type if needed

// TODO for every dublincore metadata key provide a simple interface to return the metadata, if there is no value just tell user it is not defined

// TODO design a interface extract metadata by dublincore keys, this interface can reuse the interface by previous ToDO

// TODO extract chapter content by id or order

// TODO extract all Metadata, also include some extra information beyond dublincore

// TODO extract all chapter text content and concat by order
