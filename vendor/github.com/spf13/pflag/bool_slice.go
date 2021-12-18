package pflag

import (
	"io"
	"strconv"
	"strings"
)

// -- boolSlice Value
type boolSliceValue struct {
	value   *[]bool
	changed bool
}

func newBoolSliceValue(val []bool, p *[]bool) *boolSliceValue {
	bsv := new(boolSliceValue)
	bsv.value = p
	*bsv.value = val
	return bsv
}

// Set converts, and assigns, the comma-separated boolean argument string representation as the [