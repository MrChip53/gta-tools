package img

type ImgEntry struct {
	idx  int
	name string
	toc  TocEntry
	data []byte
}

func (e ImgEntry) Name() string { return e.name }

func (e ImgEntry) Data() []byte {
	d := make([]byte, len(e.data))
	copy(d, e.data)
	return d
}

func (e ImgEntry) Toc() TocEntry { return e.toc }

func (e *ImgEntry) SetData(data []byte) { e.data = data }

func (e ImgEntry) Index() int { return e.idx }
