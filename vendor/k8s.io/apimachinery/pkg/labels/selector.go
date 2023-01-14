/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package labels

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Requirements is AND of all requirements.
type Requirements []Requirement

// Selector represents a label selector.
type Selector interface {
	// Matches returns true if this selector matches the given set of labels.
	Matches(Labels) bool

	// Empty returns true if this selector does not restrict the selection space.
	Empty() bool

	// String returns a human readable string that represents this selector.
	String() string

	// Add adds requirements to the Selector
	Add(r ...Requirement) Selector

	// Requirements converts this interface into Requirements to expose
	// more detailed selection information.
	// If there are querying parameters, it will return converted requirements and selectable=true.
	// If this selector doesn't want to select anything, it will return selectable=false.
	Requirements() (requirements Requirements, selectable bool)

	// Make a deep copy of the selector.
	DeepCopySelector() Selector
}

// Everything returns a selector that matches all labels.
func Everything() Selector {
	return internalSelector{}
}

type nothingSelector struct{}

func (n nothingSelector) Matches(_ Labels) bool              { return false }
func (n nothingSelector) Empty() bool                        { return false }
func (n nothingSelector) String() string                     { return "" }
func (n nothingSelector) Add(_ ...Requirement) Selector      { return n }
func (n nothingSelector) Requirements() (Requirements, bool) { return nil, false }
func (n nothingSelector) DeepCopySelector() Selector         { return n }

// Nothing returns a selector that matches no labels
func Nothing() Selector {
	return nothingSelector{}
}

// NewSelector returns a nil selector
func NewSelector() Selector {
	return internalSelector(nil)
}

type internalSelector []Requirement

func (s internalSelector) DeepCopy() internalSelector {
	if s == nil {
		return nil
	}
	result := make([]Requirement, len(s))
	for i := range s {
		s[i].DeepCopyInto(&result[i])
	}
	return result
}

func (s internalSelector) DeepCopySelector() Selector {
	return s.DeepCopy()
}

// ByKey sorts requirements by key to obtain deterministic parser
type ByKey []Requirement

func (a ByKey) Len() int { return len(a) }

func (a ByKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByKey) Less(i, j int) bool { return a[i].key < a[j].key }

// Requirement contains values, a key, and an operator that relates the key and values.
// The zero value of Requirement is invalid.
// Requirement implements both set based match and exact match
// Requirement should be initialized via NewRequirement constructor for creating a valid Requirement.
// +k8s:deepcopy-gen=true
type Requirement struct {
	key      string
	operator selection.Operator
	// In huge majority of cases we have at most one value here.
	// It is generally faster to operate on a single-element slice
	// than on a single-element map, so we have a slice here.
	strValues []string
}

// NewRequirement is the constructor for a Requirement.
// If any of these rules is violated, an error is returned:
// (1) The operator can only be In, NotIn, Equals, DoubleEquals, NotEquals, Exists, or DoesNotExist.
// (2) If the operator is In or NotIn, the values set must be non-empty.
// (3) If the operator is Equals, DoubleEquals, or NotEquals, the values set must contain one value.
// (4) If the operator is Exists or DoesNotExist, the value set must be empty.
// (5) If the operator is Gt or Lt, the values set must contain only one value, which will be interpreted as an integer.
// (6) The key is invalid due to its length, or sequence
//     of characters. See validateLabelKey for more details.
//
// The empty string is a valid value in the input values set.
func NewRequirement(key string, op selection.Operator, vals []string) (*Requirement, error) {
	if err := validateLabelKey(key); err != nil {
		return nil, err
	}
	switch op {
	case selection.In, selection.NotIn:
		if len(vals) == 0 {
			return nil, fmt.Errorf("for 'in', 'notin' operators, values set can't be empty")
		}
	case selection.Equals, selection.DoubleEquals, selection.NotEquals:
		if len(vals) != 1 {
			return nil, fmt.Errorf("exact-match compatibility requires one single value")
		}
	case selection.Exists, selection.DoesNotExist:
		if len(vals) != 0 {
			return nil, fmt.Errorf("values set must be empty for exists and does not exist")
		}
	case selection.GreaterThan, selection.LessThan:
		if len(vals) != 1 {
			return nil, fmt.Errorf("for 'Gt', 'Lt' operators, exactly one value is required")
		}
		for i := range vals {
			if _, err := strconv.ParseInt(vals[i], 10, 64); err != nil {
				return nil, fmt.Errorf("for 'Gt', 'Lt' operators, the value must be an integer")
			}
		}
	default:
		return nil, fmt.Errorf("operator '%v' is not recognized", op)
	}

	for i := range vals {
		if err := validateLabelValue(vals[i]); err != nil {
			return nil, err
		}
	}
	sort.Strings(vals)
	return &Requirement{key: key, operator: op, strValues: vals}, nil
}

func (r *Requirement) hasValue(value string) bool {
	for i := range r.strValues {
		if r.strValues[i] == value {
			return true
		}
	}
	return false
}

// Matches returns true if the Requirement matches the input Labels.
// There is a match in the following cases:
// (1) The operator is Exists and Labels has the Requirement's key.
// (2) The operator is In, Labels has the Requirement's key and Labels'
//     value for that key is in Requirement's value set.
// (3) The operator is NotIn, Labels has the Requirement's key and
//     Labels' value for that key is not in Requirement's value set.
// (4) The operator is DoesNotExist or NotIn and Labels does not have the
//     Requirement's key.
// (5) The operator is GreaterThanOperator or LessThanOperator, and Labels has
//     the Requirement's key and the corresponding value satisfies mathematical inequality.
func (r *Requirement) Matches(ls Labels) bool {
	switch r.operator {
	case selection.In, selection.Equals, selection.DoubleEquals:
		if !ls.Has(r.key) {
			return false
		}
		return r.hasValue(ls.Get(r.key))
	case selection.NotIn, selection.NotEquals:
		if !ls.Has(r.key) {
			return true
		}
		return !r.hasValue(ls.Get(r.key))
	case selection.Exists:
		return ls.Has(r.key)
	case selection.DoesNotExist:
		return !ls.Has(r.key)
	case selection.GreaterThan, selection.LessThan:
		if !ls.Has(r.key) {
			return false
		}
		lsValue, err := strconv.ParseInt(ls.Get(r.key), 10, 64)
		if err != nil {
			glog.V(10).Infof("ParseInt failed for value %+v in label %+v, %+v", ls.Get(r.key), ls, err)
			return false
		}

		// There should be only one strValue in r.strValues, and can be converted to a integer.
		if len(r.strValues) != 1 {
			glog.V(10).Infof("Invalid values count %+v of requirement %#v, for 'Gt', 'Lt' operators, exactly one value is required", len(r.strValues), r)
			return false
		}

		var rValue int64
		for i := range r.strValues {
			rValue, err = strconv.ParseInt(r.strValues[i], 10, 64)
			if err != nil {
				glog.V(10).Infof("ParseInt failed for value %+v in requirement %#v, for 'Gt', 'Lt' operators, the value must be an integer", r.strValues[i], r)
				return false
			}
		}
		return (r.operator == selection.GreaterThan && lsValue > rValue) || (r.operator == selection.LessThan && lsValue < rValue)
	default:
		return false
	}
}

// Key returns requirement key
func (r *Requirement) Key() string {
	return r.key
}

// Operator returns requirement operator
func (r *Requirement) Operator() selection.Operator {
	return r.operator
}

// Values returns requirement values
func (r *Requirement) Values() sets.String {
	ret := sets.String{}
	for i := range r.strValues {
		ret.Insert(r.strValues[i])
	}
	return ret
}

// Empty returns true if the internalSelector doesn't restrict selection space
func (lsel internalSelector) Empty() bool {
	if lsel == nil {
		return true
	}
	return len(lsel) == 0
}

// String returns a human-readable string that represents this
// Requirement. If called on an invalid Requirement, an error is
// returned. See NewRequirement for creating a valid Requirement.
func (r *Requirement) String() string {
	var buffer bytes.Buffer
	if r.operator == selection.DoesNotExist {
		buffer.WriteString("!")
	}
	buffer.WriteString(r.key)

	switch r.operator {
	case selection.Equals:
		buffer.WriteString("=")
	case selection.DoubleEquals:
		buffer.WriteString("==")
	case selection.NotEquals:
		buffer.WriteString("!=")
	case selection.In:
		buffer.WriteString(" in ")
	case selection.NotIn:
		buffer.WriteString(" notin ")
	case selection.GreaterThan:
		buffer.WriteString(">")
	case selection.LessThan:
		buffer.WriteString("<")
	case selection.Exists, selection.DoesNotExist:
		return buffer.String()
	}

	switch r.operator {
	case selection.In, selection.NotIn:
		buffer.WriteString("(")
	}
	if len(r.strValues) == 1 {
		buffer.WriteString(r.strValues[0])
	} else { // only > 1 since == 0 prohibited by NewRequirement
		buffer.WriteString(strings.Join(r.strValues, ","))
	}

	switch r.operator {
	case selection.In, selection.NotIn:
		buffer.WriteString(")")
	}
	return buffer.String()
}

// Add adds requirements to the selector. It copies the current selector returning a new one
func (lsel internalSelector) Add(reqs ...Requirement) Selector {
	var sel internalSelector
	for ix := range lsel {
		sel = append(sel, lsel[ix])
	}
	for _, r := range reqs {
		sel = append(sel, r)
	}
	sort.Sort(ByKey(sel))
	return sel
}

// Matches for a internalSelector returns true if all
// its Requirements match the input Labels. If any
// Requirement does not match, false is returned.
func (lsel internalSelector) Matches(l Labels) bool {
	for ix := range lsel {
		if matches := lsel[ix].Matches(l); !matches {
			return false
		}
	}
	return true
}

func (lsel internalSelector) Requirements() (Requirements, bool) { return Requirements(lsel), true }

// String returns a comma-separated string of all
// the internalSelector Requirements' human-readable strings.
func (lsel internalSelector) String() string {
	var reqs []string
	for ix := range lsel {
		reqs = append(reqs, lsel[ix].String())
	}
	return strings.Join(reqs, ",")
}

// Token represents constant definition for lexer token
type Token int

const (
	// ErrorToken represents scan error
	ErrorToken Token = iota
	// EndOfString