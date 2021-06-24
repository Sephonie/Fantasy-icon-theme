/*
Package purell offers URL normalization as described on the wikipedia page:
http://en.wikipedia.org/wiki/URL_normalization
*/
package purell

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/urlesc"
	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// A set of normalization flags determines how a URL will
// be normalized.
type NormalizationFlags uint

const (
	// Safe normalizations
	FlagLowercaseScheme           NormalizationFlags = 1 << iota // HTTP://host -> http://host, applied by default in Go1.1
	FlagLowercaseHost                                            // http://HOST -> http://host
	FlagUppercaseEscapes                                         // http://host/t%ef -> http://host/t%EF
	FlagDecodeUnnecessaryEscapes                                 // http://host/t%41 -> http://host/tA
	FlagEncodeNecessaryEscapes                                   // http://host/!"#$ -> http://host/%21%22#$
	FlagRemoveDefaultPort                                        // http://host:80 -> http://host
	FlagRemoveEmptyQuerySeparator                                // http://host/path? -> http://host/path

	// Usually safe normalizations
	FlagRemoveTrailingSlash // http://host/path/ -> http://host/path
	FlagAddTrailingSlash    // http://host/path -> http://host/path/ (should choose only one of these add/remove trailing slash flags)
	FlagRemoveDotSegments   // http://host/path/./a/b/../c -> http://host/path/a/c

	// Unsafe normalizations
	FlagRemoveDirectoryIndex   // http://host/path/index.html -> http://host/path/
	FlagRemoveFragment         // http://host/path#fragment -> http://host/path
	FlagForceHTTP              // https://host -> http://host
	FlagRemoveDuplicateSlashes // http://host/path//a///b -> http://host/path/a/b
	FlagRemoveWWW              // http://www.host/ -> http://host/
	FlagAddWWW                 // http://host/ -> http://www.host/ (should choose only one of these add/remove WWW flags)
	FlagSortQuery              // http://host/path?c=3&b=2&a=1&b=1 -> http://host/path?a=1&b=1&b=2&c=3

	// Normalizations not in the wikipedia article, required to cover tests cases
	// submitted by jehiah
	FlagDecodeDWORDHost           // http://1113982867 -> http://66.102.7.147
	FlagDecodeOctalHost           // http://0102.0146.07.0223 -> http://66.102.7.147
	FlagDecodeHexHost             // http://0x42660793 -> http://66.102.7.147
	FlagRemoveUnnecessaryHostDots // http://.host../path -> http://host/path
	FlagRemoveEmptyPortSeparator  // http://host:/path -> http://host/path

	// Convenience set of safe normalizations
	FlagsSafe NormalizationFlags = FlagLowercaseHost | FlagLowercaseScheme | FlagUppercaseEscapes | FlagDecodeUnnecessaryEscapes | FlagEncodeNecessaryEscapes | FlagRemoveDefaultPort | FlagRemoveEmptyQuerySeparator

	// For convenience sets, "greedy" uses the "remove trailing slash" and "remove www. prefix" flags,
	// while "non-greedy" uses the "add (or keep) the trailing slash" and "add www. prefix".

	// Convenience set of usually safe normalizations (includes FlagsSafe)
	FlagsUsuallySafeGreedy    NormalizationFlags = FlagsSafe | FlagRemoveTrailingSlash | FlagRemoveDotSegments
	FlagsUsuallySafeNonGreedy NormalizationFlags = FlagsSafe | FlagAddTrailingSlash | FlagRemoveDotSegments

	// Convenience set of unsafe normalizations (includes FlagsUsuallySafe)
	FlagsUnsafeGreedy    NormalizationFlags = FlagsUsuallySafeGreedy | FlagRemoveDirectoryIndex | FlagRemoveFragment | FlagForceHTTP | FlagRemoveDuplicateSlashes | FlagRemoveWWW | FlagSortQuery
	FlagsUnsafeNonGreedy NormalizationFlags = FlagsUsuallySafeNonGreedy | FlagRemoveDirectoryIndex | FlagRemoveFragment | FlagForceHTTP | FlagRemoveDuplicateSlashes | FlagAddWWW | FlagSortQuery

	// Convenience set of all available flags
	FlagsAllGreedy    = FlagsUnsafeGreedy | FlagDecodeDWORDHost | FlagDecodeOctalHost | FlagDecodeHexHost | FlagRemoveUnnecessaryHostDots | FlagRemoveEmptyPortSeparator
	FlagsAllNonGreedy = FlagsUnsafeNonGreedy | FlagDecodeDWORDHost | FlagDecodeOctalHost | FlagDecodeHexHost | FlagRemoveUnnecessaryHostDots | FlagRemoveEmptyPortSeparator
)

const (
	defaultHttpPort  = ":80"
	defaultHttpsPort = ":443"
)

// Regular expressions used by the normalizations
var rxPort = regexp.MustCompile(`(:\d+)/?$`)
var rxDirIndex = regexp.MustCompile(`(^|/)((?:default|index)\.\w{1,4})$`)
var rxDupSlashes = regexp.MustCompile(`/{2,}`)
var rxDWORDHost = regexp.MustCompile(`^(\d+)((?:\.+)?(?:\:\d*)?)$`)
var rxOctalHost = regexp.MustCompile(`^(0\d*)\.(0\d*)\.(0\d*)\.(0\d*)((?:\.+)?(?:\:\d*)?)$`)
var rxHexHost = regexp.MustCompile(`^0x([0-9A-Fa-f]+)((?:\.+)?(?:\:\d*)?)$`)
var rxHostDots = regexp.MustCompile(`^(.+?)(:\d+)?$`)
var rxEmptyPort = regexp.MustCompile(`:+$`)

// Map of flags to implementation function.
// FlagDecodeUnnecessaryEscapes has no action, since it is done automatically
// by parsing the string as an URL. Same for FlagUppercaseEscapes and FlagRemoveEmptyQuerySeparator.

// Since maps have undefined traversing order, make a slice of ordered keys
var flagsOrder = []NormalizationFlags{
	FlagLowercaseScheme,
	FlagLowercaseHost,
	FlagRemoveDefaultPort,
	FlagRemoveDirectoryIndex,
	FlagRemoveDotSegments,
	FlagRemoveFragment,
	FlagForceHTTP, // Must be after remove default port (because https=443/http=80)
	FlagRemoveDuplicateSlashes,
	FlagRemoveWWW,
	FlagAddWWW,
	FlagSortQuery,
	FlagDecodeDWORDHost,
	FlagDecodeOctalHost,
	FlagDecodeHexHost,
	FlagRemoveUnnecessaryHostDots,
	FlagRemoveEmptyPortSeparator,
	FlagRemoveTrailingSlash, // These two (add/remove trailing slash) must be last
	FlagAddTrailingSlash,
}

// ... and then the map, where order is unimportant
var flags = map[NormalizationFlags]func(*url.URL){
	FlagLowercaseScheme:           lowercaseScheme,
	FlagLowercaseHost:             lowercaseHost,
	FlagRemoveDefaultPort:         removeDefaultPort,
	FlagRemoveDirectoryIndex:      removeDirectoryIndex,
	FlagRemoveDotSegments:         removeDotSegments,
	FlagRemoveFragment:            removeFragment,
	FlagForceHTTP:                 forceHTTP,
	FlagRemoveDuplicateSlashes:    removeDuplicateSlashes,
	FlagRemoveWWW:                 removeWWW,
	FlagAddWWW:                    addWWW,
	FlagSortQuery:                 sortQuery,
	FlagDecodeDWORDHost:           decodeDWORDHost,
	FlagDecodeOctalHost:           decodeOctalHost,
	FlagDecodeHexHost:             decodeHexHost,
	FlagRemoveUnnecessaryHostDots: removeUnncessaryHostDots,
	FlagRemoveEmptyPortSeparator:  removeEmptyPortSeparator,
	FlagRemoveTrailingSlash:       removeTrailingSlash,
	FlagAddTrailingSlash:          addTrailingSlash,
}

// MustNormalizeURLString returns the normalized string, and panics if an error occurs.
// It takes an URL string as input, as well as the normalization flags.
func MustNormalizeURLString(u string, f NormalizationFlags) string {
	result, e := NormalizeURLString(u, f)
	if e != nil {
		panic(e)
	}
	return result
}

// NormalizeURLString returns the normalized string, or an error if it can't be parsed into an URL object.
// It takes an URL string as input, as well as the normalization flags.
func NormalizeURLString(u string, f NormalizationFlags) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	if f&FlagLowercaseHost == FlagLowercaseHost {
		parsed.Host = strings.ToLower(pars