package main

import (
	"fmt"
)

/* A cli to provide a similar interface to the standard UML Mic-1 emulator (https://github.com/jeapostrophe/mic1) interface */

type CLI struct {
	Mic *mic1
}

/* Reads a line from stdin and returns it with a newline on the end */
func ReadLine() string {
	var buff string
	fmt.Scanf("%s", &buff)
	return buff + "\n"
}

func ReadStdin(out chan<- string) {
	for {
		in := ReadLine()
		out <- in
	}
}

func (c *CLI) Run() {
	var input rune
	var addr int16
	stdin := make(chan string)
	run := true
	c.DisplayState()
	go ReadStdin(stdin)
	for run {
		fmt.Println("Type address to view memory, [q]uit, [c]ontinue, <Enter> for symbol table:")
		in := <-stdin
		read, _ := fmt.Sscanf(in, "%d", &addr)
		if read == 0 {
			fmt.Sscanf(in, "%c", &input)
		} else {
			input = '1'
		}
		switch input {
		case 'c':
			/* continue running until next breakpoint */
			c.Mic.DesiredState = RUN
			go c.Mic.Run()
			wait := true
			for wait {
				select {
				case output := <-c.Mic.Output:
					fmt.Print(output)
				case newState := <-c.Mic.StateChanges:
					if newState == HALT {
						wait = false
					}
				case in := <-stdin:
					for _, v := range in {
						c.Mic.Input <- string(v)
					}
				}
			}
			/* Flush the output channel */
			wait = true
			for wait {
				select {
				case output := <-c.Mic.Output:
					fmt.Print(output)
				default:
					wait = false
				}
			}
			fmt.Println("")
			c.DisplayState()
		case 'q':
			/* exit the emulator */
			run = false
		case '1':
			/* print out memory */
			val := c.Mic.Memory[addr]
			fmt.Printf("%6d : %016b %5d %5d\n", addr, val, val, val)
			wminput := true
			for wminput {
				fmt.Println("Type <Enter> to continue debugging, q to quit, f for forward range,  b for backward range")
				in := <-stdin
				fmt.Sscanf(in, "%c", &input)
				switch input {
				case 'q':
					wminput = false
					run = false
				case '\n':
					wminput = false
				case 'f':
					count := 0
					fmt.Print("Number of locations to dump: ")
					in := <-stdin
					fmt.Sscanf(in, "%d", &count)
					for i := addr; int(i) <= (int(addr) + count); i++ {
						val := c.Mic.Memory[i]
						fmt.Printf("%6d : %016b %5d %5d\n", i, val, val, val)
					}
					wminput = false
				case 'b':
					count := 0
					fmt.Print("Number of locations to dump: ")
					in := <-stdin
					fmt.Sscanf(in, "%d", &count)
					i := int(addr) - count
					if i < 0 {
						i = 0
					}
					for ; i <= int(addr); i++ {
						val := c.Mic.Memory[i]
						fmt.Printf("%6d : %016b %5d %5d\n", i, val, val, val)
					}
					wminput = false
				}
			}
		case '\n':
			/* print the symbol table */
			for _, v := range c.Mic.MemSymbols {
				fmt.Printf("%-24s : %d\n", v.Name, v.Val)
			}
		}
	}
}

func (c *CLI) DisplayState() {
	for i, v := range c.Mic.Registers {
		fmt.Printf("%6s : %016b %5d %5d\n", RegIdToNames[i], v, v, int16(v))
	}
	fmt.Printf("\n")
	fmt.Printf("%6s : %d\n", "MPC", c.Mic.MPC)
	fmt.Printf("%6s : %d\n", "Cycles", c.Mic.Cycles)
}
