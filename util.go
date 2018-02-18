package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func LoadBinaryMCFile(fp string) ([]uint32, error) {
	ret := make([]uint32, 0, 256)
	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return ret, err
	}
	if len(buff)%4 != 0 {
		return ret, errors.New(fmt.Sprintf("Binary microcode file, \"%s\" is not a multiple of 4 bytes in length", fp))
	}
	for i := 0; i < len(buff); i += 4 {
		var curWord uint32
		curWord |= uint32(buff[i]) << 24
		curWord |= uint32(buff[i+1]) << 16
		curWord |= uint32(buff[i+2]) << 8
		curWord |= uint32(buff[i+3])
		//log.Printf("%x", curWord)
		ret = append(ret, curWord)
	}
	log.Printf("Loaded %d bytes from a binary microcode file", len(buff))

	return ret, nil
}

func LoadBinaryStringMCFile(fp string) ([]uint32, error) {
	ret := make([]uint32, 0, 256)
	file, err := os.Open(fp)
	if err != nil {
		return ret, err
	}
	defer file.Close()
	s := bufio.NewScanner(file)

	for s.Scan() {
                var tmp uint32
                fmt.Sscanf(s.Text(),"%b", &tmp)
		ret = append(ret, tmp)
	}
	if err := s.Err(); err != nil {
		return ret, err
	}

	return ret, nil
}

func LoadBinaryMemFile(fp string) ([]uint16, error) {
	ret := make([]uint16, 0, 4096)
	buff, err := ioutil.ReadFile(fp)
	if err != nil {
		return ret, err
	}
	if len(buff)%2 != 0 {
		return ret, errors.New(fmt.Sprintf("Binary memory file, \"%s\" is not a multiple of 2 bytes in length", fp))
	}
	for i := 0; i < len(buff); i += 2 {
		var curWord uint16
		curWord |= uint16(buff[i]) << 8
		curWord |= uint16(buff[i+1])
		//log.Printf("%x", curWord)
		ret = append(ret, curWord)
	}

	return ret, nil
}

func LoadBinaryStringMemFile(fp string) ([]uint16, []Symbol, error) {
	ret := make([]uint16, 0, 4096)
	syms := make([]Symbol, 0, 0)
	file, err := os.Open(fp)
	if err != nil {
		return ret, syms, err
	}
	defer file.Close()
	s := bufio.NewScanner(file)

	for s.Scan() {
		line := s.Text()
		if line[0] != '#' {
                        var tmp uint16
                        fmt.Sscanf(line, "%b", &tmp)
                        ret = append(ret, tmp)
		} else {
			line := line[1:]
			ss := strings.Split(line, ":")
			if len(ss) == 2 {
				var val uint16
				name := strings.TrimSpace(ss[0])
				vals := strings.TrimSpace(ss[1])
				fmt.Sscanf(vals, "%d", &val)
				syms = append(syms, Symbol{Name: name, Val: val})
			}
		}
	}
	if err := s.Err(); err != nil {
		return ret, syms, err
	}

	return ret, syms, nil
}
