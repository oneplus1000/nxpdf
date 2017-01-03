package nxpdf

type query struct {
	pdfdata *PdfData
}

func newQuery(p *PdfData) *query {
	var q query
	q.pdfdata = p
	return &q
}

func (q *query) findDict(keyname string, val string) ([]queryResult, error) {
	var results []queryResult
	for objID, nodes := range q.pdfdata.objects {
		for _, node := range *nodes {
			if node.key.use == 1 &&
				node.key.name == keyname &&
				node.content.use == 1 &&
				node.content.str == val {

				var result queryResult
				result.objID = objID
				result.node = node
				results = append(results, result)

			}
		}
	}
	return results, nil
}

type queryResult struct {
	objID objectID
	node  pdfNode
}
