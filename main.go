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
	var mr func(mic *mic1) error
	var mcr func(mic *mic1) error

	mic := InitMic1()

	flag.Parse()

	if *mf != "" {
		fname := *mf
		log.Println("Reading binary microcode file:", fname)
		mc, err = LoadBinaryMCFile(*mf)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d microcode instructions", len(mc))
		mic.LoadMC(mc)
		mcr = func(mic *mic1) error {
			mc, err := LoadBinaryMCFile(fname)
			if err != nil {
				return err
			}
			mic.LoadMC(mc)
			return nil
		}
	} else if *msf != "" {
		fname := *msf
		log.Println("Reading binary string microcode file:", fname)
		mc, err = LoadBinaryStringMCFile(fname)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d microcode instructions", len(mc))
		mic.LoadMC(mc)
		mcr = func(mic *mic1) error {
			mc, err := LoadBinaryStringMCFile(fname)
			if err != nil {
				return err
			}
			mic.LoadMC(mc)
			return nil
		}
	} else {
		fmt.Println("Error: no microcode file given!")
		flag.Usage()
		return
	}

	if *memf != "" {
		fname := *memf
		log.Println("Reading binary memory file:", fname)
		mem, err = LoadBinaryMemFile(fname)
		if err != nil {
			log.Fatal(err.Error())
		}
		syms = make([]Symbol, 0, 0)
		log.Printf("Loaded %d memory words", len(mem))
		log.Printf("Loaded %d memory symbols", len(syms))
		mic.LoadMem(mem)
		mic.MemSymbols = syms
		mr = func(mic *mic1) error {
			mem, err := LoadBinaryMemFile(fname)
			syms := make([]Symbol, 0, 0)
			if err != nil {
				return err
			}
			mic.LoadMem(mem)
			mic.MemSymbols = syms
			return nil
		}
	} else if *memsf != "" {
		fname := *memsf
		log.Println("Reading binary string memory file:", fname)
		mem, syms, err = LoadBinaryStringMemFile(fname)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Loaded %d memory words", len(mem))
		log.Printf("Loaded %d memory symbols", len(syms))
		mic.LoadMem(mem)
		mic.MemSymbols = syms
		mr = func(mic *mic1) error {
			mem, syms, err := LoadBinaryStringMemFile(fname)
			if err != nil {
				log.Print(err.Error())
				return err
			}
			mic.LoadMem(mem)
			mic.MemSymbols = syms

			return nil
		}
	} else {
		log.Println("no memory file given!")
	}
	g, err := initGui(mic)
	if err != nil {
		log.Panicln(err)
	}
	g.MR = mr
	g.MCR = mcr
	err = g.Run()
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("Completed %d cycles", mic.Cycles)
}
