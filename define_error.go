package nxpdf

import "errors"

//ErrFontRefNotFound FontRef not found
var ErrFontRefNotFound = errors.New("FontRef not found")

//ErrGlyphNotFound glyph not found
var ErrGlyphNotFound = errors.New("glyph not found")

//ErrRuneNotFound rune not found
var ErrRuneNotFound = errors.New("rune not found")

//ErrDictNotFound dict not found
var ErrDictNotFound = errors.New("dict not found")

//ErrObjectIDNotFound object id not found
var ErrObjectIDNotFound = errors.New("object id not found")

//ErrKeyNameNotFound keyname not found
var ErrKeyNameNotFound = errors.New("keyname not found")
