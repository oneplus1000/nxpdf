package nxpdf

import "github.com/pkg/errors"

func insertText(p *PdfData, fontRef FontRef, text string, rect *Position, option *TextOption) error {

	subsetFont, found := p.subsetFonts[fontRef]
	if !found {
		return ErrFontRefNotFound
	}

	err := subsetFont.addChars(text)
	if err != nil {
		return errors.Wrapf(err, "subsetFont.addChars('%s') fail", text)
	}

	return nil
}
