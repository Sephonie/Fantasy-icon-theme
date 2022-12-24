// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

// +build !go1.10

package width

// UnicodeVersion is the Unicode version from which the tables in this package are derived.
const UnicodeVersion = "9.0.0"

// lookup returns the trie value for the first UTF-8 encoding in s and
// the width in bytes of this encoding. The size will be 0 if s does not
// hold enough bytes to complete the encoding. len(s) must be greater than 0.
func (t *widthTrie) lookup(s []byte) (v uint16, sz int) {
	c0 := s[0]
	switch {
	case c0 < 0x80: // is ASCII
		return widthValues[c0], 1
	case c0 < 0xC2:
		return 0, 1 // Illegal UTF-8: not a starter, not ASCII.
	case c0 < 0xE0: // 2-byte UTF-8
		if len(s) < 2 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c1), 2
	case c0 < 0xF0: // 3-byte UTF-8
		if len(s) < 3 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = widthIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c2), 3
	case c0 < 0xF8: // 4-byte UTF-8
		if len(s) < 4 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = widthIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		o = uint32(i)<<6 + uint32(c2)
		i = widthIndex[o]
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
func (t *widthTrie) lookupUnsafe(s []byte) uint16 {
	c0 := s[0]
	if c0 < 0x80 { // is ASCII
		return widthValues[c0]
	}
	i := widthIndex[c0]
	if c0 < 0xE0 { // 2-byte UTF-8
		return t.lookupValue(uint32(i), s[1])
	}
	i = widthIndex[uint32(i)<<6+uint32(s[1])]
	if c0 < 0xF0 { // 3-byte UTF-8
		return t.lookupValue(uint32(i), s[2])
	}
	i = widthIndex[uint32(i)<<6+uint32(s[2])]
	if c0 < 0xF8 { // 4-byte UTF-8
		return t.lookupValue(uint32(i), s[3])
	}
	return 0
}

// lookupString returns the trie value for the first UTF-8 encoding in s and
// the width in bytes of this encoding. The size will be 0 if s does not
// hold enough bytes to complete the encoding. len(s) must be greater than 0.
func (t *widthTrie) lookupString(s string) (v uint16, sz int) {
	c0 := s[0]
	switch {
	case c0 < 0x80: // is ASCII
		return widthValues[c0], 1
	case c0 < 0xC2:
		return 0, 1 // Illegal UTF-8: not a starter, not ASCII.
	case c0 < 0xE0: // 2-byte UTF-8
		if len(s) < 2 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c1), 2
	case c0 < 0xF0: // 3-byte UTF-8
		if len(s) < 3 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = widthIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		return t.lookupValue(uint32(i), c2), 3
	case c0 < 0xF8: // 4-byte UTF-8
		if len(s) < 4 {
			return 0, 0
		}
		i := widthIndex[c0]
		c1 := s[1]
		if c1 < 0x80 || 0xC0 <= c1 {
			return 0, 1 // Illegal UTF-8: not a continuation byte.
		}
		o := uint32(i)<<6 + uint32(c1)
		i = widthIndex[o]
		c2 := s[2]
		if c2 < 0x80 || 0xC0 <= c2 {
			return 0, 2 // Illegal UTF-8: not a continuation byte.
		}
		o = uint32(i)<<6 + uint32(c2)
		i = widthIndex[o]
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
func (t *widthTrie) lookupStringUnsafe(s string) uint16 {
	c0 := s[0]
	if c0 < 0x80 { // is ASCII
		return widthValues[c0]
	}
	i := widthIndex[c0]
	if c0 < 0xE0 { // 2-byte UTF-8
		return t.lookupValue(uint32(i), s[1])
	}
	i = widthIndex[uint32(i)<<6+uint32(s[1])]
	if c0 < 0xF0 { // 3-byte UTF-8
		return t.lookupValue(uint32(i), s[2])
	}
	i = widthIndex[uint32(i)<<6+uint32(s[2])]
	if c0 < 0xF8 { // 4-byte UTF-8
		return t.lookupValue(uint32(i), s[3])
	}
	return 0
}

// widthTrie. Total size: 14080 bytes (13.75 KiB). Checksum: 3b8aeb3dc03667a3.
type widthTrie struct{}

func newWidthTrie(i int) *widthTrie {
	return &widthTrie{}
}

// lookupValue determines the type of block n and looks up the value for b.
func (t *widthTrie) lookupValue(n uint32, b byte) uint16 {
	switch {
	default:
		return uint16(widthValues[n<<6+uint32(b)])
	}
}

// widthValues: 99 blocks, 6336 entries, 12672 bytes
// The third block is the zero block.
var widthValues = [6336]uint16{
	// Block 0x0, offset 0x0
	0x20: 0x6001, 0x21: 0x6002, 0x22: 0x6002, 0x23: 0x6002,
	0x24: 0x6002, 0x25: 0x6002, 0x26: 0x6002, 0x27: 0x6002, 0x28: 0x6002, 0x29: 0x6002,
	0x2a: 0x6002, 0x2b: 0x6002, 0x2c: 0x6002, 0x2d: 0x6002, 0x2e: 0x6002, 0x2f: 0x6002,
	0x30: 0x6002, 0x31: 0x6002, 0x32: 0x6002, 0x33: 0x6002, 0x34: 0x6002, 0x35: 0x6002,
	0x36: 0x6002, 0x37: 0x6002, 0x38: 0x6002, 0x39: 0x6002, 0x3a: 0x6002, 0x3b: 0x6002,
	0x3c: 0x6002, 0x3d: 0x6002, 0x3e: 0x6002, 0x3f: 0x6002,
	// Block 0x1, offset 0x40
	0x40: 0x6003, 0x41: 0x6003, 0x42: 0x6003, 0x43: 0x6003, 0x44: 0x6003, 0x45: 0x6003,
	0x46: 0x6003, 0x47: 0x6003, 0x48: 0x6003, 0x49: 0x6003, 0x4a: 0x6003, 0x4b: 0x6003,
	0x4c: 0x6003, 0x4d: 0x6003, 0x4e: 0x6003, 0x4f: 0x6003, 0x50: 0x6003, 0x51: 0x6003,
	0x52: 0x6003, 0x53: 0x6003, 0x54: 0x6003, 0x55: 0x6003, 0x56: 0x6003, 0x57: 0x6003,
	0x58: 0x6003, 0x59: 0x6003, 0x5a: 0x6003, 0x5b: 0x6003, 0x5c: 0x6003, 0x5d: 0x6003,
	0x5e: 0x6003, 0x5f: 0x6003, 0x60: 0x6004, 0x61: 0x6004, 0x62: 0x6004, 0x63: 0x6004,
	0x64: 0x6004, 0x65: 0x6004, 0x66: 0x6004, 0x67: 0x6004, 0x68: 0x6004, 0x69: 0x6004,
	0x6a: 0x6004, 0x6b: 0x6004, 0x6c: 0x6004, 0x6d: 0x6004, 0x6e: 0x6004, 0x6f: 0x6004,
	0x70: 0x6004, 0x71: 0x6004, 0x72: 0x6004, 0x73: 0x6004, 0x74: 0x6004, 0x75: 0x6004,
	0x76: 0x6004, 0x77: 0x6004, 0x78: 0x6004, 0x79: 0x6004, 0x7a: 0x6004, 0x7b: 0x6004,
	0x7c: 0x6004, 0x7d: 0x6004, 0x7e: 0x6004,
	// Block 0x2, offset 0x80
	// Block 0x3, offset 0xc0
	0xe1: 0x2000, 0xe2: 0x6005, 0xe3: 0x6005,
	0xe4: 0x2000, 0xe5: 0x6006, 0xe6: 0x6005, 0xe7: 0x2000, 0xe8: 0x2000,
	0xea: 0x2000, 0xec: 0x6007, 0xed: 0x2000, 0xee: 0x2000, 0xef: 0x6008,
	0xf0: 0x2000, 0xf1: 0x2000, 0xf2: 0x2000, 0xf3: 0x2000, 0xf4: 0x2000,
	0xf6: 0x2000, 0xf7: 0x2000, 0xf8: 0x2000, 0xf9: 0x2000, 0xfa: 0x2000,
	0xfc: 0x2000, 0xfd: 0x2000, 0xfe: 0x2000, 0xff: 0x2000,
	// Block 0x4, offset 0x100
	0x106: 0x2000,
	0x110: 0x2000,
	0x117: 0x2000,
	0x118: 0x2000,
	0x11e: 0x2000, 0x11f: 0x2000, 0x120: 0x2000, 0x121: 0x2000,
	0x126: 0x2000, 0x128: 0x2000, 0x129: 0x2000,
	0x12a: 0x2000, 0x12c: 0x2000, 0x12d: 0x2000,
	0x130: 0x2000, 0x132: 0x2000, 0x133: 0x2000,
	0x137: 0x2000, 0x138: 0x2000, 0x139: 0x2000, 0x13a: 0x2000,
	0x13c: 0x2000, 0x13e: 0x2000,
	// Block 0x5, offset 0x140
	0x141: 0x2000,
	0x151: 0x2000,
	0x153: 0x2000,
	0x15b: 0x2000,
	0x166: 0x2000, 0x167: 0x2000,
	0x16b: 0x2000,
	0x171: 0x2000, 0x172: 0x2000, 0x173: 0x2000,
	0x178: 0x2000,
	0x17f: 0x2000,
	// Block 0x6, offset 0x180
	0x180: 0x2000, 0x181: 0x2000, 0x182: 0x2000, 0x184: 0x2000,
	0x188: 0x2000, 0x189: 0x2000, 0x18a: 0x2000, 0x18b: 0x2000,
	0x18d: 0x2000,
	0x192: 0x2000, 0x193: 0x2000,
	0x1a6: 0x2000, 0x1a7: 0x2000,
	0x1ab: 0x2000,
	// Block 0x7, offset 0x1c0
	0x1ce: 0x2000, 0x1d0: 0x2000,
	0x1d2: 0x2000, 0x1d4: 0x2000, 0x1d6: 0x2000,
	0x1d8: 0x2000, 0x1da: 0x2000, 0x1dc: 0x2000,
	// Block 0x8, offset 0x200
	0x211: 0x2000,
	0x221: 0x2000,
	// Block 0x9, offset 0x240
	0x244: 0x2000,
	0x247: 0x2000, 0x249: 0x2000, 0x24a: 0x2000, 0x24b: 0x2000,
	0x24d: 0x2000, 0x250: 0x2000,
	0x258: 0x2000, 0x259: 0x2000, 0x25a: 0x2000, 0x25b: 0x2000, 0x25d: 0x2000,
	0x25f: 0x2000,
	// Block 0xa, offset 0x280
	0x280: 0x2000, 0x281: 0x2000, 0x282: 0x2000, 0x283: 0x2000, 0x284: 0x2000, 0x285: 0x2000,
	0x286: 0x2000, 0x287: 0x2000, 0x288: 0x2000, 0x289: 0x2000, 0x28a: 0x2000, 0x28b: 0x2000,
	0x28c: 0x2000, 0x28d: 0x2000, 0x28e: 0x2000, 0x28f: 0x2000, 0x290: 0x2000, 0x291: 0x2000,
	0x292: 0x2000, 0x293: 0x2000, 0x294: 0x2000, 0x295: 0x2000, 0x296: 0x2000, 0x297: 0x2000,
	0x298: 0x2000, 0x299: 0x2000, 0x29a: 0x2000, 0x29b: 0x2000, 0x29c: 0x2000, 0x29d: 0x2000,
	0x29e: 0x2000, 0x29f: 0x2000, 0x2a0: 0x2000, 0x2a1: 0x2000, 0x2a2: 0x2000, 0x2a3: 0x2000,
	0x2a4: 0x2000, 0x2a5: 0x2000, 0x2a6: 0x2000, 0x2a7: 0x2000, 0x2a8: 0x2000, 0x2a9: 0x2000,
	0x2aa: 0x2000, 0x2ab: 0x2000, 0x2ac: 0x2000, 0x2ad: 0x2000, 0x2ae: 0x2000, 0x2af: 0x2000,
	0x2b0: 0x2000, 0x2b1: 0x2000, 0x2b2: 0x2000, 0x2b3: 0x2000, 0x2b4: 0x2000, 0x2b5: 0x2000,
	0x2b6: 0x2000, 0x2b7: 0x2000, 0x2b8: 0x2000, 0x2b9: 0x2000, 0x2ba: 0x2000, 0x2bb: 0x2000,
	0x2bc: 0x2000, 0x2bd: 0x2000, 0x2be: 0x2000, 0x2bf: 0x2000,
	// Block 0xb, offset 0x2c0
	0x2c0: 0x2000, 0x2c1: 0x2000, 0x2c2: 0x2000, 0x2c3: 0x2000, 0x2c4: 0x2000, 0x2c5: 0x2000,
	0x2c6: 0x2000, 0x2c7: 0x2000, 0x2c8: 0x2000, 0x2c9: 0x2000, 0x2ca: 0x2000, 0x2cb: 0x2000,
	0x2cc: 0x2000, 0x2cd: 0x2000, 0x2ce: 0x2000, 0x2cf: 0x2000, 0x2d0: 0x2000, 0x2d1: 0x2000,
	0x2d2: 0x2000, 0x2d3: 0x2000, 0x2d4: 0x2000, 0x2d5: 0x2000, 0x2d6: 0x2000, 0x2d7: 0x2000,
	0x2d8: 0x2000, 0x2d9: 0x2000, 0x2da: 0x2000, 0x2db: 0x2000, 0x2dc: 0x2000, 0x2dd: 0x2000,
	0x2de: 0x2000, 0x2df: 0x2000, 0x2e0: 0x2000, 0x2e1: 0x2000, 0x2e2: 0x2000, 0x2e3: 0x2000,
	0x2e4: 0x2000, 0x2e5: 0x2000, 0x2e6: 0x2000, 0x2e7: 0x2000, 0x2e8: 0x2000, 0x2e9: 0x2000,
	0x2ea: 0x2000, 0x2eb: 0x2000, 0x2ec: 0x2000, 0x2ed: 0x2000, 0x2ee: 0x2000, 0x2ef: 0x2000,
	// Block 0xc, offset 0x300
	0x311: 0x2000,
	0x312: 0x2000, 0x313: 0x2000, 0x314: 0x2000, 0x315: 0x2000, 0x316: 0x2000, 0x317: 0x2000,
	0x318: 0x2000, 0x319: 0x2000, 0x31a: 0x2000, 0x31b: 0x2000, 0x31c: 0x2000, 0x31d: 0x2000,
	0x31e: 0x2000, 0x31f: 0x2000, 0x320: 0x2000, 0x321: 0x2000, 0x323: 0x2000,
	0x324: 0x2000, 0x325: 0x2000, 0x326: 0x2000, 0x327: 0x2000, 0x328: 0x2000, 0x329: 0x2000,
	0x331: 0x2000, 0x332: 0x2000, 0x333: 0x2000, 0x334: 0x2000, 0x335: 0x2000,
	0x336: 0x2000, 0x337: 0x2000, 0x338: 0x2000, 0x339: 0x2000, 0x33a: 0x2000, 0x33b: 0x2000,
	0x33c: 0x2000, 0x33d: 0x2000, 0x33e: 0x2000, 0x33f: 0x2000,
	// Block 0xd, offset 0x340
	0x340: 0x2000, 0x341: 0x2000, 0x343: 0x2000, 0x344: 0x2000, 0x345: 0x2000,
	0x346: 0x2000, 0x347: 0x2000, 0x348: 0x2000, 0x349: 0x2000,
	// Block 0xe, offset 0x380
	0x381: 0x2000,
	0x390: 0x2000, 0x391: 0x2000,
	0x392: 0x2000, 0x393: 0x2000, 0x394: 0x2000, 0x395: 0x2000, 0x396: 0x2000, 0x397: 0x2000,
	0x398: 0x2000, 0x399: 0x2000, 0x39a: 0x2000, 0x39b: 0x2000, 0x39c: 0x2000, 0x39d: 0x2000,
	0x39e: 0x2000, 0x39f: 0x2000, 0x3a0: 0x2000, 0x3a1: 0x2000, 0x3a2: 0x2000, 0x3a3: 0x2000,
	0x3a4: 0x2000, 0x3a5: 0x2000, 0x3a6: 0x2000, 0x3a7: 0x2000, 0x3a8: 0x2000, 0x3a9: 0x2000,
	0x3aa: 0x2000, 0x3ab: 0x2000, 0x3ac: 0x2000, 0x3ad: 0x2000, 0x3ae: 0x2000, 0x3af: 0x2000,
	0x3b0: 0x2000, 0x3b1: 0x2000, 0x3b2: 0x2000, 0x3b3: 0x2000, 0x3b4: 0x2000, 0x3b5: 0x2000,
	0x3b6: 0x2000, 0x3b7: 0x2000, 0x3b8: 0x2000, 0x3b9: 0x2000, 0x3ba: 0x2000, 0x3bb: 0x2000,
	0x3bc: 0x2000, 0x3bd: 0x2000, 0x3be: 0x2000, 0x3bf: 0x2000,
	// Block 0xf, offset 0x3c0
	0x3c0: 0x2000, 0x3c1: 0x2000, 0x3c2: 0x2000, 0x3c3: 0x2000, 0x3c4: 0x2000, 0x3c5: 0x2000,
	0x3c6: 0x2000, 0x3c7: 0x2000, 0x3c8: 0x2000, 0x3c9: 0x2000, 0x3ca: 0x2000, 0x3cb: 0x2000,
	0x3cc: 0x2000, 0x3cd: 0x2000, 0x3ce: 0x2000, 0x3cf: 0x2000, 0x3d1: 0x2000,
	// Block 0x10, offset 0x400
	0x400: 0x4000, 0x401: 0x4000, 0x402: 0x4000, 0x403: 0x4000, 0x404: 0x4000, 0x405: 0x4000,
	0x406: 0x4000, 0x407: 0x4000, 0x408: 0x4000, 0x409: 0x4000, 0x40a: 0x4000, 0x40b: 0x4000,
	0x40c: 0x4000, 0x40d: 0x4000, 0x40e: 0x4000, 0x40f: 0x4000, 0x410: 0x4000, 0x411: 0x4000,
	0x412: 0x4000, 0x413: 0x4000, 0x414: 0x4000, 0x415: 0x4000, 0x416: 0x4000, 0x417: 0x4000,
	0x418: 0x4000, 0x419: 0x4000, 0x41a: 0x4000, 0x41b: 0x4000, 0x41c: 0x4000, 0x41d: 0x4000,
	0x41e: 0x4000, 0x41f: 0x4000, 0x420: 0x4000, 0x421: 0x4000, 0x422: 0x4000, 0x423: 0x4000,
	0x424: 0x4000, 0x425: 0x4000, 0x426: 0x4000, 0x427: 0x4000, 0x428: 0x4000, 0x429: 0x4000,
	0x42a: 0x4000, 0x42b: 0x4000, 0x42c: 0x4000, 0x42d: 0x4000, 0x42e: 0x4000, 0x42f: 0x4000,
	0x430: 0x4000, 0x431: 0x4000, 0x432: 0x4000, 0x433: 0x4000, 0x434: 0x4000, 0x435: 0x4000,
	0x436: 0x4000, 0x437: 0x4000, 0x438: 0x4000, 0x439: 0x4000, 0x43a: 0x4000, 0x43b: 0x4000,
	0x43c: 0x4000, 0x43d: 0x4000, 0x43e: 0x4000, 0x43f: 0x4000,
	// Block 0x11, offset 0x440
	0x440: 0x4000, 0x441: 0x4000, 0x442: 0x4000, 0x443: 0x4000, 0x444: 0x4000, 0x445: 0x4000,
	0x446: 0x4000, 0x447: 0x4000, 0x448: 0x4000, 0x449: 0x4000, 0x44a: 0x4000, 0x44b: 0x4000,
	0x44c: 0x4000, 0x44d: 0x4000, 0x44e: 0x4000, 0x44f: 0x4000, 0x450: 0x4000, 0x451: 0x4000,
	0x452: 0x4000, 0x453: 0x4000, 0x454: 0x4000, 0x455: 0x4000, 0x456: 0x4000, 0x457: 0x4000,
	0x458: 0x4000, 0x459: 0x4000, 0x45a: 0x4000, 0x45b: 0x4000, 0x45c: 0x4000, 0x45d: 0x4000,
	0x45e: 0x4000, 0x45f: 0x4000,
	// Blo