package utils

import (
	"math"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const MaxNumDigits = 10

func IsValidManifestURL(u string) bool {
	return strings.HasSuffix(u, ".m3u8")
}

func GetSegmentURLPrefix(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsedURL.Scheme + "://" + parsedURL.Host + path.Dir(parsedURL.Path), nil
}

func EnrichSegmentWithDir(u string, segName string) string {
	return u + "/" + segName
}

func IsNameAdmissible(name string) bool {
	if name == "" {
		return false
	}
	namePattern := "^[a-zA-Z0-9_]+$"
	ok, err := regexp.MatchString(namePattern, name)
	if err != nil {
		panic(err.Error())
	}
	return ok
}

func GetSegmentFileName(dir string, seqID uint64) string {
	return path.Join(dir, GetSegmentFileNameNoDirNoExt(seqID)+".ts")
}

func GetSegmentFileNameNoDirNoExt(seqID uint64) string {
	return "segment" + "_" + strings.Repeat("0", numZerosPrefixed(seqID)) + strconv.FormatUint(seqID, 10)
}

func numZerosPrefixed(seqID uint64) int {
	return int(math.Max(float64(MaxNumDigits-len(strconv.FormatUint(seqID, 10))), 0))
}
