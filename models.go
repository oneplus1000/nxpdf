package gopdf

import "fmt"

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
	use   int // 1 = name , 2 = index , 3 = stream
	name  string
	index int
}

type nodeContent struct {
	use    int // 1 = str , 2 refTp , 3 = stream
	str    string
	refTo  objectID
	stream []byte
}