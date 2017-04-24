package nxpdf

import "io"

type contentCacher interface {
	writeTo(w io.Writer) (int64, error)
}

//contenteCacheText

type contenteCacheText struct {
	ssf     *subsetFont
	textRaw string
}

func (c *contenteCacheText) writeTo(w io.Writer) (int64, error) {

	for range c.textRaw {
		//c.textRaw
	}

	return 0, nil
}
