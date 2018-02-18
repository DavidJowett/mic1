/* Copyright (C) 2017 David Jowett
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

type mic1Alu struct {
	A, B, R uint16
	N, Z    int8
	S       int8
	F       int8
}

/* Calculates the outputs of the ALU */
func (m *mic1Alu) Calc() {
	switch m.F {
	case 0:
		m.R = m.A + m.B
	case 1:
		m.R = m.A & m.B
	case 2:
		m.R = m.A
	case 3:
		m.R = m.A ^ 0xFFFF
	}

	// set zero flag
	if m.R == 0 {
		m.Z = 1
	} else {
		m.Z = 0
	}

	// set negative flag
	if m.R&0x8000 != 0 {
		m.N = 1
	} else {
		m.N = 0
	}
	switch m.S {
	case 1:
		m.R = m.R >> 1
	case 2:
		m.R = m.R << 1
	}
}
