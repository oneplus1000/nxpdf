package nxpdf

import "github.com/pkg/errors"

func merge(a, b *PdfData) (*PdfData, error) {
	maxRealIDOfA, maxFakeIDOfA, err := maxID(a)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	newB, err := shiftID(b, maxRealIDOfA+1, maxFakeIDOfA+1)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	c := (*a)
	for objID, obj := range newB.objects {
		c.objects[objID] = obj
	}

	return &c, nil
}

func shiftID(src *PdfData, realIDOffset uint32, fakeIDOffset uint32) (*PdfData, error) {
	dest := newPdfData()
	for srcID := range src.objects {
		var destID objectID
		destID.isReal = srcID.isReal
		if destID.isReal {
			destID.id = srcID.id + realIDOffset
		} else {
			destID.id = srcID.id + fakeIDOffset
		}
		srcNodes := src.objects[srcID]
		size := srcNodes.len()
		for i := 0; i < size; i++ {
			srcNode := (*srcNodes)[i]
			destNode := srcNode.clone()
			if destNode.content.use == 2 {
				if destNode.content.refTo.isReal {
					destNode.content.refTo.id = destNode.content.refTo.id + realIDOffset
				} else {
					destNode.content.refTo.id = destNode.content.refTo.id + fakeIDOffset
				}
			}
			dest.push(destID, destNode)
		}
	}
	return dest, nil
}

func maxID(a *PdfData) (uint32, uint32, error) {
	maxRealID := uint32(0)
	maxFakeID := uint32(0)
	for objID := range a.objects {
		if objID.id > maxRealID {
			maxRealID = objID.id
		}
		if objID.id > maxFakeID {
			maxFakeID = objID.id
		}
	}
	return maxRealID, maxRealID, nil
}
