package store

import "bytes"

// SegmentFileName
// Returns a file name that includes the given segment name, your own custom name
// and extension. The format of the filename is: <segmentName>(_<name>)(.<ext>).
// NOTE: .<ext> is added to the result file name only if ext is not empty.
// NOTE: _<segmentSuffix> is added to the result file name only if it's not the empty string
// NOTE: all custom files should be named using this method, or otherwise some structures may fail
// to handle them properly (such as if they are added to compound files).
func SegmentFileName(segmentName, segmentSuffix, ext string) string {
	buf := new(bytes.Buffer)
	buf.WriteString(segmentName)
	if len(segmentSuffix) > 0 {
		buf.WriteString("_")
		buf.WriteString(segmentSuffix)
	}

	if len(ext) > 0 {
		buf.WriteString(".")
		buf.WriteString(ext)
	}
	return buf.String()
}
