// Copyright 2018 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Multiboot info as defined in
// https://www.gnu.org/software/grub/manual/multiboot/multiboot.html#Boot-information-format
package multiboot

import (
	"bytes"
	"encoding/binary"

	"github.com/u-root/u-root/pkg/ubinary"
)

var sizeofInfo = uint32(binary.Size(Info{}))

type Flag uint32

const (
	flagInfoMemory Flag = 1 << iota
	flagInfoBootDev
	flagInfoCmdLine
	flagInfoMods
	flagInfoAoutSyms
	flagInfoElfSHDR
	flagInfoMemMap
	flagInfoDriveInfo
	flagInfoConfigTable
	flagInfoBootLoaderName
	flagInfoAPMTable
	flagInfoVideoInfo
	flagInfoFrameBuffer
)

// Info represents the Multiboot v1 info passed to the loaded kernel.
type Info struct {
	Flags    Flag
	MemLower uint32
	MemUpper uint32

	// BootDevice is not supported, always zero.
	BootDevice uint32

	CmdLine uint32

	ModsCount uint32
	ModsAddr  uint32

	// Syms is not supported, always zero array.
	Syms [4]uint32

	MmapLength uint32
	MmapAddr   uint32

	// Following fields except BootLoaderName are not suppoted yet,
	// the values are always set to zeros.

	DriversLength uint32
	DrivesrAddr   uint32

	ConfigTable uint32

	BootLoaderName uint32

	APMTable uint32

	VBEControlInfo  uint32
	VBEModeInfo     uint32
	VBEMode         uint16
	VBEInterfaceSeg uint16
	VBEInterfaceOff uint16
	VBEInterfaceLen uint16

	FramebufferAddr   uint16
	FramebufferPitch  uint16
	FramebufferWidth  uint32
	FramebufferHeight uint32
	FramebufferBPP    byte
	FramebufferType   byte
	ColorInfo         [6]byte
}

type infoWrapper struct {
	Info

	CmdLine        string
	BootLoaderName string
}

// marshal writes out the exact bytes of multiboot info
// expected by the kernel being loaded.
func (iw *infoWrapper) marshal(base uintptr) ([]byte, error) {
	offset := sizeofInfo + uint32(base)
	iw.Info.CmdLine = offset
	offset += uint32(len(iw.CmdLine)) + 1
	iw.Info.BootLoaderName = offset

	buf := bytes.Buffer{}
	if err := binary.Write(&buf, ubinary.NativeEndian, iw.Info); err != nil {
		return nil, err
	}

	for _, s := range []string{iw.CmdLine, iw.BootLoaderName} {
		if _, err := buf.WriteString(s); err != nil {
			return nil, err
		}
		if err := buf.WriteByte(0); err != nil {
			return nil, err
		}
	}

	size := (buf.Len() + 3) &^ 3
	_, err := buf.Write(bytes.Repeat([]byte{0}, size-buf.Len()))
	return buf.Bytes(), err
}

func (iw infoWrapper) size() (uint, error) {
	b, err := iw.marshal(0)
	return uint(len(b)), err
}
