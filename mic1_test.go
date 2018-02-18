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

import (
	"testing"
        "fmt"
)

func TestUnpack(t *testing.T) {
        var tmp uint32
        fmt.Sscanf("00000000110000000000000000000000", "%b", &tmp)
	ins := Unpack(tmp)
	if ins.RD != 1 {
		t.Errorf("Unpacking instruction failed!")
	}
	if ins.MAR != 1 {
		t.Errorf("Unpacking instruction failed!")
	}
}
