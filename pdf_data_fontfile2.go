package nxpdf

import (
	"bytes"
	"compress/zlib"
	"math/big"
	"sort"
	"strconv"

	"github.com/oneplus1000/nxpdf/font"
	"github.com/pkg/errors"
	"github.com/signintech/gopdf/fontmaker/core"
)

//EntrySelectors entry selectors
var EntrySelectors = []int{
	0, 0, 1, 1, 2, 2,
	2, 2, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 4,
	4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4,
}

func (p *PdfData) appendFontFile2(
	ssf *subsetFont,
	fontRef FontRef,
	fontFile2RefID objectID,
	maxRealID uint32,
	maxFakeID uint32,
) (uint32, uint32, error) {

	b, err := p.makeFont(ssf)
	if err != nil {
		return maxRealID, maxFakeID, errors.Wrap(err, "makeFont fail")
	}

	var zbuff bytes.Buffer
	gzipwriter := zlib.NewWriter(&zbuff)
	_, err = gzipwriter.Write(b)
	if err != nil {
		return maxRealID, maxFakeID, errors.Wrap(err, "gzipwriter.Write(...) fail")
	}
	gzipwriter.Close()

	fontFile2Nodes := pdfNodes{}
	p.objects[fontFile2RefID] = &fontFile2Nodes

	lengthNode := pdfNode{
		key: nodeKey{
			name: "Length",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: strconv.Itoa(zbuff.Len()),
		},
	}

	filterNode := pdfNode{
		key: nodeKey{
			name: "Filter",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: "/FlateDecode",
		},
	}

	length1Node := pdfNode{
		key: nodeKey{
			name: "Length1",
			use:  NodeKeyUseName,
		},
		content: nodeContent{
			use: NodeContentUseString,
			str: strconv.Itoa(len(b)),
		},
	}

	streamNode := pdfNode{
		key: nodeKey{
			use: NodeKeyUseStream,
		},
		content: nodeContent{
			use:    NodeContentUseStream,
			stream: zbuff.Bytes(),
		},
	}

	fontFile2Nodes.append(lengthNode)
	fontFile2Nodes.append(filterNode)
	fontFile2Nodes.append(length1Node)
	fontFile2Nodes.append(streamNode)

	return maxRealID, maxFakeID, nil
}

func (p *PdfData) makeFont(ssf *subsetFont) ([]byte, error) {
	var buff Buff
	ttfp := &ssf.ttfp
	tables := make(map[string]font.TableDirectoryEntry)
	tables["cvt "] = ttfp.GetTables()["cvt "] //มีช่องว่างด้วยนะ (space is important)
	tables["fpgm"] = ttfp.GetTables()["fpgm"]
	tables["glyf"] = ttfp.GetTables()["glyf"]
	tables["head"] = ttfp.GetTables()["head"]
	tables["hhea"] = ttfp.GetTables()["hhea"]
	tables["hmtx"] = ttfp.GetTables()["hmtx"]
	tables["loca"] = ttfp.GetTables()["loca"]
	tables["maxp"] = ttfp.GetTables()["maxp"]
	tables["prep"] = ttfp.GetTables()["prep"]
	tableCount := len(tables)
	selector := EntrySelectors[tableCount]

	glyphTable, locaTable, err := p.makeGlyfAndLocaTable(ssf)
	if err != nil {
		return nil, err
	}

	WriteUInt32(&buff, 0x00010000)
	WriteUInt16(&buff, uint(tableCount))
	WriteUInt16(&buff, ((1 << uint(selector)) * 16))
	WriteUInt16(&buff, uint(selector))
	WriteUInt16(&buff, (uint(tableCount)-(1<<uint(selector)))*16)

	var tags []string
	for tag := range tables {
		tags = append(tags, tag) //copy all tag
	}
	sort.Strings(tags) //order
	idx := 0
	tablePosition := int(12 + 16*tableCount)
	for idx < tableCount {
		entry := tables[tags[idx]]
		//write data
		offset := uint64(tablePosition)
		buff.SetPosition(tablePosition)
		if tags[idx] == "glyf" {
			entry.Length = uint(len(glyphTable))
			entry.CheckSum = CheckSum(glyphTable)
			WriteBytes(&buff, glyphTable, 0, entry.PaddedLength())
		} else if tags[idx] == "loca" {
			if ttfp.IsShortIndex {
				entry.Length = uint(len(locaTable) * 2)
			} else {
				entry.Length = uint(len(locaTable) * 4)
			}

			data := make([]byte, entry.PaddedLength())
			length := len(locaTable)
			byteIdx := 0
			if ttfp.IsShortIndex {
				for j := 0; j < length; j++ {
					val := locaTable[j] / 2
					data[byteIdx] = byte(val >> 8)
					byteIdx++
					data[byteIdx] = byte(val)
					byteIdx++
				}
			} else {
				for j := 0; j < length; j++ {
					val := locaTable[j]
					data[byteIdx] = byte(val >> 24)
					byteIdx++
					data[byteIdx] = byte(val >> 16)
					byteIdx++
					data[byteIdx] = byte(val >> 8)
					byteIdx++
					data[byteIdx] = byte(val)
					byteIdx++
				}
			}
			entry.CheckSum = CheckSum(data)
			WriteBytes(&buff, data, 0, len(data))
		} else {
			WriteBytes(&buff, ttfp.FontData(), int(entry.Offset), entry.PaddedLength())
		}
		endPosition := buff.Position()
		//tablePosition = endPosition

		//write table
		buff.SetPosition(idx*16 + 12)
		WriteTag(&buff, tags[idx])
		WriteUInt32(&buff, uint(entry.CheckSum))
		WriteUInt32(&buff, uint(offset)) //offset
		WriteUInt32(&buff, uint(entry.Length))

		tablePosition = endPosition
		idx++
	}
	//DebugSubType(buff.Bytes())
	//me.buffer.Write(buff.Bytes())
	return buff.Bytes(), nil
}

func (p *PdfData) makeGlyfAndLocaTable(ssf *subsetFont) ([]byte, []int, error) {

	ttfp := &ssf.ttfp
	var glyf core.TableDirectoryEntry

	numGlyphs := int(ttfp.NumGlyphs())

	_, glyphArray := p.completeGlyphClosure(ssf, ssf.glyphIndexs)
	glyphCount := len(glyphArray)
	sort.Ints(glyphArray)

	size := 0
	for idx := 0; idx < glyphCount; idx++ {
		size += p.getGlyphSize(ssf, glyphArray[idx])
	}
	glyf.Length = uint(size)

	glyphTable := make([]byte, glyf.PaddedLength())
	locaTable := make([]int, numGlyphs+1)

	glyphOffset := 0
	glyphIndex := 0
	for idx := 0; idx < numGlyphs; idx++ {
		locaTable[idx] = glyphOffset
		if glyphIndex < glyphCount && glyphArray[glyphIndex] == idx {
			glyphIndex++
			bytes := p.getGlyphData(ssf, idx)
			length := len(bytes)
			if length > 0 {
				for i := 0; i < length; i++ {
					glyphTable[glyphOffset+i] = bytes[i]
				}
				glyphOffset += length
			}
		}
	} //end for
	locaTable[numGlyphs] = glyphOffset
	return glyphTable, locaTable, nil
}

func (p *PdfData) getGlyphSize(ssf *subsetFont, glyph int) int {
	ttfp := &ssf.ttfp
	glyf := ttfp.GetTables()["glyf"]
	start := int(glyf.Offset + ttfp.LocaTable[glyph])
	next := int(glyf.Offset + ttfp.LocaTable[glyph+1])
	return next - start
}

func (p *PdfData) getGlyphData(ssf *subsetFont, glyph int) []byte {
	ttfp := &ssf.ttfp
	glyf := ttfp.GetTables()["glyf"]
	start := int(glyf.Offset + ttfp.LocaTable[glyph])
	next := int(glyf.Offset + ttfp.LocaTable[glyph+1])
	count := next - start
	var data []byte
	i := 0
	for i < count {
		data = append(data, ttfp.FontData()[start+i])
		i++
	}
	return data
}

func (p *PdfData) completeGlyphClosure(ssf *subsetFont, glyphs map[rune]uint) (map[rune]uint, []int) {
	var glyphArray []int
	//copy
	isContainZero := false
	for _, v := range glyphs {
		glyphArray = append(glyphArray, int(v))
		if v == 0 {
			isContainZero = true
		}
	}
	if !isContainZero {
		glyphArray = append(glyphArray, 0)
	}

	i := 0
	count := len(glyphs)
	for i < count {
		p.AddCompositeGlyphs(ssf, &glyphArray, glyphArray[i])
		i++
	}
	return glyphs, glyphArray
}

const weHaveAScale = 8
const moreComponents = 32
const arg1And2AreWords = 1
const weHaveAnXAndYScale = 64
const weHaveATwoByTwo = 128

//AddCompositeGlyphs add composite glyph
//composite glyph is a Unicode entity that can be defined as a sequence of one or more other characters.
func (p *PdfData) AddCompositeGlyphs(ssf *subsetFont, glyphArray *[]int, glyph int) {
	start := p.GetOffset(ssf, int(glyph))
	if start == p.GetOffset(ssf, int(glyph)+1) {
		return
	}

	offset := start
	ttfp := &ssf.ttfp
	fontData := ttfp.FontData()
	numContours, step := ReadShortFromByte(fontData, offset)
	offset += step
	if numContours >= 0 {
		return
	}

	offset += 8
	for {
		flags, step1 := ReadUShortFromByte(fontData, offset)
		offset += step1
		cGlyph, step2 := ReadUShortFromByte(fontData, offset)
		offset += step2
		//check cGlyph is contain in glyphArray?
		glyphContainsKey := false
		for _, g := range *glyphArray {
			if g == int(cGlyph) {
				glyphContainsKey = true
				break
			}
		}
		if !glyphContainsKey {
			*glyphArray = append(*glyphArray, int(cGlyph))
		}

		if (flags & moreComponents) == 0 {
			return
		}
		offsetAppend := 4
		if (flags & arg1And2AreWords) == 0 {
			offsetAppend = 2
		}
		if (flags & weHaveAScale) != 0 {
			offsetAppend += 2
		} else if (flags & weHaveAnXAndYScale) != 0 {
			offsetAppend += 4
		}
		if (flags & weHaveATwoByTwo) != 0 {
			offsetAppend += 8
		}
		offset += offsetAppend
	}
}

//GetOffset get offset from glyf table
func (p *PdfData) GetOffset(ssf *subsetFont, glyph int) int {
	ttfp := &ssf.ttfp
	glyf := ttfp.GetTables()["glyf"]
	offset := int(glyf.Offset + ttfp.LocaTable[glyph])
	return offset
}

//ReadShortFromByte read short from byte array
func ReadShortFromByte(data []byte, offset int) (int64, int) {
	buff := data[offset : offset+2]
	num := big.NewInt(0)
	num.SetBytes(buff)
	u := num.Uint64()
	var v int64
	if u >= 0x8000 {
		v = int64(u) - 65536
	} else {
		v = int64(u)
	}
	return v, 2
}

//ReadUShortFromByte read ushort from byte array
func ReadUShortFromByte(data []byte, offset int) (uint64, int) {
	buff := data[offset : offset+2]
	num := big.NewInt(0)
	num.SetBytes(buff)
	return num.Uint64(), 2
}

//CheckSum check sum
func CheckSum(data []byte) uint {

	var byte3, byte2, byte1, byte0 uint64
	byte3 = 0
	byte2 = 0
	byte1 = 0
	byte0 = 0
	length := len(data)
	i := 0
	for i < length {
		byte3 += uint64(data[i])
		i++
		byte2 += uint64(data[i])
		i++
		byte1 += uint64(data[i])
		i++
		byte0 += uint64(data[i])
		i++
	}
	//var result uint32
	result := uint32(byte3<<24) + uint32(byte2<<16) + uint32(byte1<<8) + uint32(byte0)
	return uint(result)
}
