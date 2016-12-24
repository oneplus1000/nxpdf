package gopdf

import (
	"bytes"
	"fmt"
	"sort"
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

func (p PdfData) Bytes() []byte {
	var buff bytes.Buffer
	var realIDs []int
	for objID := range p.objects {
		if objID.isReal {
			realIDs = append(realIDs, int(objID.id))
		}
	}
	sort.Ints(realIDs)
	for _, realID := range realIDs {
		buff.WriteString(fmt.Sprintf("%d 0 obj\n", realID))
		nodes := p.objects[initObjectID(uint32(realID), true)]
		buff.WriteString("<<\n")
		for _, node := range *nodes {
			if node.key.use == 1 {
				buff.WriteString(fmt.Sprintf("/%s", node.key.name))
			}

			if node.content.use == 1 {
				buff.WriteString(fmt.Sprintf(" %s\n", node.content.str))
			} else if node.content.use == 2 {
				if node.content.refTo.isReal {
					buff.WriteString(fmt.Sprintf(" %d 0 R\n", node.content.refTo.id))
				} else {
					items := p.objects[initObjectID(node.content.refTo.id, false)]
					buff.WriteString(" [")
					for _, item := range *items {
						buff.WriteString(item.content.str)
						buff.WriteString(" ")
					}
					buff.WriteString("]\n")
				}
			}
		}
		buff.WriteString(">>\n\n")
	}

	return buff.Bytes()
}

type objectID struct {
	isReal bool
	id     uint32
}

func (o objectID) String() string {
	if o.isReal {
		return fmt.Sprintf("%d", o.id)
	}
	return fmt.Sprintf("%df", o.id)

}

func initObjectID(id uint32, isReal bool) objectID {
	var o objectID
	o.id = id
	o.isReal = isReal
	return o
}

type pdfNodes []pdfNode

func (p *pdfNodes) append(n pdfNode) {
	*p = append(*p, n)
}

type pdfNode struct {
	key     nodeKey
	content nodeContent
}

type nodeKey struct {
	use   int // 1 = name , 2 = index
	name  string
	index int
}

type nodeContent struct {
	use   int // 1 = str , 2 refTp
	str   string
	refTo objectID
}
