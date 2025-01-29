package codecs

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/store"
)

const (
	CODEC_MAGIC  = 0x3fd76c17
	FOOTER_MAGIC = 0xc02893e8
)

// CheckIndexHeaderID
// Expert: just reads and verifies the object ID of an index header
func CheckIndexHeaderID(in store.DataInput, expectedID []byte) ([]byte, error) {
	id := make([]byte, len(expectedID))
	_, err := in.Read(id)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(id, expectedID) {
		return nil, fmt.Errorf("file mismatch, expected id=%s", string(expectedID))
	}
	return id, nil
}

func CheckIndexHeader(ctx context.Context, in store.DataInput, codec string, minVersion, maxVersion int,
	expectedID []byte, expectedSuffix string) (int, error) {

	version, err := CheckHeader(ctx, in, codec, minVersion, maxVersion)
	if err != nil {
		return 0, err
	}
	if _, err := CheckIndexHeaderID(in, expectedID); err != nil {
		return 0, err
	}
	if _, err := CheckIndexHeaderSuffix(in, expectedSuffix); err != nil {
		return 0, err
	}
	return version, nil
}

// RetrieveChecksum
// Returns (but does not validate) the checksum previously written by checkFooter.
func RetrieveChecksum(ctx context.Context, in store.IndexInput) (int64, error) {
	if in.Length() < footerLength() {
		return 0, errors.New("misplaced codec footer (file truncated?)")
	}
	if _, err := in.Seek(in.Length()-footerLength(), 0); err != nil {
		return 0, err
	}
	if err := ValidateFooter(ctx, in); err != nil {
		return 0, err
	}
	return ReadCRC(ctx, in)
}

func footerLength() int64 {
	return 16
}

func ValidateFooter(ctx context.Context, in store.IndexInput) error {
	remaining := in.Length() - in.GetFilePointer()
	expected := footerLength()
	if remaining < expected {
		return errors.New("misplaced codec footer (file truncated?)")
	} else if remaining > expected {
		return errors.New("misplaced codec footer (file extended?)")
	}

	magic, err := in.ReadUint32(ctx)
	if err != nil {
		return err
	}
	if magic != FOOTER_MAGIC {
		return errors.New("codec footer mismatch (file truncated?)")
	}

	algorithmID, err := in.ReadUint32(ctx)
	if err != nil {
		return err
	}
	if algorithmID != 0 {
		return errors.New("codec footer mismatch: unknown algorithmID")
	}
	return nil
}

func CheckFooter(ctx context.Context, in store.ChecksumIndexInput) (int64, error) {
	if err := ValidateFooter(ctx, in); err != nil {
		return 0, err
	}
	actualChecksum := in.GetChecksum()
	expectedChecksum, err := ReadCRC(ctx, in)
	if err != nil {
		return 0, err
	}
	if uint32(expectedChecksum) != actualChecksum {
		return 0, errors.New("checksum failed (hardware problem?)")
	}
	return int64(actualChecksum), nil
}

func ChecksumEntireFile(ctx context.Context, input store.IndexInput) (int64, error) {
	clone := input.Clone().(store.IndexInput)
	if _, err := clone.Seek(0, 0); err != nil {
		return 0, err
	}
	in := store.NewBufferedChecksumIndexInput(clone)
	if in.Length() < footerLength() {

	}
	if _, err := in.Seek(in.Length()-footerLength(), 0); err != nil {
		return 0, err
	}
	return CheckFooter(ctx, in)
}

// ReadCRC
// Reads CRC32 value as a 64-bit long from the input.
func ReadCRC(ctx context.Context, input store.IndexInput) (int64, error) {
	value, err := input.ReadUint64(ctx)
	if err != nil {
		return 0, err
	}

	if value&0xFFFFFFFF00000000 != 0 {
		return 0, errors.New("illegal CRC-32 checksum")
	}
	return int64(value), nil
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
