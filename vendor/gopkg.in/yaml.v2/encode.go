
package yaml

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type encoder struct {
	emitter yaml_emitter_t
	event   yaml_event_t
	out     []byte
	flow    bool
	// doneInit holds whether the initial stream_start_event has been
	// emitted.
	doneInit bool
}

func newEncoder() *encoder {
	e := &encoder{}
	yaml_emitter_initialize(&e.emitter)
	yaml_emitter_set_output_string(&e.emitter, &e.out)
	yaml_emitter_set_unicode(&e.emitter, true)
	return e
}

func newEncoderWithWriter(w io.Writer) *encoder {
	e := &encoder{}
	yaml_emitter_initialize(&e.emitter)
	yaml_emitter_set_output_writer(&e.emitter, w)
	yaml_emitter_set_unicode(&e.emitter, true)
	return e
}

func (e *encoder) init() {
	if e.doneInit {
		return
	}
	yaml_stream_start_event_initialize(&e.event, yaml_UTF8_ENCODING)
	e.emit()
	e.doneInit = true
}

func (e *encoder) finish() {
	e.emitter.open_ended = false
	yaml_stream_end_event_initialize(&e.event)
	e.emit()
}

func (e *encoder) destroy() {
	yaml_emitter_delete(&e.emitter)
}

func (e *encoder) emit() {
	// This will internally delete the e.event value.
	e.must(yaml_emitter_emit(&e.emitter, &e.event))
}

func (e *encoder) must(ok bool) {
	if !ok {
		msg := e.emitter.problem
		if msg == "" {
			msg = "unknown problem generating YAML content"
		}
		failf("%s", msg)
	}
}

func (e *encoder) marshalDoc(tag string, in reflect.Value) {
	e.init()
	yaml_document_start_event_initialize(&e.event, nil, nil, true)
	e.emit()
	e.marshal(tag, in)
	yaml_document_end_event_initialize(&e.event, true)
	e.emit()
}

func (e *encoder) marshal(tag string, in reflect.Value) {
	if !in.IsValid() || in.Kind() == reflect.Ptr && in.IsNil() {
		e.nilv()
		return
	}
	iface := in.Interface()
	switch m := iface.(type) {
	case time.Time, *time.Time:
		// Although time.Time implements TextMarshaler,
		// we don't want to treat it as a string for YAML
		// purposes because YAML has special support for
		// timestamps.
	case Marshaler:
		v, err := m.MarshalYAML()
		if err != nil {
			fail(err)
		}
		if v == nil {
			e.nilv()
			return
		}
		in = reflect.ValueOf(v)
	case encoding.TextMarshaler:
		text, err := m.MarshalText()
		if err != nil {
			fail(err)
		}