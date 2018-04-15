/* Copyright (C) 2019 David Jowett
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301  USA
 */
package main

import (
	"fmt"
)

type instruction struct {
	AMUX int8
	COND int8
	ALU  int8
	SH   int8
	MBR  int8
	MAR  int8
	RD   int8
	WR   int8
	ENC  int8
	C    int8
	B    int8
	A    int8
	ADDR uint8
	BR   bool
}

/* Unpacks an binary instruction into an instruction struct */
func Unpack(ins uint32) instruction {
	ret := instruction{}
	ret.AMUX = int8((ins & 0x80000000) >> 31)
	ret.COND = int8((ins & 0x60000000) >> 29)
	ret.ALU = int8((ins & 0x18000000) >> 27)
	ret.SH = int8((ins & 0x06000000) >> 25)
	ret.MBR = int8((ins & 0x01000000) >> 24)
	ret.MAR = int8((ins & 0x00800000) >> 23)
	ret.RD = int8((ins & 0x00400000) >> 22)
	ret.WR = int8((ins & 0x00200000) >> 21)
	ret.ENC = int8((ins & 0x00100000) >> 20)
	ret.C = int8((ins & 0x000F0000) >> 16)
	ret.B = int8((ins & 0x0000F000) >> 12)
	ret.A = int8((ins & 0x00000F00) >> 8)
	ret.ADDR = uint8(ins & 0x000000FF)
	ret.BR = false

	return ret
}

/* Returns a human readable format of the microcode */
func (i *instruction) ToString() string {
	s := ""
	areg := RegIdToNames[i.A]
	breg := RegIdToNames[i.B]
	creg := RegIdToNames[i.C]
	addr := fmt.Sprintf("%d", i.ADDR)

	/* Check if A bus is MBR */
	if i.AMUX == 1 {
		areg = "MBR"
	}

	/* Check MAR flag */
	if i.MAR == 1 {
		s += "mar := " + breg + "; "
	}

	/* Generate ALU strings if the output is going somewhere or is being used */
	if i.MBR == 1 || i.ENC == 1 || i.COND == 1 || i.COND == 2 {
		sp := ""
		ss := ""
		astr := ""
		cstr := ""
		switch i.SH {
		case 1:
			sp = "rshift("
			ss = ")"
		case 2:
			sp = "lshift("
			ss = ")"
		}
		switch i.ALU {
		case 0:
			astr = areg + " + " + breg
		case 1:
			astr = "band(" + areg + ", " + breg + ")"
		case 2:
			astr = areg
		case 3:
			astr = "not(" + areg + ")"

		}
		cstr = sp + astr + ss + "; "
		if i.MBR == 1 {
			s += "MBR := " + cstr
		}
		if i.ENC == 1 {
			s += creg + " := " + cstr
		}
		if i.ENC != 1 && i.MBR != 1 {
			s += "ALU := " + cstr
		}

	}

	/* Check read and write flags */
	if i.RD == 1 {
		s += "rd; "
	}
	if i.WR == 1 {
		s += "wr; "
	}

	/* generate goto section */
	switch i.COND {
	case 1:
		s += "if n goto " + addr + "; "
	case 2:
		s += "if z goto " + addr + "; "
	case 3:
		s += "goto " + addr + "; "
	}

	return s
}
