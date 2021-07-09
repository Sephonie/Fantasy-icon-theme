package restful

import (
	"errors"
	"os"
	"reflect"
	"sync"

	"github.com/emicklei/go-restful/log"
)

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

// WebService holds a collection of Route value