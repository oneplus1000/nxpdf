package nxpdf

func insertText(p *PdfData, fontRef FontRef, text string, rect *Position, option *TextOption) error {

	_, found := p.subsetFonts[fontRef]
	if !found {
		return ErrFontRefNotFound
	}

	return nil
}
