package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestSACPDecodeValid(t *testing.T) {
	orig := SACP_pack{
		ReceiverID: 1,
		SenderID:   2,
		Attribute:  0,
		Sequence:   0x1234,
		CommandSet: 0x56,
		CommandID:  0x78,
		Data:       []byte{0x9a, 0xbc},
	}

	encoded := orig.Encode()

	var got SACP_pack
	if err := got.Decode(encoded); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if got.ReceiverID != orig.ReceiverID ||
		got.SenderID != orig.SenderID ||
		got.Attribute != orig.Attribute ||
		got.Sequence != orig.Sequence ||
		got.CommandSet != orig.CommandSet ||
		got.CommandID != orig.CommandID ||
		string(got.Data) != string(orig.Data) {
		t.Fatalf("Decoded packet mismatch: %+v vs %+v", got, orig)
	}
}

func TestSACPDecodeInvalidHeader(t *testing.T) {
	p := SACP_pack{ReceiverID: 1, SenderID: 2, Attribute: 0, Sequence: 1}
	encoded := p.Encode()

	encoded[0] ^= 0xff // corrupt first header byte

	var got SACP_pack
	err := got.Decode(encoded)
	if err != errInvalidSACP {
		t.Fatalf("expected errInvalidSACP, got %v", err)
	}
}

// recordingConn is a minimal net.Conn implementation that captures written bytes.
type recordingConn struct{ bytes.Buffer }

func (r *recordingConn) Read(b []byte) (int, error)         { return 0, io.EOF }
func (r *recordingConn) Write(b []byte) (int, error)        { return r.Buffer.Write(b) }
func (r *recordingConn) Close() error                       { return nil }
func (r *recordingConn) LocalAddr() net.Addr                { return nil }
func (r *recordingConn) RemoteAddr() net.Addr               { return nil }
func (r *recordingConn) SetDeadline(t time.Time) error      { return nil }
func (r *recordingConn) SetReadDeadline(t time.Time) error  { return nil }
func (r *recordingConn) SetWriteDeadline(t time.Time) error { return nil }

func getPackageCountFromStartPacket(b []byte) (uint16, error) {
	var p SACP_pack
	if err := p.Decode(b); err != nil {
		return 0, err
	}
	if len(p.Data) < 2 {
		return 0, errInvalidSize
	}
	nameLen := binary.LittleEndian.Uint16(p.Data[:2])
	off := 2 + int(nameLen)
	if len(p.Data) < off+4+2 {
		return 0, errInvalidSize
	}
	off += 4
	pkgCount := binary.LittleEndian.Uint16(p.Data[off : off+2])
	return pkgCount, nil
}

func TestPackageCountExactMultiple(t *testing.T) {
	gcode := make([]byte, SACP_data_len*2)
	conn := &recordingConn{}
	_ = SACP_start_upload(conn, "f.gcode", gcode, time.Millisecond)
	pkgCount, err := getPackageCountFromStartPacket(conn.Bytes())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkgCount != 2 {
		t.Fatalf("expected package count 2, got %d", pkgCount)
	}
}

func TestPackageCountNonExactMultiple(t *testing.T) {
	gcode := make([]byte, SACP_data_len*2+123)
	conn := &recordingConn{}
	_ = SACP_start_upload(conn, "f.gcode", gcode, time.Millisecond)
	pkgCount, err := getPackageCountFromStartPacket(conn.Bytes())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkgCount != 3 {
		t.Fatalf("expected package count 3, got %d", pkgCount)
	}
}
