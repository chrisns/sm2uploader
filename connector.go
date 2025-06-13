package main

import (
	"errors"
	"io"
	"net"
	"time"
)

const (
	FILE_SIZE_MIN = 1
	FILE_SIZE_MAX = 2 << 30 // 2GB
)

var (
	errFileEmpty      = errors.New("File is empty.")
	errFileTooLarge   = errors.New("File is too large.")
	ErrNotImplemented = errors.New("not implemented")
)

type Payload struct {
	File  io.Reader
	Name  string
	Size  int64
	Print bool
}

func (p *Payload) SetName(name string) {
	p.Name = normalizedFilename(name)
}

func (p *Payload) ReadableSize() string {
	return humanReadableSize(p.Size)
}

func (p *Payload) GetContent(nofix bool) (cont []byte, err error) {
	if nofix || !p.ShouldBeFix() {
		cont, err = io.ReadAll(p.File)
	} else {
		cont, err = postProcess(p.File)
		p.Size = int64(len(cont))
	}
	return cont, err
}

func (p *Payload) ShouldBeFix() bool {
	return shouldBeFix(p.Name)
}

func NewPayload(file io.Reader, name string, size int64, print bool) *Payload {
	return &Payload{
		File:  file,
		Name:  normalizedFilename(name),
		Size:  size,
		Print: print,
	}
}

type connector struct {
	handlers []Handler
}

type Handler interface {
	Ping(*Printer) bool
	Connect() error
	Disconnect() error
	Upload(*Payload) error
	SetToolTemperature(int, int) error
	SetBedTemperature(int, int) error
	Home() error
}

func (c *connector) RegisterHandler(h Handler) {
	c.handlers = append(c.handlers, h)
}

// Upload to upload a file to a printer
func (c *connector) Upload(printer *Printer, payload *Payload) error {
	// Iterate through all handlers
	for _, h := range c.handlers {
		// Check if handler can ping the printer
		if h.Ping(printer) {
			// Connect to the printer
			if err := h.Connect(); err != nil {
				return err
			}
			defer h.Disconnect()

			if payload.Size > FILE_SIZE_MAX {
				return errFileTooLarge
			}
			if payload.Size < FILE_SIZE_MIN {
				return errFileEmpty
			}
			// Upload the file to the printer
			if err := h.Upload(payload); err != nil {
				return err
			}

			// Return nil if successful
			return nil
		}
	}
	// Return error if printer is not available
	return errors.New("Printer " + printer.IP + " is not available.")
}

func (c *connector) PreHeatCommands(printer *Printer, tool_1_temperature int, tool_2_temperature int, bed_temperature int, home bool) error {
	// Iterate through all handlers
	for _, h := range c.handlers {
		// Check if handler can ping the printer
		if h.Ping(printer) {
			// Connect to the printer
			if err := h.Connect(); err != nil {
				return err
			}
			defer h.Disconnect()

			// Send the GCode command to the printer
			if tool_1_temperature > 0 {
				if err := h.SetToolTemperature(0, tool_1_temperature); err != nil && !errors.Is(err, ErrNotImplemented) {
					return err
				}
			}
			if tool_2_temperature > 0 {
				if err := h.SetToolTemperature(1, tool_2_temperature); err != nil && !errors.Is(err, ErrNotImplemented) {
					return err
				}
			}
			if bed_temperature > 0 {
				if err := h.SetBedTemperature(0, bed_temperature); err != nil && !errors.Is(err, ErrNotImplemented) {
					return err
				}
				if err := h.SetBedTemperature(1, bed_temperature); err != nil && !errors.Is(err, ErrNotImplemented) {
					return err
				}
			}
			if home {
				if err := h.Home(); err != nil && !errors.Is(err, ErrNotImplemented) {
					return err
				}
			}
			// Return nil if successful
			return nil
		}
	}
	// Return error if printer is not available
	return errors.New("Printer " + printer.IP + " is not available.")
}

var Connector = &connector{}

// ping the printer to see if it is available
func ping(ip string, port string, timeout int) bool {
	if timeout <= 0 {
		timeout = 2
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), time.Second*time.Duration(timeout))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
