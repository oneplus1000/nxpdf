package nxpdf

import (
	"fmt"

	"github.com/pkg/errors"
)

func (p *PdfData) appendSubsetFont(ssf *subsetFont, fontRef FontRef, maxRealID uint32, maxFakeID uint32) (uint32, uint32, error) {

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
			str: "/" + string(fontRef),
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

	ssfNodes.append(typeNode)
	ssfNodes.append(subtypeNode)
	ssfNodes.append(baseFontNode)
	ssfNodes.append(encodingNode)
	ssfNodes.append(descendantFontsNode)

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
	maxRealID, maxFakeID, err := p.appendCidFont(ssf, fontRef, cidFontRefID, maxRealID, maxFakeID)
	if err != nil {
		return maxRealID, maxFakeID, errors.Wrap(err, "")
	}

	return maxRealID, maxFakeID, nil
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

	//fontDescriptor
	p.appendFontDescriptor(ssf, fontRef, fontDescriptorRefID, maxRealID, maxFakeID)

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

	fontDescriptorNodes.append(typeNode)

	return maxRealID, maxFakeID, nil
}
