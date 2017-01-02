package nxpdf

import "fmt"

type query struct {
	pdfdata *PdfData
}

func newQuery(p *PdfData) *query {
	var q query
	q.pdfdata = p
	return &q
}

func (q *query) findDict(keyname string, val string) error {
	for objID, nodes := range q.pdfdata.objects {
		for _, node := range *nodes {
			if node.key.use == 1 && node.key.name == keyname && node.content.use == 1 && node.content.str == val {
				fmt.Printf("%d\n", objID.id)
				break
			}
		}
	}
	return nil
}

//
type queryResult struct {
	objID objectID
}
