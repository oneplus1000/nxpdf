package nxpdf

import "fmt"

type objectID struct {
	isReal bool
	id     uint32
	//fromRealID uint32
}

var objectIDEmpty = objectID{}

func (o objectID) String() string {
	if o.isReal {
		return fmt.Sprintf("%d", o.id)
	}
	return fmt.Sprintf("%df", o.id)

}

func initObjectIDReal(id uint32) objectID {
	var o objectID
	o.id = id
	o.isReal = true
	//o.fromRealID = id
	return o
}

func initObjectIDFake(id uint32, fromRealID uint32) objectID {
	var o objectID
	o.id = id
	o.isReal = false
	//o.fromRealID = fromRealID
	return o
}

type pdfNodes []pdfNode

func (p *pdfNodes) len() int {
	return len(*p)
}

func (p *pdfNodes) append(n pdfNode) {
	*p = append(*p, n)
}

func (p *pdfNodes) remove(index int) {
	*p = append((*p)[:index], (*p)[index+1:]...)
}

type pdfNode struct {
	key     nodeKey
	content nodeContent
}

func (p pdfNode) clone() pdfNode {
	return p
}

const (
	NodeKeyUseName      = 1
	NodeKeyUseIndex     = 2
	NodeKeyUseStream    = 3
	NodeKeyUseSingleObj = 4
)

type nodeKey struct {
	use   int // 1 = name , 2 = index , 3 = stream , 4 = single obj
	name  string
	index int
}

const (
	NodeContentUseString    = 1
	NodeContentUseRefTo     = 2
	NodeContentUseStream    = 3
	NodeContentUseSingleObj = 4
)

type nodeContent struct {
	use    int // 1 = str , 2 refTo , 3 = stream , 4 = single obj
	str    string
	refTo  objectID
	stream []byte
}
