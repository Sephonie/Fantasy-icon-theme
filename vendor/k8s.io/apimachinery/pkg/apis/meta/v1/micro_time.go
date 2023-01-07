/*
Copyright 2016 The Kubernetes Authors.

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

package v1

import (
	"encoding/json"
	"time"

	openapi "k8s.io/kube-openapi/pkg/common"

	"github.com/go-openapi/spec"
	"github.com/google/gofuzz"
)

const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"

// MicroTime is version of Time with microsecond level precision.
//
// +protobuf.options.marshal=false
// +protobuf.as=Timestamp
// +protobuf.options.(gogoproto.goproto_stringer)=false
type MicroTime struct {
	time.Time `protobuf:"-"`
}

// DeepCopy returns a deep-copy of the MicroTime value.  The underlying time.Time
// type is effectively immutable in the time API, so it is safe to
// copy-by-assign, despite the presence of (unexported) Pointer fields.
func (t *MicroTime) DeepCopyInto(out *MicroTime) {
	*out = *t
}

// String returns the representation of the time.
func (t MicroTime) String() string {
	return t.Time.String()
}

// NewMicroTime returns a wrapped instance of the provided time
func NewMicroTime(time time.Time) MicroTime {
	return MicroTime{time}
}

// DateMicro returns the MicroTime corresponding to the supplied parameters
// by wrapping time.Date.
func DateMicro(year int, month time.Month, day, hour, m