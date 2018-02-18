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
