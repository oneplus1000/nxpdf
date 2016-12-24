package gopdf

import (
	"fmt"
	"io"

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
	trailer pdf.Value
	result  *PdfData
	srcIDs  map[uint32]int
	fakeID  uint32
}

func newUnmarshalHelper(trailer pdf.Value) *unmarshalHelper {
	var uh unmarshalHelper
	uh.trailer = trailer
	uh.result = newPdfData()
	uh.srcIDs = make(map[uint32]int)
	uh.fakeID = 4000
	return &uh
}

func (u *unmarshalHelper) start() error {
	parent := u.trailer
	err := u.doDict(initObjectID(0, true), parent)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (u *unmarshalHelper) doDict(myID objectID, parent pdf.Value) error {

	keys := parent.Keys()
	for _, key := range keys {
		if key == "Parent" {
			continue
		}
		child := parent.Key(key)
		if child.Kind() == pdf.Dict || child.Kind() == pdf.Stream {
			refID, _ := child.RefTo()
			if refID != 0 {
				u.pushRef(myID, key, initObjectID(refID, true))
				err := u.doDict(initObjectID(refID, true), child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				if child.Kind() == pdf.Stream {
					u.pushStream(initObjectID(refID, true), child.Reader())
				}
			} else {
				u.pushVal(myID, key, child)
			}

		} else if child.Kind() == pdf.Array {
			fakeID := u.nextFakeID()
			err := u.doArray(initObjectID(fakeID, false), child)
			u.pushRef(myID, key, initObjectID(fakeID, false))
			if err != nil {
				return errors.Wrap(err, "")
			}
		} else {
			u.pushVal(myID, key, child)
		}
	}
	return nil
}

func (u *unmarshalHelper) nextFakeID() uint32 {
	u.fakeID++
	return u.fakeID
}

func (u *unmarshalHelper) doArray(myID objectID, parent pdf.Value) error {

	size := parent.Len()
	for i := 0; i < size; i++ {
		child := parent.Index(i)
		if child.Kind() == pdf.Dict || child.Kind() == pdf.Stream {
			refID, _ := child.RefTo()
			if refID != 0 {
				u.pushItemRef(myID, i, initObjectID(refID, true))
				err := u.doDict(initObjectID(refID, true), child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				if child.Kind() == pdf.Stream {
					u.pushStream(initObjectID(refID, true), child.Reader())
				}
			} else {
				u.pushItemVal(myID, i, child)
			}

		} else if child.Kind() == pdf.Array {
			fakeID := u.nextFakeID()
			err := u.doArray(initObjectID(fakeID, false), child)
			u.pushItemRef(myID, i, initObjectID(fakeID, false))
			if err != nil {
				return errors.Wrap(err, "")
			}
		} else {
			u.pushItemVal(myID, i, child)
		}
	}

	return nil
}

func (u *unmarshalHelper) pushVal(myid objectID, name string, val pdf.Value) {
	fmt.Printf("%s %s %s\n", myid, name, val.String())
	n := pdfNode{
		key: nodeKey{
			use:  1,
			name: name,
		},
		content: nodeContent{
			use: 1,
			str: val.String(),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushStream(myid objectID, r io.ReadCloser) {
	fmt.Printf("%s [stream]\n", myid)

}

func (u *unmarshalHelper) pushItemVal(myid objectID, index int, val pdf.Value) {
	fmt.Printf("%s [%d] %s\n", myid, index, val.String())
	n := pdfNode{
		key: nodeKey{
			use:   2,
			index: index,
		},
		content: nodeContent{
			use: 1,
			str: val.String(),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushItemRef(myid objectID, index int, refID objectID) {
	fmt.Printf("%s [%d] '%s 0 R'\n", myid, index, refID)
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
	fmt.Printf("%s %s '%s 0 R'\n", myid, name, refID)
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
