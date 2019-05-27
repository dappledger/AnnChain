package utils

import (
	"encoding/binary"
	"errors"
	"io"
)

func uvarintSize(i uint64) int {
	if i == 0 {
		return 0
	}
	if i < 1<<8 {
		return 1
	}
	if i < 1<<16 {
		return 2
	}
	if i < 1<<24 {
		return 3
	}
	if i < 1<<32 {
		return 4
	}
	if i < 1<<40 {
		return 5
	}
	if i < 1<<48 {
		return 6
	}
	if i < 1<<56 {
		return 7
	}
	return 8
}

func WriteVarint(w io.Writer, i int) error {
	var negate = false
	if i < 0 {
		negate = true
		i = -i
	}
	var size = uvarintSize(uint64(i))
	var err error
	if negate {
		// e.g. 0xF1 for a single negative byte
		err = BinWrite(w, uint8(size+0xF0))
	} else {
		err = BinWrite(w, uint8(size))
	}
	if err != nil {
		return err
	}
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		err = BinWrite(w, buf[(8-size):])
	}
	return err
}

func ReadByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	return buf[0], err
}

func ReadUint8(r io.Reader) (uint8, error) {
	bys, err := ReadByte(r)
	if err != nil {
		return 0, err
	}
	return uint8(bys), nil
}

func ReadVarint(r io.Reader) (int, error) {
	ui8, err := ReadUint8(r)
	if err != nil {
		return 0, err
	}
	var negate = false
	if (ui8 >> 4) == 0xF {
		negate = true
		ui8 = ui8 & 0x0F
	}
	if ui8 > 8 {
		return 0, errors.New("Varint overflow")
	}
	if ui8 == 0 {
		if negate {
			err = errors.New("Varint does not allow negative zero")
		}
		return 0, err
	}
	var buf [8]byte
	_, err = io.ReadFull(r, buf[(8-ui8):])
	if err != nil {
		return 0, err
	}
	var i = int(binary.BigEndian.Uint64(buf[:]))
	if negate {
		return -i, nil
	}
	return i, nil
}
