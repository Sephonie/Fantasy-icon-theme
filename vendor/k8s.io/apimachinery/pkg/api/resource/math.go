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

package resource

import (
	"math/big"

	inf "gopkg.in/inf.v0"
)

const (
	// maxInt64Factors is the highest value that will be checked when removing factors of 10 from an int64.
	// It is also the maximum decimal digits that can be represented with an int64.
	maxInt64Factors = 18
)

var (
	// Commonly needed big.Int values-- treat as read only!
	bigTen      = big.NewInt(10)
	bigZero     = big.NewInt(0)
	bigOne      = big.NewInt(1)
	bigThousand = big.NewInt(1000)
	big1024     = big.NewInt(1024)

	// Commonly needed inf.Dec values-- treat as read only!
	decZero      = inf.NewDec(0, 0)
	decOne       = inf.NewDec(1, 0)
	decMinusOne  = inf.NewDec(-1, 0)
	decThousand  = inf.NewDec(1000, 0)
	dec1024      = inf.NewDec(1024, 0)
	decMinus1024 = inf.NewDec(-1024, 0)

	// Largest (in magnitude) number allowed.
	maxAllowed = infDecAmount{inf.NewDec((1<<63)-1, 0)} // == max int64

	// The maximum value we can represent milli-units for.
	// Compare with the return value of Quantity.Value() to
	// see if it's safe to use Quantity.MilliValue().
	MaxMilliValue = int64(((1 << 63) - 1) / 1000)
