package utils

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/store"
)

const (
	CODEC_MAGIC  = 0x3fd76c17
	FOOTER_MAGIC = 0xc02893e8
)

// WriteHeader Writes a codec header, which records both a string to identify the file and a version number. This header can be parsed and validated with checkHeader().
// CodecHeader --> Magic,CodecName,Version
// Magic --> Uint32. This identifies the start of the header. It is always 1071082519.
// CodecName --> String. This is a string to identify this file.
// Version --> Uint32. Records the version of the file.
// Note that the length of a codec header depends only upon the name of the codec, so this length can be computed at any time with headerLength(String).
// out: Output stream
// codec: String to identify this file. It should be simple ASCII, less than 128 characters in length.
// version: Version number
// Throws: 	IOException – If there is an I/O error writing to the underlying medium.
//
//	IllegalArgumentException – If the codec name is not simple ASCII, or is more than 127 characters in length
func WriteHeader(ctx context.Context, out store.DataOutput, codec string, version int) error {
	if err := out.WriteUint32(ctx, CODEC_MAGIC); err != nil {
		return err
	}
	if err := out.WriteString(ctx, codec); err != nil {
		return err
	}
	if err := out.WriteUint32(ctx, uint32(version)); err != nil {
		return err
	}
	return nil
}

const (
	ID_LENGTH = 16
)

func WriteIndexHeader(ctx context.Context, out store.DataOutput, codec string, version int, id []byte, suffix string) error {
	if len(id) != ID_LENGTH {
		return fmt.Errorf("Invalid id: " + base64.StdEncoding.EncodeToString(id))
	}
	if err := WriteHeader(ctx, out, codec, version); err != nil {
		return err
	}
	if _, err := out.Write(id); err != nil {
		return err
	}

	suffixBytes := []byte(suffix)

	suffixSize := len([]rune(suffix))
	if suffixSize != len(suffixBytes) || suffixSize >= 256 {
		return errors.New("suffix must be simple ASCII, less than 256 characters in length")
	}

	if err := out.WriteByte(byte(suffixSize)); err != nil {
		return err
	}
	if _, err := out.Write(suffixBytes); err != nil {
		return err
	}
	return nil
}

// CheckHeader
// Reads and validates a header previously written with writeHeader(DataOutput, String, int).
// When reading a file, supply the expected codec and an expected version range (minVersion to maxVersion).
// in: Input stream, positioned at the point where the header was previously written. Typically this is
// located at the beginning of the file. codec – The expected codec name. minVersion – The minimum supported
// expected version number. maxVersion – The maximum supported expected version number.
// Returns: The actual version found, when a valid header is found that matches codec, with an actual version
// where minVersion <= actual <= maxVersion. Otherwise an exception is thrown.
// Throws:
// CorruptIndexException – If the first four bytes are not CODEC_MAGIC, or if the actual codec found is not codec.
// IndexFormatTooOldException – If the actual version is less than minVersion.
// IndexFormatTooNewException – If the actual version is greater than maxVersion.
// IOException – If there is an I/O error reading from the underlying medium.
//
// See Also: writeHeader(DataOutput, String, int)
func CheckHeader(ctx context.Context, in store.DataInput, codec string, minVersion, maxVersion int) (int, error) {
	// Safety to guard against reading a bogus string:
	actualHeader, err := in.ReadUint32(ctx)
	if err != nil {
		return 0, err
	}
	if actualHeader != CODEC_MAGIC {
		return 0, errors.New("codec header mismatch")
	}
	return CheckHeaderNoMagic(ctx, in, codec, minVersion, maxVersion)
}

// CheckHeaderNoMagic Like checkHeader(DataInput, String, int, int) except this version assumes
// the first int has already been read and validated from the input.
func CheckHeaderNoMagic(ctx context.Context, in store.DataInput, codec string, minVersion, maxVersion int) (int, error) {
	actualCodec, err := in.ReadString(ctx)
	if err != nil {
		return 0, err
	}
	if actualCodec != codec {
		return 0, errors.New("codec mismatch")
	}

	actualVersion, err := in.ReadUint32(ctx)
	if err != nil {
		return 0, err
	}

	if int(actualVersion) < minVersion || int(actualVersion) > maxVersion {
		return 0, errors.New("IndexFormatTooOld")
	}
	return int(actualVersion), nil
}

func CheckIndexHeaderSuffix(in store.DataInput, expectedSuffix string) (string, error) {
	b, err := in.ReadByte()
	if err != nil {
		return "", err
	}
	suffixLength := int(b)

	suffixBytes := make([]byte, suffixLength)
	_, err = in.Read(suffixBytes)
	if err != nil {
		return "", err
	}
	suffix := string(suffixBytes)
	if suffix != expectedSuffix {
		return "", fmt.Errorf("file mismatch, expected suffix=%s, got=%s", expectedSuffix, suffix)
	}
	return suffix, nil
}

// FooterLength Computes the length of a codec footer.
// Returns: length of the entire codec footer.
// See Also: writeFooter(IndexOutput)
func FooterLength() int {
	return 16
}

//const (
//	CODEC_MAGIC = 0x3fd76c17
//
//)

// WriteFooter Writes a codec footer, which records both a checksum algorithm ID and a checksum. This footer can be parsed and validated with checkFooter().
// CodecFooter --> Magic,AlgorithmID,Checksum
// Magic --> Uint32. This identifies the start of the footer. It is always -1071082520.
// AlgorithmID --> Uint32. This indicates the checksum algorithm used. Currently this is always 0, for zlib-crc32.
// Checksum --> Uint64. The actual checksum value for all previous bytes in the stream, including the bytes from Magic and AlgorithmID.
// Params: out – Output stream
// Throws: IOException – If there is an I/O error writing to the underlying medium.
func WriteFooter(out store.IndexOutput) error {
	if err := out.WriteUint32(nil, FOOTER_MAGIC); err != nil {
		return err
	}

	if err := out.WriteUint32(nil, 0); err != nil {
		return err
	}

	return writeCRC(out)
}

func writeCRC(output store.IndexOutput) error {
	value, err := output.GetChecksum()
	if err != nil {
		return err
	}

	// TODO: fix it
	//if (int(value) & 0xFFFFFFFF00000000) != 0 {
	//	return fmt.Errorf("illegal CRC-32 checksum: %d (resource=%+v)", value, output)
	//}
	return output.WriteUint64(nil, uint64(value))
}
