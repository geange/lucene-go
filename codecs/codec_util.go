package codecs

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
)

const (
	CODEC_MAGIC  = 0x3fd76c17
	FOOTER_MAGIC = ^CODEC_MAGIC
)

// WriteHeader Writes a codec header, which records both a string to identify the file and a version number. This header can be parsed and validated with checkHeader().
// CodecHeader --> Magic,CodecName,Version
// Magic --> Uint32. This identifies the start of the header. It is always 1071082519.
// CodecName --> String. This is a string to identify this file.
// Version --> Uint32. Records the version of the file.
// Note that the length of a codec header depends only upon the name of the codec, so this length can be computed at any time with headerLength(String).
// Params: out – Output stream codec – String to identify this file. It should be simple ASCII, less than 128 characters in length. version – Version number
// Throws: 	IOException – If there is an I/O error writing to the underlying medium.
//
//	IllegalArgumentException – If the codec name is not simple ASCII, or is more than 127 characters in length
func WriteHeader(out store.DataOutput, codec string, version int) error {
	if err := out.WriteUint32(CODEC_MAGIC); err != nil {
		return err
	}
	if err := out.WriteString(codec); err != nil {
		return err
	}
	if err := out.WriteUint32(uint32(version)); err != nil {
		return err
	}
	return nil
}

// CheckHeader Reads and validates a header previously written with writeHeader(DataOutput, String, int).
// When reading a file, supply the expected codec and an expected version range (minVersion to maxVersion).
// Params: in – Input stream, positioned at the point where the header was previously written. Typically this is located at the beginning of the file. codec – The expected codec name. minVersion – The minimum supported expected version number. maxVersion – The maximum supported expected version number.
// Returns: The actual version found, when a valid header is found that matches codec, with an actual version where minVersion <= actual <= maxVersion. Otherwise an exception is thrown.
// Throws: 	CorruptIndexException – If the first four bytes are not CODEC_MAGIC, or if the actual codec found is not codec.
//
//	IndexFormatTooOldException – If the actual version is less than minVersion.
//	IndexFormatTooNewException – If the actual version is greater than maxVersion.
//	IOException – If there is an I/O error reading from the underlying medium.
//
// See Also: writeHeader(DataOutput, String, int)
func CheckHeader(in store.DataInput, codec string, minVersion, maxVersion int) (int, error) {
	// Safety to guard against reading a bogus string:
	actualHeader, err := in.ReadUint32()
	if err != nil {
		return 0, err
	}
	if actualHeader != CODEC_MAGIC {
		return 0, errors.New("codec header mismatch")
	}
	return CheckHeaderNoMagic(in, codec, minVersion, maxVersion)
}

// CheckHeaderNoMagic Like checkHeader(DataInput, String, int, int) except this version assumes
// the first int has already been read and validated from the input.
func CheckHeaderNoMagic(in store.DataInput, codec string, minVersion, maxVersion int) (int, error) {
	actualCodec, err := in.ReadString()
	if err != nil {
		return 0, err
	}
	if actualCodec != codec {
		return 0, errors.New("codec mismatch")
	}

	actualVersion, err := in.ReadUint32()
	if err != nil {
		return 0, err
	}

	if int(actualVersion) < minVersion || int(actualVersion) > maxVersion {
		return 0, errors.New("IndexFormatTooOld")
	}
	return int(actualVersion), nil
}
