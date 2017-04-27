package nxpdf

import "github.com/pkg/errors"

func insertText(p *PdfData, fontRef FontRef, text string, pageIndex int /* zero to n..*/, rect *Position, option *TextOption) error {

	ssf, found := p.subsetFonts[fontRef]
	if !found {
		return ErrFontRefNotFound
	}

	err := ssf.addChars(text)
	if err != nil {
		return errors.Wrapf(err, "subsetFont.addChars('%s') fail", text)
	}

	if p.mapPageAndContentCachers == nil {
		p.mapPageAndContentCachers = make(map[int](*[]contentCacher))
	}

	ccText := contenteCacheText{
		ssf:     ssf,
		textRaw: text,
	}

	if contentCachers, ok := p.mapPageAndContentCachers[pageIndex]; ok {
		*contentCachers = append(*contentCachers, &ccText)
	} else {
		contentCachers := []contentCacher{
			&ccText,
		}
		p.mapPageAndContentCachers[pageIndex] = &contentCachers
	}

	return nil
}
