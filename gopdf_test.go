package gopdf

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestRead(t *testing.T) {
	testRead(t, "testing/pdf/pdf_from_gopdf.pdf", "testing/out/pdf_from_gopdf_out.pdf")
	testRead(t, "testing/out/pdf_from_gopdf_out.pdf", "testing/out/pdf_from_gopdf_out2.pdf")
	//testRead(t, "testing/pdf/png.pdf")
	//testRead(t, "testing/pdf/twopage.pdf")
}

func testRead(t *testing.T, path string, outpath string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	r := bytes.NewReader(data)
	pdfData, err := ReadPdf(r)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	_ = pdfData
	//fmt.Printf("%s", pdfData.String())
	data, err = pdfData.Bytes()
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	ioutil.WriteFile(outpath, data, 0777)
}
