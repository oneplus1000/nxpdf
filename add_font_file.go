package nxpdf

import "github.com/pkg/errors"

func addFontFile(p *PdfData, fontfile []byte) (FontRef, error) {

	hash, err := hashSha1(fontfile)
	if err != nil {
		return FontRefEmpty, errors.Wrap(err, "hashSha1(...) fail")
	}

	if p.subsetFonts == nil {
		p.subsetFonts = make(map[FontRef](*subsetFont))
	}

	var ss subsetFont
	ss.fontfileRaw = fontfile
	p.subsetFonts[FontRef(hash)] = &ss

	var fontRef FontRef
	fontRef = FontRef(hash)
	return fontRef, nil
}
