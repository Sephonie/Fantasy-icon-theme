package yaml

import (
	"io"
)

// Set the reader error and return 0.
func yaml_parser_set_reader_error(parser *yaml_parser_t, problem string, offset int, value int) bool {
	parser.error = yaml_READER_ERROR
	parser.problem = problem
	parser.problem_offset = offset
	parser.problem_value = value
	return false
}

// Byte order marks.
const (
	bom_UTF8    = "\xef\xbb\xbf"
	bom_UTF16LE = "\xff\xfe"
	bom_UTF16BE = "\xfe\xff"
)

// Determine the input stream encoding by checking the BOM symbol. If no BOM is
// found, the UTF-8 encoding is assumed. Return 1 on success, 0 on failure.
func yaml_parser_determine_encoding(parser *yaml_parser_t) bool {
	// Ensure that we had enough bytes in the raw buffer.
	for !parser.eof && len(parser.raw_buffer)-parser.raw_buffer_pos < 3 {
		if !yaml_parser_update_raw_buffer(parser) {
			return false
		}
	}

	// Determine the encoding.
	buf := parser.raw_buffer
	pos := parser.raw_buffer_pos
	avail := len(buf) - pos
	if avail >= 2 && buf[pos] == bom_UTF16LE[0] && buf[pos+1] == bom_UTF16LE[1] {
		parser.encoding = yaml_UTF16LE_ENCODING
		parser.raw_buffer_pos += 2
		parser.offset += 2
	} else if avail >= 2 && buf[pos] == bom_UTF16BE[0] && buf[pos+1] == bom_UTF16BE[1] {
		parser.encoding = yaml_UTF16BE_ENCODING
		parser.raw_buffer_pos += 2
		parser.offset += 2
	} else if avail >= 3 && buf[pos] == bom_UTF8[0] && buf[pos+1] == bom_UTF8[1] && buf[pos+2] == bom_UTF8[2] {
		parser.encoding = yaml_UTF8_ENCODING
		parser.raw_buffer_pos += 3
		parser.offset += 3
	} else {
		parser.encoding = yaml_UTF8_ENCODING
	}
	return true
}

// Update the raw buffer.
func yaml_parser_update_raw_buffer(parser *yaml_parser_t) bool {
	size_read := 0

	// Return if the raw buffer is full.
	if parser.raw_buffer_pos == 0 && len(parser.raw_buffer) == cap(parser.raw_buffer) {
		return true
	}

	// Return on EOF.
	if parser.eof {
		return true
	}

	// Move the remaining bytes in the raw buffer to the beginning.
	if parser.raw_buffer_pos > 0 && parser.raw_buffer_pos < len(parser.raw_buffer) {
		copy(parser.raw_buffer, parser.raw_buffer[parser.raw_buffer_pos:])
	}
	parser.raw_buffer = parser.raw_buffer[:len(parser.raw_buffer)-parser.raw_buffer_pos]
	parser.raw_buffer_pos = 0

	