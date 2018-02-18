/* Copyright (C) 2018 David Jowett
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
	"flag"
        "fmt"
	"log"
)

func main() {
	mf := flag.String("mc", "", "Microcode in a binary file")
	msf := flag.String("mcs", "", "Microcode in a binary string file")
	memf := flag.String("m", "", "Memory in a binary file")
	memsf := flag.String("ms", "", "Memory in a binary stirng file")

	var mc []uint32
	var mem []uint16
	var syms []Symbol
	var err error

	mic := InitMic1()

	flag.Parse()

	if *mf != "" {
		log.Println("Reading binary microcode file:", *mf)
		mc, err = LoadBinaryMCFile(*mf)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d microcode instructions", len(mc))
		mic.LoadMC(mc)
	} else if *msf != "" {
		log.Println("Reading binary string microcode file:", *msf)
		mc, err = LoadBinaryStringMCFile(*msf)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d microcode instructions", len(mc))
		mic.LoadMC(mc)
	} else {
                fmt.Println("Error: no microcode file given!")
                flag.Usage()
		return
	}

	if *memf != "" {
		log.Println("Reading binary memory file:", *memf)
		mem, err = LoadBinaryMemFile(*memf)
		if err != nil {
			log.Fatal(err.Error())
		}
		syms = make([]Symbol, 0, 0)
		log.Printf("Loaded %d memory words", len(mem))
		log.Printf("Loaded %d memory symbols", len(syms))
		mic.LoadMem(mem)
		mic.MemSymbols = syms
	} else if *memsf != "" {
		log.Println("Reading binary string memory file:", *memsf)
		mem, syms, err = LoadBinaryStringMemFile(*memsf)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d memory words", len(mem))
		log.Printf("Loaded %d memory symbols", len(syms))
		mic.LoadMem(mem)
		mic.MemSymbols = syms
	} else {
		log.Println("no memory file given!")
	}
	g, err := initGui(mic)
	if err != nil {
		log.Panicln(err)
	}
	err = g.Run()
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("Completed %d cycles", mic.Cycles)
}
