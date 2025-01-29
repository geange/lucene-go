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

// FooterLength Computes the length of a codec footer.
// Returns: length of the entire codec footer.
// See Also: writeFooter(IndexOutput)
func FooterLength() int {
	return 16
}

// WriteFooter Writes a codec footer, which records both a checksum algorithm ID and a checksum. This footer can be parsed and validated with checkFooter().
// CodecFooter --> Magic,AlgorithmID,Checksum
// Magic --> Uint32. This identifies the start of the footer. It is always -1071082520.
// AlgorithmID --> Uint32. This indicates the checksum algorithm used. Currently this is always 0, for zlib-crc32.
// Checksum --> Uint64. The actual checksum value for all previous bytes in the stream, including the bytes from Magic and AlgorithmID.
// Params: out – Output stream
// Throws: IOException – If there is an I/O error writing to the underlying medium.
func WriteFooter(ctx context.Context, out store.IndexOutput) error {
	if err := out.WriteUint32(ctx, FOOTER_MAGIC); err != nil {
		return err
	}

	if err := out.WriteUint32(ctx, 0); err != nil {
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
