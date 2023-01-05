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
)

const mostNegative = -(mostPositive + 1)
const mostPositive = 1<<63 - 1

// int64Add returns a+b, or false if that would overflow int64.
func int64Add(a, b int64) (int64, bool) {
	c := a + b
	switch {
	case a > 0 && b > 0:
		if c < 0 {
			return 0, false
		}
	case a < 0 && b < 0:
		if c > 0 {
			return 0, false
		}
		if a == mostNegative && b == mostNegative {
			return 0, false
		}
	}
	return c, true
}

// int64Multiply returns a*b, or false if that would overflow or underflow int64.
func int64Multiply(a, b int64) (int64, bool) {
	if a == 0 || b == 0 || a == 1 || b == 1 {
		return a * b, true
	}
	if a == mostNegative || b == mostNegative {
		return 0, false
	}
	c := a * b
	return c, c/b == a
}

// int64MultiplyScale returns a*b, assuming b is greater than one, or false if that would overflow or underflow int64.
// Use when b is known to be greater than one.
func int64MultiplyScale(a int64, b int64) (int64, bool) {
	if a == 0 || a == 1 {
		return a * b, true
	}
	if a == mostNegative && b != 1 {
		return 0, false
	}
	c := a * b
	return c, c/b == a
}

// int64MultiplyScale10 multiplies a by 10, or returns false if that would overflow. This method is faster than
// int64Multiply(a, 10) because the compiler can optimize constant factor multiplication.
func int64MultiplyScale10(a int64) (int64, bool) {
	if a == 0 || a == 1 {
		return a * 10, true
	}
	if a == mostNegative {
		return 0, false
	}
	c := a * 10
	return c, c/10 == a
}

// int64MultiplyScale100 multiplies a by 100, or returns false if that would overflow. This method is faster than
// int64Multiply(a, 100) because the compiler can optimize constant factor multiplication.
func int64MultiplyScale100(a int64) (int64, bool) {
	if a == 0 || a == 1 {
		return a * 100, true
	}
	if a == mostNegative {
		return 0, false
	}
	c := a * 100
	return c, c/100 == a
}

// int64MultiplyScale1000 multiplies a by 1000, or returns false if that would overflow. This method is faster than
// int64Multiply(a, 1000) because the compiler can optimize constant factor multiplication.
func int64MultiplyScale1000(a int64) (int64, bool) {
	if a == 0 || a == 1 {
		return a * 1000, true
	}
	if a == mostNegative {
		return 0, false
	}
	c := a * 1000
	return c, c/1000 == a
}

// positiveScaleInt64 multiplies base by 10^scale, returning false if the
// value overflows. Passing a negative scale is undefined.
func positiveScaleInt64(base int64, scale Scale) (int64, bool) {
	switch scale {
	case 0:
		return base, true
	case 1:
		return int64MultiplyScale10(base)
	case 2:
		return int64MultiplyScale100(base)
	case 3:
		return int64MultiplyScale1000(base)
	case 6:
		return int64MultiplyScale(base, 1000000)
	case 9:
		return int64MultiplyScale(base, 1000000000)
	default:
		value := base
		var ok bool
		for i := Scale(0); i < scale; i++ {
			if value, ok = int64MultiplyScale(value, 10); !ok {
				return 0, false
			}
		}
		return value, true
	}
}

// negativeScaleInt64 reduces base by the provided scale, rounding up, until the
// value is zero or the scale is reached. Passing a negative scale is undefined.
// The value returned, if not exact, is rounded away from zero.
func negativeScaleInt64(base int64, scale Scale) (result int64, exact bool) {
	if scale == 0 {
		return base, true
	}

	value := base
	var fraction bool
	for i := Scale(0); i < scale; i++ {
		if !fraction && value%10 != 0 {
			fraction = true
		}
		value = value / 10
		if value == 0 {
			if fraction {
				if base > 0 {
					return 1, false
				}
				return -1, false
			}
			return 0, true
		}
	}
	if fraction {
		if base > 0 {
			value += 1
		} else {
			value += -1
		}
	}
	return value, !fraction
}

func pow10Int64(