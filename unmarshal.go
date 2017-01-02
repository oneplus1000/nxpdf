package nxpdf

import (
	"bytes"
	"fmt"

	"io/ioutil"

	"strings"

	"github.com/oneplus1000/pdf"
	"github.com/pkg/errors"
)

func unmarshal(rd *pdf.Reader) (*PdfData, error) {
	uh := newUnmarshalHelper(rd.Trailer())
	err := uh.start()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return uh.result, nil
}

type unmarshalHelper struct {
	trailer         pdf.Value
	result          *PdfData
	unmarshalledIDs map[uint32]objectID
	fakeID          uint32
}

func newUnmarshalHelper(trailer pdf.Value) *unmarshalHelper {
	var uh unmarshalHelper
	uh.trailer = trailer
	uh.result = newPdfData()
	uh.unmarshalledIDs = make(map[uint32]objectID)
	uh.fakeID = 4000
	return &uh
}

func (u *unmarshalHelper) start() error {
	parent := u.trailer
	objID := initObjectID(0, true)
	/*err := u.doDict(initObjectID(0, true), parent)
	if err != nil {
		return errors.Wrap(err, "")
	}*/
	err := u.doing(objID, 0, parent)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (u *unmarshalHelper) doing(myID objectID, fromRealID uint32, parent pdf.Value) error {

	parentKeys := parent.Keys()
	parentSize := 0
	parentKind := parent.Kind()
	if parentKind == pdf.Array {
		parentSize = parent.Len()
	} else if parentKind == pdf.Dict || parentKind == pdf.Stream {
		parentSize = len(parentKeys)
	}

	for i := 0; i < parentSize; i++ {

		var child pdf.Value
		var childKey string
		if parentKind == pdf.Array {
			child = parent.Index(i)
		} else if parentKind == pdf.Dict || parentKind == pdf.Stream {
			child = parent.Key(parentKeys[i])
			childKey = parentKeys[i]
		}

		childKind := child.Kind()
		childRefID, _ := child.RefTo()
		if childKind == pdf.Dict || childKind == pdf.Array || childKind == pdf.Stream {
			if isEmbedObj(myID, fromRealID, childRefID) {
				fakeRefObjID := initObjectID(u.nextFakeID(), false)
				if parentKind == pdf.Array {
					u.pushItemRef(myID, i, fakeRefObjID)
				} else if parentKind == pdf.Dict {
					u.pushRef(myID, childKey, fakeRefObjID)
				}
				err := u.doing(fakeRefObjID, fromRealID, child)
				if err != nil {
					return errors.Wrap(err, "")
				}

				if childKind == pdf.Stream {
					err := u.pushStream(fakeRefObjID, child)
					if err != nil {
						return errors.Wrap(err, "")
					}
				}

			} else {
				isDup := false
				childRefObjID := initObjectID(childRefID, true)
				if oldChildRefObjID, ok := u.unmarshalledIDs[childRefID]; ok {
					childRefObjID = oldChildRefObjID
					isDup = true
				} else {
					u.unmarshalledIDs[childRefID] = childRefObjID
				}
				if parentKind == pdf.Array {
					u.pushItemRef(myID, i, childRefObjID)
				} else if parentKind == pdf.Dict {
					u.pushRef(myID, childKey, childRefObjID)
				}
				if isDup {
					continue
				}
				err := u.doing(childRefObjID, childRefObjID.id, child)
				if err != nil {
					return errors.Wrap(err, "")
				}

				if childKind == pdf.Stream {
					err := u.pushStream(childRefObjID, child)
					if err != nil {
						return errors.Wrap(err, "")
					}
				}
			}

		} else {

			if isEmbedObj(myID, fromRealID, childRefID) {
				if parentKind == pdf.Array {
					u.pushItemVal(myID, i, child)
				} else if parentKind == pdf.Dict || parentKind == pdf.Stream {
					u.pushVal(myID, childKey, child)
				}
			} else {
				childRefObjID := initObjectID(childRefID, true)
				if parentKind == pdf.Array {
					u.pushItemRef(myID, i, childRefObjID)
				} else if parentKind == pdf.Dict || parentKind == pdf.Stream {
					u.pushRef(myID, childKey, childRefObjID)
				}
				u.pushSingleValObj(childRefObjID, "", child)
			}
		}
	}

	if myID.isReal && parentSize == 0 { //realID but empty
		var empty pdfNodes
		u.result.objects[myID] = &empty
	}

	return nil
}

func isEmbedObj(myID objectID, fromRealID uint32, childRefID uint32) bool {
	childRefObjID := initObjectID(childRefID, true)
	if myID.isReal {
		if myID == childRefObjID { //embed
			return true
		}
		return false //ref
	}

	fromRealObjID := initObjectID(fromRealID, true)
	if fromRealObjID == childRefObjID { //embed
		return true
	}
	return false //ref

}

func (u *unmarshalHelper) nextFakeID() uint32 {
	u.fakeID++
	return u.fakeID
}

func (u *unmarshalHelper) pushVal(myid objectID, name string, val pdf.Value) {
	if printDebug {
		fmt.Printf("%s %s %s\n", myid, name, val.String())
	}
	n := pdfNode{
		key: nodeKey{
			use:  1,
			name: name,
		},
		content: nodeContent{
			use: 1,
			str: format(val),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushSingleValObj(myid objectID, name string, val pdf.Value) {
	if printDebug {
		fmt.Printf("%s %s %s\n", myid, name, val.String())
	}
	n := pdfNode{
		key: nodeKey{
			use:  4,
			name: name,
		},
		content: nodeContent{
			use: 4,
			str: format(val),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushStream(myid objectID, val pdf.Value) error {

	rd := val.RawReader()
	stream, err := ioutil.ReadAll(rd)
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer rd.Close()

	if printDebug {
		fmt.Printf("%s [stream=%d]\n", myid, len(stream))
	}

	n := pdfNode{
		key: nodeKey{
			use: 3,
		},
		content: nodeContent{
			use:    3,
			stream: stream,
		},
	}
	u.result.push(myid, n)

	return nil
}

func (u *unmarshalHelper) pushItemVal(myid objectID, index int, val pdf.Value) {
	if printDebug {
		fmt.Printf("%s [%d] %s\n", myid, index, val.String())
	}
	n := pdfNode{
		key: nodeKey{
			use:   2,
			index: index,
		},
		content: nodeContent{
			use: 1,
			str: format(val),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushItemRef(myid objectID, index int, refID objectID) {
	if printDebug {
		fmt.Printf("%s [%d] '%s 0 R'\n", myid, index, refID)
	}
	n := pdfNode{
		key: nodeKey{
			use:   2,
			index: index,
		},
		content: nodeContent{
			use:   2,
			refTo: refID,
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushRef(myid objectID, name string, refID objectID) {
	if printDebug {
		fmt.Printf("%s %s '%s 0 R'\n", myid, name, refID)
	}
	n := pdfNode{
		key: nodeKey{
			use:  1,
			name: name,
		},
		content: nodeContent{
			use:   2,
			refTo: refID,
		},
	}
	u.result.push(myid, n)
}

func format(val pdf.Value) string {
	var data string
	if val.Kind() == pdf.String {
		if val.StringType() == 3 {
			var buff bytes.Buffer
			str := val.TextFromUTF16()
			buff.WriteString("<")
			for _, ru := range str {
				buff.WriteString(digit(fmt.Sprintf("%X", ru), 4))
			}
			buff.WriteString(">")
			data = buff.String()
		} else {
			data = fmt.Sprintf("(%s)", cleanString(val.String()))
		}
	} else if val.Kind() == pdf.Null {
		data = "null"
	} else {
		data = val.String()
		data = strings.Replace(data, " ", "#20", -1)
	}

	return data
}

func cleanString(str string) string {
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, "(", "\\(", -1)
	str = strings.Replace(str, ")", "\\)", -1)
	return str
}

func digit(n string, digit int) string {
	size := len(n)
	var buff bytes.Buffer
	for size < digit {
		buff.WriteString("0")
		size++
	}
	buff.WriteString(n)
	return buff.String()
}

const printDebug = false
