package nxpdf

import (
	"bytes"
	"io/ioutil"

	"github.com/oneplus1000/pdf"
	"github.com/pkg/errors"
)

//ReadPdf read pdf file into PdfData
func ReadPdf(pdffile []byte) (*PdfData, error) {
	byteReader := bytes.NewReader(pdffile)
	pdfReader, err := pdf.NewReader(byteReader, byteReader.Size())
	if err != nil {
		return nil, errors.Wrap(err, "pdf.NewReader(...) fail")
	}
	return unmarshal(pdfReader)
}

//AddFontFile add font file into pdf file
func AddFontFile(p *PdfData, fontfile []byte) (FontRef, error) {
	return addFontFile(p, fontfile)
}

//AddFontFilePath add font file into pdf file by path of fontfile
func AddFontFilePath(p *PdfData, fontpath string) (FontRef, error) {
	b, err := ioutil.ReadFile(fontpath)
	if err != nil {
		return FontRefEmpty, errors.Wrapf(err, "ioutil.ReadFile(%s) fail", fontpath)
	}
	return AddFontFile(p, b)
}

//InsertText insert text to pdf
func InsertText(p *PdfData, fontRef FontRef, text string, rect *Position, option *TextOption) error {
	return insertText(p, fontRef, text, rect, option)
}

//MergePdf merge b into a
func MergePdf(a, b *PdfData) error {
	return merge(a, b)
}

//BuildPdf create pdf file
func BuildPdf(p *PdfData) ([]byte, error) {

	err := p.build()
	if err != nil {
		return nil, errors.Wrap(err, "p.build() fail")
	}

	b, err := p.bytes()
	if err != nil {
		return nil, errors.Wrap(err, "p.Byte() fail")
	}

	return b, nil
}
