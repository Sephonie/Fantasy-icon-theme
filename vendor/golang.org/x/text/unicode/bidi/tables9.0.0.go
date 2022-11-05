// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

// +build !go1.10

package bidi

// UnicodeVersion is the Unicode version from which the tables in this package are derived.
const UnicodeVersion = "9.0.0"

// xorMasks contains masks to be xor-ed with brackets to get the reverse
// version.
var xorMasks = []int32{ // 8 elements
	0, 1, 6, 7, 3, 15, 29, 63,
} // Size: 56 bytes

// lookup returns the trie value for the first UTF-8 encoding in s and
// the width in bytes of this encoding. The size will be 0 if s does not
// hold enough bytes to complete the encoding. len(s) must be greater than 0.
func (t *bidiTrie) lookup(s []byte) (v uint8, sz int) {
	c0 := s[0]
	switch {
	case c0 < 0x80: // is ASCII
		return bidiValues[c0], 1
	case c0 < 0xC2:
		return 0, 1 // Illegal UTF-8: not a starter, not ASCII.
	case c0 < 0xE0: // 2-byte UTF-8
		if len(s) < 2 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c1), 2
	case c0 < 0xF0: // 3-byte UTF-8
		if len(s) < 3 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = bidiIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c2), 3
	case c0 < 0xF8: // 4-byte UTF-8
		if len(s) < 4 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = bidiIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		o = uint32(i)<<6 + uint32(c2)
		i = bidiIndex[o]
		c3 := s[3]
		if c3 < 0x80 || 0xC0 <= c3 {
			return 0, 3 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c3), 4
	}
	// Illegal rune
	return 0, 1
}

// lookupUnsafe returns the trie value for the first UTF-8 encoding in s.
// s must start with a full and valid UTF-8 encoded rune.
func (t *bidiTrie) lookupUnsafe(s []byte) uint8 {
	c0 := s[0]
	if c0 < 0x80 { // is ASCII
		return bidiValues[c0]
	}
	i := bidiIndex[c0]
	if c0 < 0xE0 { // 2-byte UTF-8
		return t.lookupValue(uint32(i), s[1])
	}
	i = bidiIndex[uint32(i)<<6+uint32(s[1])]
	if c0 < 0xF0 { // 3-byte UTF-8
		return t.lookupValue(uint32(i), s[2])
	}
	i = bidiIndex[uint32(i)<<6+uint32(s[2])]
	if c0 < 0xF8 { // 4-byte UTF-8
		return t.lookupValue(uint32(i), s[3])
	}
	return 0
}

// lookupString returns the trie value for the first UTF-8 encoding in s and
// the width in bytes of this encoding. The size will be 0 if s does not
// hold enough bytes to complete the encoding. len(s) must be greater than 0.
func (t *bidiTrie) lookupString(s string) (v uint8, sz int) {
	c0 := s[0]
	switch {
	case c0 < 0x80: // is ASCII
		return bidiValues[c0], 1
	case c0 < 0xC2:
		return 0, 1 // Illegal UTF-8: not a starter, not ASCII.
	case c0 < 0xE0: // 2-byte UTF-8
		if len(s) < 2 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c1), 2
	case c0 < 0xF0: // 3-byte UTF-8
		if len(s) < 3 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = bidiIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c2), 3
	case c0 < 0xF8: // 4-byte UTF-8
		if len(s) < 4 {
			return 0, 0
		}
		i := bidiIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = bidiIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		o = uint32(i)<<6 + uint32(c2)
		i = bidiIndex[o]
		c3 := s[3]
		if c3 < 0x80 || 0xC0 <= c3 {
			return 0, 3 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c3), 4
	}
	// Illegal rune
	return 0, 1
}

// lookupStringUnsafe returns the trie value for the first UTF-8 encoding in s.
// s must start with a full and valid UTF-8 encoded rune.
func (t *bidiTrie) lookupStringUnsafe(s string) uint8 {
	c0 := s[0]
	if c0 < 0x80 { // is ASCII
		return bidiValues[c0]
	}
	i := bidiIndex[c0]
	if c0 < 0xE0 { // 2-byte UTF-8
		return t.lookupValue(uint32(i), s[1])
	}
	i = bidiIndex[uint32(i)<<6+uint32(s[1])]
	if c0 < 0xF0 { // 3-byte UTF-8
		return t.lookupValue(uint32(i), s[2])
	}
	i = bidiIndex[uint32(i)<<6+uint32(s[2])]
	if c0 < 0xF8 { // 4-byte UTF-8
		return t.lookupValue(uint32(i), s[3])
	}
	return 0
}

// bidiTrie. Total size: 15744 bytes (15.38 KiB). Checksum: b4c3b70954803b86.
type bidiTrie struct{}

func newBidiTrie(i int) *bidiTrie {
	return &bidiTrie{}
}

// lookupValue determines the type of block n and looks up the value for b.
func (t *bidiTrie) lookupValue(n uint32, b byte) uint8 {
	switch {
	default:
		return uint8(bidiValues[n<<6+uint32(b)])
	}
}

// bidiValues: 222 blocks, 14208 entries, 14208 bytes
// The third block is the zero block.
var bidiValues = [14208]uint8{
	// Block 0x0, offset 0x0
	0x00: 0x000b, 0x01: 0x000b, 0x02: 0x000b, 0x03: 0x000b, 0x04: 0x000b, 0x05: 0x000b,
	0x06: 0x000b, 0x07: 0x000b, 0x08: 0x000b, 0x09: 0x0008, 0x0a: 0x0007, 0x0b: 0x0008,
	0x0c: 0x0009, 0x0d: 0x0007, 0x0e: 0x000b, 0x0f: 0x000b, 0x10: 0x000b, 0x11: 0x000b,
	0x12: 0x000b, 0x13: 0x000b, 0x14: 0x000b, 0x15: 0x000b, 0x16: 0x000b, 0x17: 0x000b,
	0x18: 0x000b, 0x19: 0x000b, 0x1a: 0x000b, 0x1b: 0x000b, 0x1c: 0x0007, 0x1d: 0x0007,
	0x1e: 0x0007, 0x1f: 0x0008, 0x20: 0x0009, 0x21: 0x000a, 0x22: 0x000a, 0x23: 0x0004,
	0x24: 0x0004, 0x25: 0x0004, 0x26: 0x000a, 0x27: 0x000a, 0x28: 0x003a, 0x29: 0x002a,
	0x2a: 0x000a, 0x2b: 0x0003, 0x2c: 0x0006, 0x2d: 0x0003, 0x2e: 0x0006, 0x2f: 0x0006,
	0x30: 0x0002, 0x31: 0x0002, 0x32: 0x0002, 0x33: 0x0002, 0x34: 0x0002, 0x35: 0x0002,
	0x36: 0x0002, 0x37: 0x0002, 0x38: 0x0002, 0x39: 0x0002, 0x3a: 0x0006, 0x3b: 0x000a,
	0x3c: 0x000a, 0x3d: 0x000a, 0x3e: 0x000a, 0x3f: 0x000a,
	// Block 0x1, offset 0x40
	0x40: 0x000a,
	0x5b: 0x005a, 0x5c: 0x000a, 0x5d: 0x004a,
	0x5e: 0x000a, 0x5f: 0x000a, 0x60: 0x000a,
	0x7b: 0x005a,
	0x7c: 0x000a, 0x7d: 0x004a, 0x7e: 0x000a, 0x7f: 0x000b,
	// Block 0x2, offset 0x80
	// Block 0x3, offset 0xc0
	0xc0: 0x000b, 0xc1: 0x000b, 0xc2: 0x000b, 0xc3: 0x000b, 0xc4: 0x000b, 0xc5: 0x0007,
	0xc6: 0x000b, 0xc7: 0x000b, 0xc8: 0x000b, 0xc9: 0x000b, 0xca: 0x000b, 0xcb: 0x000b,
	0xcc: 0x000b, 0xcd: 0x000b, 0xce: 0x000b, 0xcf: 0x000b, 0xd0: 0x000b, 0xd1: 0x000b,
	0xd2: 0x000b, 0xd3: 0x000b, 0xd4: 0x000b, 0xd5: 0x000b, 0xd6: 0x000b, 0xd7: 0x000b,
	0xd8: 0x000b, 0xd9: 0x000b, 0xda: 0x000b, 0xdb: 0x000b, 0xdc: 0x000b, 0xdd: 0x000b,
	0xde: 0x000b, 0xdf: 0x000b, 0xe0: 0x0006, 0xe1: 0x000a, 0xe2: 0x0004, 0xe3: 0x0004,
	0xe4: 0x0004, 0xe5: 0x0004, 0xe6: 0x000a, 0xe7: 0x000a, 0xe8: 0x000a, 0xe9: 0x000a,
	0xeb: 0x000a, 0xec: 0x000a, 0xed: 0x000b, 0xee: 0x000a, 0xef: 0x000a,
	0xf0: 0x0004, 0xf1: 0x0004, 0xf2: 0x0002, 0xf3: 0x0002, 0xf4: 0x000a,
	0xf6: 0x000a, 0xf7: 0x000a, 0xf8: 0x000a, 0xf9: 0x0002, 0xfb: 0x000a,
	0xfc: 0x000a, 0xfd: 0x000a, 0xfe: 0x000a, 0xff: 0x000a,
	// Block 0x4, offset 0x100
	0x117: 0x000a,
	0x137: 0x000a,
	// Block 0x5, offset 0x140
	0x179: 0x000a, 0x17a: 0x000a,
	// Block 0x6, offset 0x180
	0x182: 0x000a, 0x183: 0x000a, 0x184: 0x000a, 0x185: 0x000a,
	0x186: 0x000a, 0x187: 0x000a, 0x188: 0x000a, 0x189: 0x000a, 0x18a: 0x000a, 0x18b: 0x000a,
	0x18c: 0x000a, 0x18d: 0x000a, 0x18e: 0x000a, 0x18f: 0x000a,
	0x192: 0x000a, 0x193: 0x000a, 0x194: 0x000a, 0x195: 0x000a, 0x196: 0x000a, 0x197: 0x000a,
	0x198: 0x000a, 0x199: 0x000a, 0x19a: 0x000a, 0x19b: 0x000a, 0x19c: 0x000a, 0x19d: 0x000a,
	0x19e: 0x000a, 0x19f: 0x000a,
	0x1a5: 0x000a, 0x1a6: 0x000a, 0x1a7: 0x000a, 0x1a8: 0x000a, 0x1a9: 0x000a,
	0x1aa: 0x000a, 0x1ab: 0x000a, 0x1ac: 0x000a, 0x1ad: 0x000a, 0x1af: 0x000a,
	0x1b0: 0x000a, 0x1b1: 0x000a, 0x1b2: 0x000a, 0x1b3: 0x000a, 0x1b4: 0x000a, 0x1b5: 0x000a,
	0x1b6: 0x000a, 0x1b7: 0x000a, 0x1b8: 0x000a, 0x1b9: 0x000a, 0x1ba: 0x000a, 0x1bb: 0x000a,
	0x1bc: 0x000a, 0x1bd: 0x000a, 0x1be: 0x000a, 0x1bf: 0x000a,
	// Block 0x7, offset 0x1c0
	0x1c0: 0x000c, 0x1c1: 0x000c, 0x1c2: 0x000c, 0x1c3: 0x000c, 0x1c4: 0x000c, 0x1c5: 0x000c,
	0x1c6: 0x000c, 0x1c7: 0x000c, 0x1c8: 0x000c, 0x1c9: 0x000c, 0x1ca: 0x000c, 0x1cb: 0x000c,
	0x1cc: 0x000c, 0x1cd: 0x000c, 0x1ce: 0x000c, 0x1cf: 0x000c, 0x1d0: 0x000c, 0x1d1: 0x000c,
	0x1d2: 0x000c, 0x1d3: 0x000c, 0x1d4: 0x000c, 0x1d5: 0x000c, 0x1d6: 0x000c, 0x1d7: 0x000c,
	0x1d8: 0x000c, 0x1d9: 0x000c, 0x1da: 0x000c, 0x1db: 0x000c, 0x1dc: 0x000c, 0x1dd: 0x000c,
	0x1de: 0x000c, 0x1df: 0x000c, 0x1e0: 0x000c, 0x1e1: 0x000c, 0x1e2: 0x000c, 0x1e3: 0x000c,
	0x1e4: 0x000c, 0x1e5: 0x000c, 0x1e6: 0x000c, 0x1e7: 0x000c, 0x1e8: 0x000c, 0x1e9: 0x000c,
	0x1ea: 0x000c, 0x1eb: 0x000c, 0x1ec: 0x000c, 0x1ed: 0x000c, 0x1ee: 0x000c, 0x1ef: 0x000c,
	0x1f0: 0x000c, 0x1f1: 0x000c, 0x1f2: 0x000c, 0x1f3: 0x000c, 0x1f4: 0x000c, 0x1f5: 0x000c,
	0x1f6: 0x000c, 0x1f7: 0x000c, 0x1f8: 0x000c, 0x1f9: 0x000c, 0x1fa: 0x000c, 0x1fb: 0x000c,
	0x1fc: 0x000c, 0x1fd: 0x000c, 0x1fe: 0x000c, 0x1ff: 0x000c,
	// Block 0x8, offset 0x200
	0x200: 0x000c, 0x201: 0x000c, 0x202: 0x000c, 0x203: 0x000c, 0x204: 0x000c, 0x205: 0x000c,
	0x206: 0x000c, 0x207: 0x000c, 0x208: 0x000c, 0x209: 0x000c, 0x20a: 0x000c, 0x20b: 0x000c,
	0x20c: 0x000c, 0x20d: 0x000c, 0x20e: 0x000c, 0x20f: 0x000c, 0x210: 0x000c, 0x211: 0x000c,
	0x212: 0x000c, 0x213: 0x000c, 0x214: 0x000c, 0x215: 0x000c, 0x216: 0x000c, 0x217: 0x000c,
	0x218: 0x000c, 0x219: 0x000c, 0x21a: 0x000c, 0x21b: 0x000c, 0x21c: 0x000c, 0x21d: 0x000c,
	0x21e: 0x000c, 0x21f: 0x000c, 0x220: 0x000c, 0x221: 0x000c, 0x222: 0x000c, 0x223: 0x000c,
	0x224: 0x000c, 0x225: 0x000c, 0x226: 0x000c, 0x227: 0x000c, 0x228: 0x000c, 0x229: 0x000c,
	0x22a: 0x000c, 0x22b: 0x000c, 0x22c: 0x000c, 0x22d: 0x000c, 0x22e: 0x000c, 0x22f: 0x000c,
	0x234: 0x000a, 0x235: 0x000a,
	0x23e: 0x000a,
	// Block 0x9, offset 0x240
	0x244: 0x000a, 0x245: 0x000a,
	0x247: 0x000a,
	// Block 0xa, offset 0x280
	0x2b6: 0x000a,
	// Block 0xb, offset 0x2c0
	0x2c3: 0x000c, 0x2c4: 0x000c, 0x2c5: 0x000c,
	0x2c6: 0x000c, 0x2c7: 0x000c, 0x2c8: 0x000c, 0x2c9: 0x000c,
	// Block 0xc, offset 0x300
	0x30a: 0x000a,
	0x30d: 0x000a, 0x30e: 0x000a, 0x30f: 0x0004, 0x310: 0x0001, 0x311: 0x000c,
	0x312: 0x000c, 0x313: 0x000c, 0x314: 0x000c, 0x315: 0x000c, 0x316: 0x000c, 0x317: 0x000c,
	0x318: 0x000c, 0x319: 0x000c, 0x31a: 0x000c, 0x31b: 0x000c, 0x31c: 0x000c, 0x31d: 0x000c,
	0x31e: 0x000c, 0x31f: 0x000c, 0x320: 0x000c, 0x321: 0x000c, 0x322: 0x000c, 0x323: 0x000c,
	0x324: 0x000c, 0x325: 0x000c, 0x326: 0x000c, 0x327: 0x000c, 0x328: 0x000c, 0x329: 0x000c,
	0x32a: 0x000c, 0x32b: 0x000c, 0x32c: 0x000c, 0x32d: 0x000c, 0x32e: 0x000c, 0x32f: 0x000c,
	0x330: 0x000c, 0x331: 0x000c, 0x332: 0x000c, 0x333: 0x000c, 0x334: 0x000c, 0x335: 0x000c,
	0x336: 0x000c, 0x337: 0x000c, 0x338: 0x000c, 0x339: 0x000c, 0x33a: 0x000c, 0x33b: 0x000c,
	0x33c: 0x000c, 0x33d: 0x000c, 0x33e: 0x0001, 0x33f: 0x000c,
	// Block 0xd, offset 0x340
	0x340: 0x0001, 0x341: 0x000c, 0x342: 0x000c, 0x343: 0x0001, 0x344: 0x000c, 0x345: 0x000c,
	0x346: 0x0001, 0x347: 0x000c, 0x348: 0x0001, 0x349: 0x0001, 0x34a: 0x0001, 0x34b: 0x0001,
	0x34c: 0x0001, 0x34d: 0x0001, 0x34e: 0x0001, 0x34f: 0x0001, 0x350: 0x0001, 0x351: 0x0001,
	0x352: 0x0001, 0x353: 0x0001, 0x354: 0x0001, 0x355: 0x0001, 0x356: 0x0001, 0x357: 0x0001,
	0x358: 0x0001, 0x359: 0x0001, 0x35a: 0x0001, 0x35b: 0x0001, 0x35c: 0x0001, 0x35d: 0x0001,
	0x35e: 0x0001, 0x35f: 0x0001, 0x360: 0x0001, 0x361: 0x0001, 0x362: 0x0001, 0x363: 0x0001,
	0x364: 0x0001, 0x365: 0x0001, 0x366: 0x0001, 0x367: 0x0001, 0x368: 0x0001, 0x369: 0x0001,
	0x36a: 0x0001, 0x36b: 0x0001, 0x36c: 0x0001, 0x36d: 0x0001, 0x36e: 0x0001, 0x36f: 0x0001,
	0x370: 0x0001, 0x371: 0x0001, 0x372: 0x0001, 0x373: 0x0001, 0x374: 0x0001, 0x375: 0x0001,
	0x376: 0x0001, 0x377: 0x0001, 0x378: 0x0001, 0x379: 0x0001, 0x37a: 0x0001, 0x37b: 0x0001,
	0x37c: 0x0001, 0x37d: 0x0001, 0x37e: 0x0001, 0x37f: 0x0001,
	// Block 0xe, offset 0x380
	0x380: 0x0005, 0x381: 0x0005, 0x382: 0x0005, 0x383: 0x0005, 0x384: 0x0005, 0x385: 0x0005,
	0x386: 0x000a, 0x387: 0x000a, 0x388: 0x000d, 0x389: 0x0004, 0x38a: 0x0004, 0x38b: 0x000d,
	0x38c: 0x0006, 0x38d: 0x000d, 0x38e: 0x000a, 0x38f: 0x000a, 0x390: 0x000c, 0x391: 0x000c,
	0x392: 0x000c, 0x393: 0x000c, 0x394: 0x000c, 0x395: 0x000c, 0x396: 0x000c, 0x397: 0x000c,
	0x398: 0x000c, 0x399: 0x000c, 0x39a: 0x000c, 0x39b: 0x000d, 0x39c: 0x000d, 0x39d: 0x000d,
	0x39e: 0x000d, 0x39f: 0x000d, 0x3a0: 0x000d, 0x3a1: 0x000d, 0x3a2: 0x000d, 0x3a3: 0x000d,
	0x3a4: 0x000d, 0x3a5: 0x000d, 0x3a6: 0x000d, 0x3a7: 0x000d, 0x3a8: 0x000d, 0x3a9: 0x000d,
	0x3aa: 0x000d, 0x3ab: 0x000d, 0x3ac: 0x000d, 0x3ad: 0x000d, 0x3ae: 0x000d, 0x3af: 0x000d,
	0x3b0: 0x000d, 0x3b1: 0x000d, 0x3b2: 0x000d, 0x3b3: 0x000d, 0x3b4: 0x000d, 0x3b5: 0x000d,
	0x3b6: 0x000d, 0x3b7: 0x000d, 0x3b8: 0x000d, 0x3b9: 0x000d, 0x3ba: 0x000d, 0x3bb: 0x000d,
	0x3bc: 0x000d, 0x3bd: 0x000d, 0x3be: 0x000d, 0x3bf: 0x000d,
	// Block 0xf, offset 0x3c0
	0x3c0: 0x000d, 0x3c1: 0x000d, 0x3c2: 0x000d, 0x3c3: 0x000d, 0x3c4: 0x000d, 0x3c5: 0x000d,
	0x3c6: 0x000d, 0x3c7: 0x000d, 0x3c8: 0x000d, 0x3c9: 0x000d, 0x3ca: 0x000d, 0x3cb: 0x000c,
	0x3cc: 0x000c, 0x3cd: 0x000c, 0x3ce: 0x000c, 0x3cf: 0x000c, 0x3d0: 0x000c, 0x3d1: 0x000c,
	0x3d2: 0x000c, 0x3d3: 0x000c, 0x3d4: 0x000c, 0x3d5: 0x000c, 0x3d6: 0x000c, 0x3d7: 0x000c,
	0x3d8: 0x000c, 0x3d9: 0x000c, 0x3da: 0x000c, 0x3db: 0x000c, 0x3dc: 0x000c, 0x3dd: 0x000c,
	0x3de: 0x000c, 0x3df: 0x000c, 0x3e0: 0x0005, 0x3e1: 0x0005, 0x3e2: 0x0005, 0x3e3: 0x0005,
	0x3e4: 0x0005, 0x3e5: 0x0005, 0x3e6: 0x0005, 0x3e7: 0x0005, 0x3e8: 0x0005, 0x3e9: 0x0005,
	0x3ea: 0x0004, 0x3eb: 0x0005, 0x3ec: 0x0005, 0x3ed: 0x000d, 0x3ee: 0x000d, 0x3ef: 0x000d,
	0x3f0: 0x000c, 0x3f1: 0x000d, 0x3f2: 0x000d, 0x3f3: 0x000d, 0x3f4: 0x000d, 0x3f5: 0x000d,
	0x3f6: 0x000d, 0x3f7: 0x000d, 0x3f8: 0x000d, 0x3f9: 0x000d, 0x3fa: 0x000d, 0x3fb: 0x000d,
	0x3fc: 0x000d, 0x3fd: 0x000d, 0x3fe: 0x000d, 0x3ff: 0x000d,
	// Block 0x10, offset 0x400
	0x400: 0x000d, 0x401: 0x000d, 0x402: 0x000d, 0x403: 0x000d, 0x404: 0x000d, 0x405: 0x000d,
	0x406: 0x000d, 0x407: 0x000d, 0x408: 0x000d, 0x409: 0x000d, 0x40a: 0x000d, 0x40b: 0x000d,
	0x40c: 0x000d, 0x40d: 0x000d, 0x40e: 0x000d, 0x40f: 0x000d, 0x410: 0x000d, 0x411: 0x000d,
	0x412: 0x000d, 0x413: 0x000d, 0x414: 0x000d, 0x415: 0x000d, 0x416: 0x000d, 0x417: 0x000d,
	0x418: 0x000d, 0x419: 0x000d, 0x41a: 0x000d, 0x41b: 0x000d, 0x41c: 0x000d, 0x41d: 0x000d,
	0x41e: 0x000d, 0x41f: 0x000d, 0x420: 0x000d, 0x421: 0x000d, 0x422: 0x000d, 0x423: 0x000d,
	0x424: 0x000d, 0x425: 0x000d, 0x426: 0x000d, 0x427: 0x000d, 0x428: 0x000d, 0x429: 0x000d,
	0x42a: 0x000d, 0x42b: 0x000d, 0x42c: 0x000d, 0x42d: 0x000d, 0x42e: 0x000d, 0x42f: 0x000d,
	0x430: 0x000d, 0x431: 0x000d, 0x432: 0x000d, 0x433: 0x000d, 0x434: 0x000d, 0x435: 0x000d,
	0x436: 0x000d, 0x437: 0x000d, 0x438: 0x000d, 0x439: 0x000d, 0x43a: 0x000d, 0x43b: 0x000d,
	0x43c: 0x000d, 0x43d: 0x000d, 0x43e: 0x000d, 0x43f: 0x000d,
	// Block 0x11, offset 0x440
	0x440: 0x000d, 0x441: 0x000d, 0x442: 0x000d, 0x443: 0x000d, 0x444: 0x000d, 0x445: 0x000d,
	0x446: 0x000d, 0x447: 0x000d, 0x448: 0x000d, 0x449: 0x000d, 0x44a: 0x000d, 0x44b: 0x000d,
	0x44c: 0x000d, 0x44d: 0x000d, 0x44e: 0x000d, 0x44f: 0x000d, 0x450: 0x000d, 0x451: 0x000d,
	0x452: 0x000d, 0x453: 0x000d, 0x454: 0x000d, 0x455: 0x000d, 0x456: 0x000c, 0x457: 0x000c,
	0x458: 0x000c, 0x459: 0x000c, 0x45a: 0x000c, 0x45b: 0x000c, 0x45c: 0x000c, 0x45d: 0x0005,
	0x45e: 0x000a, 0x45f: 0x000c, 0x460: 0x000c, 0x461: 0x000c, 0x462: 0x000c, 0x463: 0x000c,
	0x464: 0x000c, 0x465: 0x000d, 0x466: 0x000d, 0x467: 0x000c, 0x468: 0x000c, 0x469: 0x000a,
	0x46a: 0x000c, 0x46b: 0x000c, 0x46c: 0x000c, 0x46d: 0x000c, 0x46e: 0x000d, 0x46f: 0x000d,
	0x470: 0x0002, 0x471: 0x0002, 0x472: 0x0002, 0x473: 0x0002, 0x474: 0x0002, 0x475: 0x0002,
	0x476: 0x0002, 0x477: 0x0002, 0x478: 0x0002, 0x479: 0x0002, 0x47a: 0x000d, 0x47b: 0x000d,
	0x47c: 0x000d, 0x47d: 0x000d, 0x47e: 0x000d, 0x47f: 0x000d,
	// Block 0x12, offset 0x480
	0x480: 0x000d, 0x481: 0x000d, 0x482: 0x000d, 0x483: 0x000d, 0x484: 0x000d, 0x485: 0x000d,
	0x486: 0x000d, 0x487: 0x000d, 0x488: 0x000d, 0x489: 0x000d, 0x48a: 0x000d, 0x48b: 0x000d,
	0x48c: 0x000d, 0x48d: 0x000d, 0x48e: 0x000d, 0x48f: 0x000d, 0x490: 0x000d, 0x491: 0x000c,
	0x492: 0x000d, 0x493: 0x000d, 0x494: 0x000d, 0x495: 0x000d, 0x496: 0x000d, 0x497: 0x000d,
	0x498: 0x000d, 0x499: 0x000d, 0x49a: 0x000d, 0x49b: 0x000d, 0x49c: 0x000d, 0x49d: 0x000d,
	0x49e: 0x000d, 0x49f: 0x000d, 0x4a0: 0x000d, 0x4a1: 0x000d, 0x4a2: 0x000d, 0x4a3: 0x000d,
	0x4a4: 0x000d, 0x4a5: 0x000d, 0x4a6: 0x000d, 0x4a7: 0x000d, 0x4a8: 0x000d, 0x4a9: 0x000d,
	0x4aa: 0x000d, 0x4ab: 0x000d, 0x4ac: 0x000d, 0x4ad: 0x000d, 0x4ae: 0x000d, 0x4af: 0x000d,
	0x4b0: 0x000c, 0x4b1: 0x000c, 0x4b2: 0x000c, 0x4b3: 0x000c, 0x4b4: 0x000c, 0x4b5: 0x000c,
	0x4b6: 0x000c, 0x4b7: 0x000c, 0x4b8: 0x000c, 0x4b9: 0x000c, 0x4ba: 0x000c, 0x4bb: 0x000c,
	0x4bc: 0x000c, 0x4bd: 0x000c, 0x4be: 0x000c, 0x4bf: 0x000c,
	// Block 0x13, offset 0x4c0
	0x4c0: 0x000c, 0x4c1: 0x000c, 0x4c2: 0x000c, 0x4c3: 0x000c, 0x4c4: 0x000c, 0x4c5: 0x000c,
	0x4c6: 0x000c, 0x4c7: 0x000c, 0x4c8: 0x000c, 0x4c9: 0x000c, 0x4ca: 0x000c, 0x4cb: 0x000d,
	0x4cc: 0x000d, 0x4cd: 0x000d, 0x4ce: 0x000d, 0x4cf: 0x000d, 0x4d0: 0x000d, 0x4d1: 0x000d,
	0x4d2: 0x000d, 0x4d3: 0x000d, 0x4d4: 0x000d, 0x4d5: 0x000d, 0x4d6: 0x000d, 0x4d7: 0x000d,
	0x4d8: 0x000d, 0x4d9: 0x000d, 0x4da: 0x000d, 0x4db: 0x000d, 0x4dc: 0x000d, 0x4dd: 0x000d,
	0x4de: 0x000d, 0x4df: 0x000d, 0x4e0: 0x000d, 0x4e1: 0x000d, 0x4e2: 0x000d, 0x4e3: 0x000d,
	0x4e4: 0x000d, 0x4e5: 0x000d, 0x4e6: 0x000d, 0x4e7: 0x000d, 0x4e8: 0x000d, 0x4e9: 0x000d,
	0x4ea: 0x000d, 0x4eb: 0x000d, 0x4ec: 0x000d, 0x4ed: 0x000d, 0x4ee: 0x000d, 0x4ef: 0x000d,
	0x4f0: 0x000d, 0x4f1: 0x000d, 0x4f2: 0x000d, 0x4f3: 0x000d, 0x4f4: 0x000d, 0x4f5: 0x000d,
	0x4f6: 0x000d, 0x4f7: 0x000d, 0x4f8: 0x000d, 0x4f9: 0x000d, 0x4fa: 0x000d, 0x4fb: 0x000d,
	0x4fc: 0x000d, 0x4fd: 0x000d, 0x4fe: 0x000d, 0x4ff: 0x000d,
	// Block 0x14, offset 0x500
	0x500: 0x000d, 0x501: 0x000d, 0x502: 0x000d, 0x503: 0x000d, 0x504: 0x000d, 0x505: 0x000d,
	0x506: 0x000d, 0x507: 0x000d, 0x508: 0x000d, 0x509: 0x000d, 0x50a: 0x000d, 0x50b: 0x000d,
	0x50c: 0x000d, 0x50d: 0x000d, 0x50e: 0x000d, 0x50f: 0x000d, 0x510: 0x000d, 0x511: 0x000d,
	0x512: 0x000d, 0x513: 0x000d, 0x514: 0x000d, 0x515: 0x000d, 0x516: 0x000d, 0x517: 0x000d,
	0x518: 0x000d, 0x519: 0x000d, 0x51a: 0x000d, 0x51b: 0x000d, 0x51c: 0x000d, 0x51d: 0x000d,
	0x51e: 0x000d, 0x51f: 0x000d, 0x520: 0x000d, 0x521: 0x000d, 0x522: 0x000d, 0x523: 0x000d,
	0x524: 0x000d, 0x525: 0x000d, 0x526: 0x000c, 0x527: 0x000c, 0x528: 0x000c, 0x529: 0x000c,
	0x52a: 0x000c, 0x52b: 0x000c, 0x52c: 0x000c, 0x52d: 0x000c, 0x52e: 0x000c, 0x52f: 0x000c,
	0x530: 0x000c, 0x531: 0x000d, 0x532: 0x000d, 0x533: 0x000d