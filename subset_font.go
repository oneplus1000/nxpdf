package nxpdf

type subsetFont struct {
	fontfileRaw []byte
	glyphIndexs map[rune]int
}

func (s *subsetFont) addChars(text string) error {
	return nil
}
