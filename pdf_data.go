package nxpdf

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

//PdfData hold data of pdf file
type PdfData struct {
	objects map[objectID]*pdfNodes
}

func newPdfData() *PdfData {
	var p PdfData
	p.objects = make(map[objectID]*pdfNodes)
	return &p
}

func (p *PdfData) push(myID objectID, node pdfNode) {
	if _, ok := p.objects[myID]; ok {
		p.objects[myID].append(node)
	} else {
		var nodes pdfNodes
		nodes.append(node)
		p.objects[myID] = &nodes
	}
}

//Bytes return []byte of pdf file
func (p PdfData) Bytes() ([]byte, error) {
	var buff bytes.Buffer
	var buffTrailer bytes.Buffer
	var realIDs []int
	for objID := range p.objects {
		if objID.isReal {
			realIDs = append(realIDs, int(objID.id))
		}
	}
	sort.Ints(realIDs)
	buff.WriteString("%PDF-1.7")
	var xreftable []int
	for _, realID := range realIDs {
		realObjID := initObjectIDReal(uint32(realID))
		if realID > 0 {
			buff.WriteString("\n")
			xreftable = append(xreftable, buff.Len())
			buff.WriteString(fmt.Sprintf("%d 0 obj", realID))
			data, err := p.bytesOfNodesByID(realObjID)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			buff.Write(data)
			buff.WriteString("\nendobj\n")
		} else {
			data, err := p.bytesOfNodesByID(realObjID)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			buffTrailer.Write(data)
		}
	}
	startxref := buff.Len()
	buff.WriteString("\nxref\n")
	buff.Write(p.bytesOfXref(xreftable))
	buff.WriteString("trailer")
	buffTrailer.WriteTo(&buff)
	buff.WriteString("\nstartxref\n")
	buff.WriteString(fmt.Sprintf("%d", startxref))
	buff.WriteString("\n%%EOF\n")

	return buff.Bytes(), nil
}

func (p PdfData) bytesOfXref(xreftable []int) []byte {
	size := len(xreftable)
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("0 %d\n", size+1))
	buff.WriteString("0000000000 65535 f\n")
	for _, xrefrow := range xreftable {
		buff.WriteString(fmt.Sprintf("%s 00000 n\n", formatXrefline(xrefrow)))
	}
	return buff.Bytes()
}

func formatXrefline(n int) string {
	str := strconv.Itoa(n)
	for len(str) < 10 {
		str = "0" + str
	}
	return str
}

func (p PdfData) bytesOfNodesByID(id objectID) ([]byte, error) {

	var buff bytes.Buffer
	nodes := p.objects[id]
	isArray := p.isArrayNodes(nodes)
	indexOfStream, isStream := p.isStream(nodes)
	isSingleValObj := p.isSingleValObjNodes(nodes)
	if isArray {
		buff.WriteString("[")
	} else if isSingleValObj {
		buff.WriteString("\n")
	} else {
		buff.WriteString("\n<<\n")
	}

	if nodes != nil {

		for _, node := range *nodes {
			//key
			if node.key.use == 1 {
				buff.WriteString(fmt.Sprintf("/%s", node.key.name))
			}
			//content
			buff.WriteString(" ")
			if node.content.use == 1 || node.content.use == 4 {
				buff.WriteString(fmt.Sprintf("%s", node.content.str))
			} else if node.content.use == 2 {
				if node.content.refTo.isReal {
					buff.WriteString(fmt.Sprintf("%d 0 R", node.content.refTo.id))
				} else {
					data, err := p.bytesOfNodesByID(node.content.refTo)
					if err != nil {
						return nil, errors.Wrap(err, "")
					}
					buff.Write(data)
				}
			}

			if !isArray && !isSingleValObj {
				buff.WriteString("\n")
			}
		}

	} //end nodes != nil

	if isArray {
		buff.WriteString(" ]")
	} else if isSingleValObj {
		buff.WriteString("")
	} else {
		buff.WriteString(">>")
	}

	if isStream && indexOfStream != -1 {
		p.writeStream(nodes, indexOfStream, &buff)
	}

	return buff.Bytes(), nil
}

func (p PdfData) writeStream(nodes *pdfNodes, indexOfStream int, buff *bytes.Buffer) {

	stream := (*nodes)[indexOfStream].content.stream
	buff.WriteString("\nstream\n")
	buff.Write(stream)
	if stream[len(stream)-1] != 0xA {
		buff.WriteString("\n")
	}
	buff.WriteString("endstream")

}

func (p PdfData) isArrayNodes(nodes *pdfNodes) bool {
	if nodes == nil {
		return false
	}
	for _, node := range *nodes {
		if node.key.use == 2 {
			return true
		}
	}
	return false
}

func (p PdfData) isSingleValObjNodes(nodes *pdfNodes) bool {
	if nodes == nil {
		return false
	}
	for _, node := range *nodes {
		if node.key.use == 4 {
			return true
		}
	}
	return false
}

func (p PdfData) isStream(nodes *pdfNodes) (int, bool) {
	if nodes == nil {
		return -1, false
	}

	for i, node := range *nodes {
		if node.key.use == 3 {
			return i, true
		}
	}
	return -1, false
}
