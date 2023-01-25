package index

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

const (
	// SEGMENTS Name of the index segment file
	SEGMENTS = "segments"

	// PENDING_SEGMENTS Name of pending index segment file
	PENDING_SEGMENTS = "pending_segments"

	// OLD_SEGMENTS_GEN Name of the generation reference file name
	OLD_SEGMENTS_GEN = "segments.gen"
)

func FileNameFromGeneration(base, ext string, gen int64) string {
	if gen == -1 {
		return ""
	} else if gen == 0 {
		return SegmentFileName(base, "", ext)
	} else {
		//assert gen > 0;
		// The '6' part in the length is: 1 for '.', 1 for '_' and 4 as estimate
		// to the gen length as string (hopefully an upper limit so SB won't
		// expand in the middle.
		res := new(bytes.Buffer)

		res.WriteString(base)
		res.WriteString("_")
		res.WriteString(strconv.FormatInt(gen, 36))

		if len(ext) > 0 {
			res.WriteString(".")
			res.WriteString(ext)
		}

		return res.String()
	}
}

// SegmentFileName Returns a file name that includes the given segment name, your own custom name and
// extension. The format of the filename is: <segmentName>(_<name>)(.<ext>).
// NOTE: .<ext> is added to the result file name only if ext is not empty.
// NOTE: _<segmentSuffix> is added to the result file name only if it's not the empty string
// NOTE: all custom files should be named using this method, or otherwise some structures may fail to
// handle them properly (such as if they are added to compound files).
func SegmentFileName(segmentName, segmentSuffix, ext string) string {
	if len(ext) > 0 || len(segmentSuffix) > 0 {
		if strings.HasPrefix(ext, ".") {
			return segmentName
		}

		sb := new(bytes.Buffer)
		sb.WriteString(segmentName)
		if len(segmentSuffix) > 0 {
			sb.WriteString("_")
			sb.WriteString(segmentSuffix)
		}
		if len(ext) > 0 {
			sb.WriteString(".")
			sb.WriteString(ext)
		}
		return sb.String()
	}

	return segmentName
}

// CODEC_FILE_PATTERN All files created by codecs much match this pattern (checked in SegmentInfo).
var (
	CODEC_FILE_PATTERN = regexp.MustCompilePOSIX("_[a-z0-9]+(_.*)?\\..*")
)
