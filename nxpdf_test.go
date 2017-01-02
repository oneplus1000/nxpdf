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
	testRead(t, "testing/out/jpg_out.pdf", "")

	testRead(t, "testing/pdf/pdf_from_chrome_50_win10.pdf", "testing/out/pdf_from_chrome_50_win10_out.pdf")
	testRead(t, "testing/out/pdf_from_chrome_50_win10_out.pdf", "")

	testRead(t, "testing/pdf/pdf_from_docx.pdf", "testing/out/pdf_from_docx_out.pdf")
	testRead(t, "testing/out/pdf_from_docx_out.pdf", "")

	//testRead(t, "testing/pdf/pdf_from_iia.pdf", "testing/out/pdf_from_iia_out.pdf")
	//testRead(t, "testing/out/pdf_from_iia_out.pdf", "")
}

func TestMerge(t *testing.T) {
	//testMerge(t, "testing/pdf/twopage.pdf", "testing/pdf/pdf_from_gopdf.pdf", "testing/out/twopage_and_pdf_from_gopdf_out.pdf")
}

func testMerge(t *testing.T, path1 string, path2 string, outpath string) {
	a, err := read(path1)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	b, err := read(path2)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	c, err := MergePdf(a, b)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}

	data, err := c.Bytes()
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	if outpath != "" {
		ioutil.WriteFile(outpath, data, 0777)
	}
}

func read(path1 string) (*PdfData, error) {
	data, err := ioutil.ReadFile(path1)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	pdfData, err := ReadPdf(r)
	if err != nil {
		return nil, err
	}
	return pdfData, nil
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
