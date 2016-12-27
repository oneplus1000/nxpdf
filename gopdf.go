package gopdf

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/oneplus1000/pdf"
	"github.com/pkg/errors"
)

//ReadPdf read pdf file into PdfData
func ReadPdf(rd io.Reader) (*PdfData, error) {
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll fail")
	}

	byteReader := bytes.NewReader(b)
	pdfReader, err := pdf.NewReader(byteReader, byteReader.Size())
	if err != nil {
		return nil, errors.Wrap(err, "pdf.NewReader(...) fail")
	}
	return unmarshal(pdfReader)
}

func MergePdf() error {
	return nil
}

func EmptyPdf() {

}
