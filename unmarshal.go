package gopdf

import (
	"fmt"
	"io"

	"io/ioutil"

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
	err := u.doDict(initObjectID(0, true), parent)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (u *unmarshalHelper) doDict(myID objectID, parent pdf.Value) error {

	keys := parent.Keys()
	for _, key := range keys {

		/*if key == "Parent" {
			continue
		}*/

		child := parent.Key(key)
		if child.Kind() == pdf.Dict || child.Kind() == pdf.Stream {
			refID, _ := child.RefTo()
			refObjID := initObjectID(refID, true)
			if refID != 0 && refObjID != myID {
				u.pushRef(myID, key, refObjID)
				if _, ok := u.unmarshalledIDs[refID]; ok {
					fmt.Printf("--------------->%s\n", key)
					continue
				}
				u.unmarshalledIDs[refID] = refObjID
				err := u.doDict(refObjID, child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				if child.Kind() == pdf.Stream {
					err := u.pushStream(refObjID, child.Reader())
					if err != nil {
						return errors.Wrap(err, "")
					}
				}
			} else if refID != 0 && refObjID == myID {
				//fmt.Printf("---------%s %d\n", key, refObjID.id)
				fakeID := u.nextFakeID()
				err := u.doDict(initObjectID(fakeID, false), child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				u.pushRef(myID, key, initObjectID(fakeID, false))
			} else {
				u.pushVal(myID, key, child)
			}

		} else if child.Kind() == pdf.Array {
			fakeID := u.nextFakeID()
			err := u.doArray(initObjectID(fakeID, false), child)
			if err != nil {
				return errors.Wrap(err, "")
			}
			u.pushRef(myID, key, initObjectID(fakeID, false))
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
			refObjID := initObjectID(refID, true)
			if refID != 0 && refObjID != myID {
				u.pushItemRef(myID, i, refObjID)
				if _, ok := u.unmarshalledIDs[refID]; ok {
					continue
				}
				u.unmarshalledIDs[refID] = refObjID
				err := u.doDict(refObjID, child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				if child.Kind() == pdf.Stream {
					err := u.pushStream(refObjID, child.Reader())
					if err != nil {
						return errors.Wrap(err, "")
					}
				}
			} else if refID != 0 && refObjID == myID {
				fakeID := u.nextFakeID()
				err := u.doDict(initObjectID(fakeID, false), child)
				if err != nil {
					return errors.Wrap(err, "")
				}
				u.pushItemRef(myID, i, initObjectID(fakeID, false))
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
			str: val.String(),
		},
	}
	u.result.push(myid, n)
}

func (u *unmarshalHelper) pushStream(myid objectID, r io.ReadCloser) error {
	if printDebug {
		fmt.Printf("%s [stream]\n", myid)
	}

	stream, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer r.Close()

	/*var buff bytes.Buffer
	var zbuff bytes.Buffer
	zw := zlib.NewWriter(&zbuff)
	defer zw.Close()
	_, err = zw.Write(stream)
	if err != nil {
		return errors.Wrap(err, "zlib.Write fail")
	}
	zw.Flush()
	zbuff.WriteTo(&buff)

	fmt.Printf("%s [stream]%d\n%+v", myid, len(stream), buff.Bytes())
	*/
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
			str: val.String(),
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

const printDebug = true