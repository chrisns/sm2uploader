/*
Author: https://github.com/kanocz
*/
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const (
	SACP_data_len = 60 * 1024 // just as defined in original python code
)

var (
	errInvalidSACP     = errors.New("data doesn't look like SACP packet")
	errInvalidSACPVer  = errors.New("SACP version missmatch")
	errInvalidChksum   = errors.New("SACP checksum doesn't match data")
	errInvalidSize     = errors.New("SACP package is too short")
	errTimeoutExceeded = errors.New("timeout exceeded")
)

type SACP_pack struct {
	// 0xAA byte
	// 0x55 byte
	// DataLength uint16
	// 0x01 (SACP version)
	ReceiverID byte
	// head_chksum byte
	SenderID   byte
	Attribute  byte
	Sequence   uint16
	CommandSet byte
	CommandID  byte
	Data       []byte
	// data_checksum uint16
}

func (sacp SACP_pack) Encode() []byte {
	result := make([]byte, 15+len(sacp.Data))

	result[0] = 0xAA
	result[1] = 0x55
	binary.LittleEndian.PutUint16(result[2:4], uint16(len(sacp.Data)+6+2))
	result[4] = 0x01
	result[5] = sacp.ReceiverID
	result[6] = sacp.headChksum(result[:6])
	result[7] = sacp.SenderID
	result[8] = sacp.Attribute
	binary.LittleEndian.PutUint16(result[9:11], sacp.Sequence)
	result[11] = sacp.CommandSet
	result[12] = sacp.CommandID

	if len(sacp.Data) > 0 { // this also include check on nil
		copy(result[13:], sacp.Data)
	}

	binary.LittleEndian.PutUint16(result[len(result)-2:], sacp.U16Chksum(result[7:], uint16(len(sacp.Data))+6))

	return result[:]
}

func (sacp *SACP_pack) Decode(data []byte) error {
	if len(data) < 13 {
		return errInvalidSize
	}
	// ensure the packet starts with the expected 0xAA55 header bytes
	if data[0] != 0xAA || data[1] != 0x55 {
		return errInvalidSACP
	}
	dataLen := binary.LittleEndian.Uint16(data[2:4])
	if dataLen != uint16(len(data)-7) {
		return errInvalidSize
	}
	if data[4] != 0x01 {
		return errInvalidSACPVer
	}
	if sacp.headChksum(data[:6]) != data[6] {
		return errInvalidChksum
	}
	if binary.LittleEndian.Uint16(data[len(data)-2:]) != sacp.U16Chksum(data[7:], dataLen-2) {
		return errInvalidChksum
	}

	sacp.ReceiverID = data[5]
	sacp.SenderID = data[7]
	sacp.Attribute = data[8]
	sacp.Sequence = binary.LittleEndian.Uint16(data[9:11])
	sacp.CommandSet = data[11]
	sacp.CommandID = data[12]
	sacp.Data = data[13 : len(data)-2]

	return nil
}

func (sacp *SACP_pack) headChksum(data []byte) byte {
	crc := byte(0)
	poly := byte(7)
	for i := 0; i < len(data); i++ {
		for j := 0; j < 8; j++ {
			bit := ((data[i] & 0xff) >> (7 - j) & 0x01) == 1
			c07 := (crc >> 7 & 0x01) == 1
			crc = crc << 1
			if (!c07 && bit) || (c07 && !bit) {
				crc ^= poly
			}
		}
	}
	crc = crc & 0xff
	return crc
}

func (sacp *SACP_pack) U16Chksum(package_data []byte, length uint16) uint16 {
	check_num := uint32(0)
	if length > 0 {
		for i := 0; i < int(length-1); i += 2 {
			check_num += uint32((uint32(package_data[i])&0xff)<<8 | uint32(package_data[i+1])&0xff)
			check_num &= 0xffffffff
		}
		if length%2 != 0 {
			check_num += uint32(package_data[length-1])
		}
	}
	for check_num > 0xFFFF {
		check_num = ((check_num >> 16) & 0xFFFF) + (check_num & 0xFFFF)
	}
	check_num = ^check_num
	return uint16(check_num & 0xFFFF)
}

func writeSACPstring(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint16(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func writeSACPbytes(w io.Writer, s []byte) error {
	if err := binary.Write(w, binary.LittleEndian, uint16(len(s))); err != nil {
		return err
	}
	_, err := w.Write(s)
	return err
}

func writeLE[T any](w io.Writer, u T) {
	binary.Write(w, binary.LittleEndian, u)
}

func SACP_connect(ip string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.Dial("tcp4", ip+":8888")
	if err != nil {
		// log.Printf("Error connecting to %s: %v", ip, err)
		return nil, err
	}

	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write(SACP_pack{
		ReceiverID: 2,
		SenderID:   0,
		Attribute:  0,
		Sequence:   1,
		CommandSet: 0x01,
		CommandID:  0x05,
		Data: []byte{
			11, 0, 's', 'm', '2', 'u', 'p', 'l', 'o', 'a', 'd', 'e', 'r',
			0, 0,
			0, 0,
		},
	}.Encode())

	if err != nil {
		// log.Println("Error writing \"hello\": ", err)
		conn.Close()
		return nil, err
	}

	for {
		p, err := SACP_read(conn, timeout)
		if err != nil || p == nil {
			// log.Println("Error reading \"hello\" response: ", err)
			conn.Close()
			return nil, err
		}

		if Debug {
			log.Printf("-- SACP_connect got:\n%v", p)
		}

		if p.CommandSet == 1 && p.CommandID == 5 {
			break
		}
	}

	if Debug {
		log.Println("-- Connected to printer")
	}

	return conn, nil
}

func SACP_read(conn net.Conn, timeout time.Duration) (*SACP_pack, error) {
	var buf [SACP_data_len + 15]byte

	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)

	n, err := conn.Read(buf[:4])
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, errInvalidSize
	}

	dataLen := binary.LittleEndian.Uint16(buf[2:4])
	n, err = conn.Read(buf[4 : dataLen+7])
	if err != nil {
		return nil, err
	}
	if n != int(dataLen+3) {
		return nil, errInvalidSize
	}

	var sacp SACP_pack
	err = sacp.Decode(buf[:dataLen+7])

	return &sacp, err
}

var sequence uint16 = 2

func SACP_set_tool_temperature(conn net.Conn, tool_id uint8, temperature uint16, timeout time.Duration) error {
	data := bytes.Buffer{}

	data.WriteByte(0x08)

	// Tool ID, starting at 0x00
	data.WriteByte(tool_id)

	// Temperature
	writeLE(&data, uint16(temperature))

	return SACP_send_command(conn, 0x10, 0x02, data, timeout)
}

func SACP_set_bed_temperature(conn net.Conn, tool_id uint8, temperature uint16, timeout time.Duration) error {
	data := bytes.Buffer{}

	data.WriteByte(0x05)

	// Tool ID, starting at 0x00
	data.WriteByte(tool_id)

	// Temperature
	writeLE(&data, uint16(temperature))

	return SACP_send_command(conn, 0x14, 0x02, data, timeout)
}

func SACP_home(conn net.Conn, timeout time.Duration) error {
	data := bytes.Buffer{}
	data.WriteByte(0x00)

	// 0x31 is also used when homing in Luban???
	// 0x35 homes everything
	return SACP_send_command(conn, 0x01, 0x35, data, timeout)
}

func SACP_send_command(conn net.Conn, command_set uint8, command_id uint8, data bytes.Buffer, timeout time.Duration) error {

	sequence++

	start := time.Now()
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write(SACP_pack{
		ReceiverID: 1,
		SenderID:   0,
		Attribute:  0,
		Sequence:   sequence,
		CommandSet: command_set,
		CommandID:  command_id,
		Data:       data.Bytes(),
	}.Encode())

	if err != nil {
		return err
	}

	if Debug {
		log.Printf("-- Sequence: %d Sent GCode: %x", sequence, data.Bytes())
	}

	for {
		remaining := timeout - time.Since(start)
		if remaining <= 0 {
			return errTimeoutExceeded
		}
		conn.SetReadDeadline(time.Now().Add(remaining))
		p, err := SACP_read(conn, remaining)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() && time.Since(start) >= timeout {
				return errTimeoutExceeded
			}
			return err
		}

		if Debug {
			log.Printf("-- Got reply from printer: %v", p)
		}

		if p.Sequence == sequence && p.CommandSet == command_set && p.CommandID == command_id {
			if len(p.Data) == 1 && p.Data[0] == 0 {
				return nil
			}
		}
	}
}

func SACP_start_upload(conn net.Conn, filename string, gcode []byte, timeout time.Duration) error {
	// prepare data for upload begin packet
	package_count := uint16((len(gcode) + SACP_data_len - 1) / SACP_data_len)
	md5hash := md5.Sum(gcode)

	data := bytes.Buffer{}

	if err := writeSACPstring(&data, filename); err != nil {
		return err
	}
	writeLE(&data, uint32(len(gcode)))
	writeLE(&data, package_count)
	if err := writeSACPstring(&data, hex.EncodeToString(md5hash[:])); err != nil {
		return err
	}

	if Debug {
		log.Println("-- Starting upload ...")
	}

	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write(SACP_pack{
		ReceiverID: 2,
		SenderID:   0,
		Attribute:  0,
		Sequence:   1,
		CommandSet: 0xb0,
		CommandID:  0x00,
		Data:       data.Bytes(),
	}.Encode())

	if err != nil {
		return err
	}

	for {
		// always receive packet, then send response
		conn.SetReadDeadline(time.Now().Add(timeout))
		p, err := SACP_read(conn, timeout)
		if err != nil {
			return err
		}

		if p == nil {
			return errInvalidSize
		}

		if Debug {
			log.Printf("-- Got reply from printer: %v", p)
		}

		switch {
		case p.CommandSet == 0xb0 && p.CommandID == 0:
			// just ignore, don't know that this message means :)
		case p.CommandSet == 0xb0 && p.CommandID == 1:
			// sending next chunk
			if len(p.Data) < 4 {
				return errInvalidSize
			}
			md5_len := binary.LittleEndian.Uint16(p.Data[:2])
			if len(p.Data) < 2+int(md5_len)+2 {
				return errInvalidSize
			}

			pkgRequested := binary.LittleEndian.Uint16(p.Data[2+md5_len : 2+md5_len+2])
			var pkgData []byte

			if pkgRequested == package_count-1 { // last package
				pkgData = gcode[SACP_data_len*int(pkgRequested):]
			} else { // regular package
				pkgData = gcode[SACP_data_len*int(pkgRequested) : SACP_data_len*int(pkgRequested+1)]
			}

			data := bytes.Buffer{}
			data.WriteByte(0)
			if err := writeSACPstring(&data, hex.EncodeToString(md5hash[:])); err != nil {
				return err
			}
			writeLE(&data, pkgRequested)
			if err := writeSACPbytes(&data, pkgData); err != nil {
				return err
			}

			// log.Printf("  sending package %d of %d", pkgRequested+1, package_count)
			perc := float64(pkgRequested+1) / float64(package_count) * 100.0
			log.Printf("  - SACP sending %.1f%%", perc)

			conn.SetWriteDeadline(time.Now().Add(timeout))
			_, err := conn.Write(SACP_pack{
				ReceiverID: 2,
				SenderID:   0,
				Attribute:  1,
				Sequence:   p.Sequence,
				CommandSet: 0xb0,
				CommandID:  0x01,
				Data:       data.Bytes(),
			}.Encode())

			if err != nil {
				return err
			}

		case p.CommandSet == 0xb0 && p.CommandID == 2:
			// send finished!!!
			if len(p.Data) == 1 && p.Data[0] == 0 {

				if Debug {
					log.Print("-- Upload finished")
				}

				if err := SACP_disconnect(conn, timeout); err != nil {
					return err
				}

				return nil // everything is ok!
			}

			log.Print("Unable to process b0/02 with invalid data", p.Data)

		default:
			continue
		}

	}
}

func SACP_disconnect(conn net.Conn, timeout time.Duration) (err error) {
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write(SACP_pack{
		ReceiverID: 2,
		SenderID:   0,
		Attribute:  0,
		Sequence:   1,
		CommandSet: 0x01,
		CommandID:  0x06,
		Data:       []byte{},
	}.Encode())

	return err
}
