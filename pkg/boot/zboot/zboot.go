package zboot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/u-root/uio/uio"
)

type Header struct {
	Magic       uint32
	Type        uint32
	Offset      uint32
	Size        uint32
	Reserved    [2]uint32
	Compression [4]byte
}

const (
	magic      = 0x00005a4d
	typeZImage = 0x676d697a // "zimg"
)

func ExtractPayload(k io.ReaderAt) (io.ReaderAt, error) {
	header := Header{}
	if err := binary.Read(uio.Reader(k), binary.LittleEndian, &header); err != nil {
		return k, nil
	}
	fmt.Printf("header: %+v", header)

	if header.Magic != magic {
		return k, fmt.Errorf("invalid magic number: %x", header.Magic)
	}

	if header.Type != typeZImage {
		return k, fmt.Errorf("invalid type: %x", header.Type)
	}

	buf := make([]byte, header.Size)
	n, err := k.ReadAt(buf, int64(header.Offset))
	if err != nil {
		return k, fmt.Errorf("read payload: %v", err)
	}
	if n != int(header.Size) {
		return k, fmt.Errorf("read %d bytes, expected %d", n, header.Size)
	}
	fmt.Printf("read %d bytes\n", n)

	// decompressed := make([]byte, 30<<20)
	switch string(header.Compression[:]) {
	case "zstd":
		fmt.Printf("found zstd\n")

		r, err := zstd.NewReader(bytes.NewReader(buf))
		if err != nil {
			return k, fmt.Errorf("zstd.NewReader: %v", err)
		}
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return k, fmt.Errorf("zstd.Read: %v", err)
		}
		fmt.Printf("decompressed %d bytes\n", len(decompressed))
		return bytes.NewReader(decompressed), nil
	default:
		return k, fmt.Errorf("unsupported compression type: %s", string(header.Compression[:]))
	}
}
