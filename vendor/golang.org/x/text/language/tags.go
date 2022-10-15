// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

// TODO: Various sets of commonly use tags and regions.

// MustParse is like Parse, but panics if the given BCP 47 tag cannot be parsed.
// It simplifies safe initialization of Tag values.
func MustParse(s string) Tag {
	t, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParse is like Parse, but panics if the given BCP 47 tag cannot be parsed.
// It simplifies safe initialization of Tag values.
func (c CanonType) MustParse(s string) Tag {
	t, err := c.Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParseBase is like ParseBase, but panics if the given base cannot be parsed.
// It simplifies safe initialization of Base values.
func MustParseBase(s string) Base {
	b, err := ParseBase(s)
	if err != nil {
		panic(err)
	}
	return b
}

// MustParseScript is like ParseScript, but panics if the given script cannot be
// parsed. It simplifies safe initialization of Script values.
func MustParseScript(s string) Script {
	scr, err := ParseScript(s)
	if err != nil {
		panic(err)
	}
	return scr
}

// MustParseRegion is like ParseRegion, but panics if the given region cannot be
// parsed. It simplifies safe initialization of Region values.
func MustParseRegion(s string) Region {
	r, err := ParseRegion(s)
	if err != nil {
		panic(err)
	}
	return r
}

var (
	und = Tag{}

	Und Tag = Tag{}

	Afrikaans            Tag = Tag{lang: _af}                //  af
	Amharic              Tag = Tag{lang: _am}                //  am
	Arabic               Tag = Tag{lang: _ar}                //  ar
	ModernStandardArabic Tag = Tag{lang: _ar, region: _001}  //  ar-001
	Azerbaijani          Tag = Tag{lang: _az}                //  az
	Bulgarian            Tag = Tag{lang: _bg}                //  bg
	Bengali              Tag = Tag{lang: _bn}                //  bn
	Catalan              Tag = Tag{lang: _ca}                //  ca
	Czech                Tag = Tag{lang: _cs}                //  cs
	Danish               Tag = Tag{lang: _da}                //  da
	German               Tag = Tag{lang: _de}                //  de
	Greek                Tag = Tag{lang: _el}                //  el
	English              Tag = Tag{lang: _en}                //  en
	AmericanEnglish      Tag = Tag{lang: _en, region: _US}   //  en-US
	BritishEnglish       Tag = Tag{lang: _en, region: _GB}   //  en-GB
	Spanish              Tag = Tag{lang: _es}                //  es
	EuropeanSpanish      Tag = Tag{lang: _es, region: _ES}   //  es-ES
	LatinAmericanSpanish Tag = Tag{lang: _es, region: _419}  //  es-419
	Estonian             Tag = Tag{lang: _et}                //  et
	Persian              Tag = Tag{lang: _fa}                //  fa
	Finnish              Tag = Tag{lang: _fi}                //  fi
	Filipino             Tag = Tag{lang: _fil}               //  fil
	French               Tag = Tag{lang: _fr}                //  fr
	CanadianFrench       Tag = Tag{lang: _fr, region: _CA}   //  fr-CA
	Gujarati             Tag = Tag{lang: _gu}                //  gu
	Hebrew     