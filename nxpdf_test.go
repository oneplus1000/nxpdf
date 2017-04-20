package nxpdf

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRead(t *testing.T) {

	isExists, err := exists("testing/out/")
	if err != nil {
		t.Errorf("can not create testing/out/ for test")
		return
	}
	if !isExists {
		err := os.MkdirAll("testing/out/", 0777)
		if err != nil {
			t.Errorf("can not create testing/out/ for test")
			return
		}
	}

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
	testMerge(t, "testing/pdf/twopage.pdf", "testing/pdf/pdf_from_gopdf.pdf", "testing/out/twopage_and_pdf_from_gopdf_out.pdf")
	testMerge(t, "testing/pdf/jpg.pdf", "testing/pdf/pdf_from_docx.pdf", "testing/out/jpg_and_pdf_from_docx_out.pdf")
}

func TestInsertText(t *testing.T) {
	err := testInsertText("testing/pdf/pdf_from_gopdf.pdf", "testing/out/pdf_from_gopdf_out_inserttext.pdf")
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func testInsertText(path string, outpath string) error {

	pdfdata, err := read(path)
	if err != nil {
		return err
	}

	fontRef, err := AddFontFilePath(pdfdata, "testing/ttf/times.ttf")
	if err != nil {
		return err
	}

	err = InsertText(pdfdata, fontRef, "Hello", &Position{PageNum: 1, X: 10, Y: 10}, &TextOption{})
	if err != nil {
		return err
	}

	data, err := pdfdata.bytes()
	if err != nil {
		return err
	}
	if outpath != "" {
		ioutil.WriteFile(outpath, data, 0777)
	}

	return nil
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

	err = MergePdf(a, b)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}

	data, err := BuildPdf(a)
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
	pdfData, err := ReadPdf(data)
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
	pdfData, err := ReadPdf(data)
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	_ = pdfData
	//fmt.Printf("%s", pdfData.String())
	data, err = pdfData.bytes()
	if err != nil {
		t.Errorf("%+v", err)
		return
	}

	if outpath != "" {
		ioutil.WriteFile(outpath, data, 0777)
	}
}
