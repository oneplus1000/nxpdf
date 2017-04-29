package nxpdf

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/signintech/gopdf/fontmaker/core"
)

func (p *PdfData) appendSubsetFont(ssf *subsetFont, fontRef FontRef, maxRealID uint32, maxFakeID uint32) (objectID, uint32, uint32, error) {

	ssfNodes := pdfNodes{}

	maxRealID++
	var ssfNodesObjectID = objectID{
		id:     maxRealID,
		isReal: true,
	}
	p.objects[ssfNodesObjectID] = &ssfNodes

	typeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/Font",
		},
	}

	subtypeNode := pdfNode{
		key: nodeKey{
			name: "Subtype",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/Type0",
		},
	}

	baseFontNode := pdfNode{
		key: nodeKey{
			name: "BaseFont",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	encodingNode := pdfNode{
		key: nodeKey{
			name: "Encoding",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/Identity-H",
		},
	}

	maxFakeID++
	descendantFontsNodeItemRefID := objectID{
		id:     maxFakeID,
		isReal: false,
	}

	descendantFontsNode := pdfNode{
		key: nodeKey{
			name: "DescendantFonts",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: descendantFontsNodeItemRefID,
		},
	}

	maxRealID++
	toUnicodeRefID := objectID{
		id:     maxRealID,
		isReal: true,
	}
	toUnicodeNodeRef := pdfNode{
		key: nodeKey{
			use:  NodeKeyUseName,
			name: "ToUnicode",
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: toUnicodeRefID,
		},
	}

	ssfNodes.append(typeNode)
	ssfNodes.append(subtypeNode)
	ssfNodes.append(baseFontNode)
	ssfNodes.append(encodingNode)
	ssfNodes.append(descendantFontsNode)
	ssfNodes.append(toUnicodeNodeRef)

	//tounicode

	maxRealID, maxFakeID, err := p.appendToUnicode(ssf, fontRef, toUnicodeRefID, maxRealID, maxFakeID)
	if err != nil {
		return ssfNodesObjectID, maxRealID, maxFakeID, errors.Wrap(err, "")
	}
	//DescendantFonts
	maxRealID++
	cidFontRefID := objectID{
		id:     maxRealID,
		isReal: true,
	}

	descendantFontsItemNodes := pdfNodes{}
	p.objects[descendantFontsNodeItemRefID] = &descendantFontsItemNodes

	descendantFontsItem0Node := pdfNode{
		key: nodeKey{
			use: NodeKeyUseIndex,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: cidFontRefID,
		},
	}

	descendantFontsItemNodes.append(descendantFontsItem0Node)

	//CID Font
	maxRealID, maxFakeID, err = p.appendCidFont(ssf, fontRef, cidFontRefID, maxRealID, maxFakeID)
	if err != nil {
		return ssfNodesObjectID, maxRealID, maxFakeID, errors.Wrap(err, "")
	}

	return ssfNodesObjectID, maxRealID, maxFakeID, nil
}

func (p *PdfData) appendCidFont(
	ssf *subsetFont,
	fontRef FontRef,
	cidFontRefID objectID,
	maxRealID uint32,
	maxFakeID uint32,
) (uint32, uint32, error) {
	cidFontNodes := pdfNodes{}
	p.objects[cidFontRefID] = &cidFontNodes

	cidtypeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/Font",
		},
	}

	cidSubtypeNode := pdfNode{
		key: nodeKey{
			name: "Subtype",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/CIDFontType2",
		},
	}

	baseFontNode := pdfNode{
		key: nodeKey{
			name: "BaseFont",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	maxFakeID++
	cidSystemInfoNodeRefID := objectID{
		id:     maxFakeID,
		isReal: false,
	}

	cidSystemInfoNode := pdfNode{
		key: nodeKey{
			name: "CIDSystemInfo",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: cidSystemInfoNodeRefID,
		},
	}

	maxFakeID++
	wRefID := objectID{
		id:     maxFakeID,
		isReal: false,
	}

	wNode := pdfNode{
		key: nodeKey{
			name: "W",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: wRefID,
		},
	}

	maxRealID++
	fontDescriptorRefID := objectID{
		id:     maxRealID,
		isReal: true,
	}

	fontDescriptorNode := pdfNode{
		key: nodeKey{
			name: "FontDescriptor",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: fontDescriptorRefID,
		},
	}

	cidFontNodes.append(cidtypeNode)
	cidFontNodes.append(cidSubtypeNode)
	cidFontNodes.append(cidSystemInfoNode)
	cidFontNodes.append(wNode)
	cidFontNodes.append(fontDescriptorNode)
	cidFontNodes.append(baseFontNode)

	//fontDescriptor
	maxRealID, maxFakeID, err := p.appendFontDescriptor(ssf, fontRef, fontDescriptorRefID, maxRealID, maxFakeID)
	if err != nil {
		return maxRealID, maxFakeID, errors.Wrap(err, "")
	}

	//w
	wNodes := pdfNodes{}
	p.objects[wRefID] = &wNodes

	for _, glyphIndex := range ssf.glyphIndexs {

		width := ssf.glyphIndexToPdfWidth(glyphIndex)

		wItemNode := pdfNode{
			key: nodeKey{
				use: NodeKeyUseIndex,
			},
			content: nodeContent{
				use: NodeContentUseString,
				str: fmt.Sprintf("%d[%d]", glyphIndex, width),
			},
		}
		wNodes.append(wItemNode)
	}

	//CID SystemInfo
	cidSystemInfoNodes := pdfNodes{}
	p.objects[cidSystemInfoNodeRefID] = &cidSystemInfoNodes
	orderingNode := pdfNode{
		key: nodeKey{
			name: "Ordering",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "(Identity)",
		},
	}

	registryNode := pdfNode{
		key: nodeKey{
			name: "Registry",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "(Adobe)",
		},
	}

	supplementNode := pdfNode{
		key: nodeKey{
			name: "Supplement",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "0",
		},
	}

	cidSystemInfoNodes.append(orderingNode)
	cidSystemInfoNodes.append(registryNode)
	cidSystemInfoNodes.append(supplementNode)
	return maxRealID, maxFakeID, nil
}

func (p *PdfData) appendFontDescriptor(
	ssf *subsetFont,
	fontRef FontRef,
	fontDescriptorRefID objectID,
	maxRealID uint32,
	maxFakeID uint32,
) (uint32, uint32, error) {

	fontDescriptorNodes := pdfNodes{}
	p.objects[fontDescriptorRefID] = &fontDescriptorNodes

	typeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/FontDescriptor",
		},
	}

	ascentNode := pdfNode{
		key: nodeKey{
			name: "Ascent",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.Ascender(), ssf.ttfp.UnitsPerEm())),
		},
	}

	capHeightNode := pdfNode{
		key: nodeKey{
			name: "CapHeight",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.CapHeight(), ssf.ttfp.UnitsPerEm())),
		},
	}

	flagsNode := pdfNode{
		key: nodeKey{
			name: "Flags",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", ssf.ttfp.Flag()),
		},
	}

	maxFakeID++
	fontBoxNodeItemRefID := objectID{
		id:     maxFakeID,
		isReal: false,
	}
	fontBoxNode := pdfNode{
		key: nodeKey{
			name: "FontBBox",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: fontBoxNodeItemRefID,
		},
	}

	fontNameNode := pdfNode{
		key: nodeKey{
			name: "FontName",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	italicAngleNode := pdfNode{
		key: nodeKey{
			name: "ItalicAngle",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", ssf.ttfp.ItalicAngle()),
		},
	}

	stemVNode := pdfNode{
		key: nodeKey{
			name: "ItalicAngle",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "0",
		},
	}

	xHeightNode := pdfNode{
		key: nodeKey{
			name: "XHeight",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XHeight(), ssf.ttfp.UnitsPerEm())),
		},
	}

	maxRealID++
	fontFile2RefID := objectID{
		id:     maxRealID,
		isReal: true,
	}

	fontFile2RefNode := pdfNode{
		key: nodeKey{
			name: "FontFile2",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use:   NodeContentUseRefTo,
			refTo: fontFile2RefID,
		},
	}

	maxRealID, maxFakeID, err := p.appendFontFile2(ssf, fontRef, fontFile2RefID, maxRealID, maxFakeID) //fontfile2
	if err != nil {
		return maxRealID, maxFakeID, errors.Wrap(err, "")
	}

	fontDescriptorNodes.append(typeNode)
	fontDescriptorNodes.append(ascentNode)
	fontDescriptorNodes.append(capHeightNode)
	fontDescriptorNodes.append(flagsNode)
	fontDescriptorNodes.append(fontBoxNode)
	fontDescriptorNodes.append(fontNameNode)
	fontDescriptorNodes.append(italicAngleNode)
	fontDescriptorNodes.append(stemVNode)
	fontDescriptorNodes.append(xHeightNode)
	fontDescriptorNodes.append(fontFile2RefNode)

	//fontbox
	fontBoxNodeItemNodes := pdfNodes{}
	p.objects[fontBoxNodeItemRefID] = &fontBoxNodeItemNodes

	fontBoxItemXMinNode := pdfNode{
		key: nodeKey{
			use: NodeKeyUseIndex,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XMin(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemYMinNode := pdfNode{
		key: nodeKey{
			use: NodeKeyUseIndex,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.YMin(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemXMaxNode := pdfNode{
		key: nodeKey{
			use: NodeKeyUseIndex,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XMax(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemYMaxNode := pdfNode{
		key: nodeKey{
			use: NodeKeyUseIndex,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.YMax(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxNodeItemNodes.append(fontBoxItemXMinNode)
	fontBoxNodeItemNodes.append(fontBoxItemYMinNode)
	fontBoxNodeItemNodes.append(fontBoxItemXMaxNode)
	fontBoxNodeItemNodes.append(fontBoxItemYMaxNode)

	return maxRealID, maxFakeID, nil
}

//convert unit
func toPdfUnit(val int, unitsPerEm uint) int {
	return core.Round(float64(float64(val) * 1000.00 / float64(unitsPerEm)))
}

func (p PdfData) fontName(f FontRef) string {
	return string(f)
}
