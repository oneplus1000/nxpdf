package nxpdf

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//PdfData hold data of pdf file
type PdfData struct {
	subsetFonts    map[FontRef](*subsetFont)
	contentCachers []contentCacher
	objects        map[objectID]*pdfNodes
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

//build build pdf
func (p *PdfData) build() error {

	//find all ref
	pagesResults, err := newQuery(p).findDict("Type", "/Pages")
	if err != nil {
		return errors.Wrap(err, "")
	}

	if len(pagesResults) <= 0 {
		return ErrDictNotFound
	}

	kidsNode, err := newQuery(p).findPdfNodeByKeyName(pagesResults[0].objID, "Kids")
	if err != nil {
		return errors.Wrap(err, "")
	}

	kidObjectIDs := make(map[int]objectID)
	kidsNodes := p.objects[kidsNode.content.refTo]
	for i, kid := range *kidsNodes {
		kidObjectIDs[i] = kid.content.refTo
	}

	resObjectIDs := make(map[int]objectID)
	contentObjectIDs := make(map[int]objectID)
	for i, kidObjectID := range kidObjectIDs {

		resNode, err := newQuery(p).findPdfNodeByKeyName(kidObjectID, "Resources")
		if err != nil {
			return errors.Wrap(err, "")
		}
		resObjectIDs[i] = resNode.content.refTo

		contentNode, err := newQuery(p).findPdfNodeByKeyName(kidObjectID, "Contents")
		if err != nil {
			return errors.Wrap(err, "")
		}
		contentObjectIDs[i] = contentNode.content.refTo
	}
	//end find all ref

	err = p.buildSubsetFont(resObjectIDs)
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = p.buildContent()
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func (p *PdfData) buildSubsetFont(resObjectIDs map[int]objectID) error {

	var err error
	maxFakeID, _ := p.findMaxFakeID()
	maxRealID, _ := p.findMaxRealID()

	var newFontObjectIDs []objectID
	for fontRef, ss := range p.subsetFonts {
		var newFontObjectID objectID
		newFontObjectID, maxRealID, maxFakeID, err = p.appendSubsetFont(ss, fontRef, maxRealID, maxFakeID)
		if err != nil {
			return errors.Wrap(err, "")
		}
		newFontObjectIDs = append(newFontObjectIDs, newFontObjectID)
	}

	//append subset font to all res
	resIDs := make(map[objectID]bool)
	for _, objectID := range resObjectIDs {
		if _, ok := resIDs[objectID]; !ok {
			resIDs[objectID] = true
		}
	}

	for resID := range resIDs {
		fontNode, err := newQuery(p).findPdfNodeByKeyName(resID, "Font")
		if err != nil {
			return errors.Wrap(err, "")
		}
		fontNodes := p.objects[fontNode.content.refTo]
		fName := "F"
		fIndexMax := 0
		for _, node := range *fontNodes {
			fIndex := 0
			fName, fIndex, err = p.fontnameExtract(node.key.name)
			if err != nil {
				return errors.Wrap(err, "")
			}
			if fIndex > fIndexMax {
				fIndexMax = fIndex
			}
		}

		for i, newFontObjectID := range newFontObjectIDs {
			fontNode := pdfNode{
				key: nodeKey{
					name: fmt.Sprintf("%s%d", fName, fIndexMax+1+i),
					use:  NodeKeyUseName,
				},
				content: nodeContent{
					use:   NodeContentUseRefTo,
					refTo: newFontObjectID,
				},
			}
			fontNodes.append(fontNode)
		}

	}

	return nil
}

func (p *PdfData) fontnameExtract(fontname string) (string, int, error) {

	rex := regexp.MustCompile("[0-9]+")
	if !rex.MatchString(fontname) {
		return "", 0, fmt.Errorf("can not parse %s", fontname)
	}

	fname := rex.ReplaceAllString(fontname, "")

	findex, err := strconv.Atoi(strings.Replace(fontname, fname, "", -1))
	if err != nil {
		return "", 0, errors.Wrap(err, "")
	}

	return fname, findex, nil
}

func (p *PdfData) buildContent() error {

	var buffContent bytes.Buffer
	for _, cache := range p.contentCachers {
		_, err := cache.build(&buffContent)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	return nil
}

//bytes return []byte of pdf file
func (p PdfData) bytes() ([]byte, error) {
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

func (p PdfData) findMaxFakeID() (uint32, bool) {
	maxFakeID := uint32(0)
	foundFakeID := false
	for objID := range p.objects {
		if !objID.isReal {
			if objID.id >= maxFakeID {
				maxFakeID = objID.id
			}
			foundFakeID = true
		}
	}
	return maxFakeID, foundFakeID
}

func (p PdfData) findMaxRealID() (uint32, bool) {
	maxRealID := uint32(0)
	foundRealID := false
	for objID := range p.objects {
		if objID.isReal {
			if objID.id >= maxRealID {
				maxRealID = objID.id
			}
			foundRealID = true
		}
	}
	return maxRealID, foundRealID
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
			if node.key.use == NodeKeyUseName {
				buff.WriteString(fmt.Sprintf("/%s", node.key.name))
			}
			//content
			buff.WriteString(" ")
			if node.content.use == NodeContentUseString || node.content.use == NodeContentUseSingleObj {
				buff.WriteString(fmt.Sprintf("%s", node.content.str))
			} else if node.content.use == NodeContentUseRefTo {
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
		if node.key.use == NodeKeyUseIndex {
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
		if node.key.use == NodeKeyUseSingleObj {
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
		if node.key.use == NodeKeyUseStream {
			return i, true
		}
	}
	return -1, false
}
