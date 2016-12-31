package nxpdf

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestRead(t *testing.T) {
	testRead(t, "testing/pdf/pdf_from_gopdf.pdf", "testing/out/pdf_from_gopdf_out.pdf")
	testRead(t, "testing/out/pdf_from_gopdf_out.pdf", "")

	testRead(t, "testing/pdf/twopage.pdf", "testing/out/twopage_out.pdf")
	testRead(t, "testing/out/twopage_out.pdf", "")

	testRead(t, "testing/pdf/jpg.pdf", "testing/out/jpg_out.pdf")
	//fmt.Printf("----------------\n")
	testRead(t, "testing/out/jpg_out.pdf", "")

	testRead(t, "testing/pdf/pdf_from_chrome_50_win10.pdf", "testing/out/pdf_from_chrome_50_win10_out.pdf")
	testRead(t, "testing/out/pdf_from_chrome_50_win10_out.pdf", "")
}

func TestMerge(t *testing.T) {

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

	if outpath != "" {
		ioutil.WriteFile(outpath, data, 0777)
	}
}
