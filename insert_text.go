package nxpdf

import "github.com/pkg/errors"

func insertText(p *PdfData, fontRef FontRef, text string, rect *Position, option *TextOption) error {

	ssf, found := p.subsetFonts[fontRef]
	if !found {
		return ErrFontRefNotFound
	}

	err := ssf.addChars(text)
	if err != nil {
		return errors.Wrapf(err, "subsetFont.addChars('%s') fail", text)
	}

	cacheText := contenteCacheText{
		ssf:     ssf,
		textRaw: text,
	}
	p.contentCachers = append(p.contentCachers, &cacheText)

	return nil
}
