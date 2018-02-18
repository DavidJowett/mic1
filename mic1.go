package main

import (
	"sync"
)

var RegIdToNames = []string{"PC", "AC", "SP", "IR", "TIR", "0", "+1", "-1", "AMASK", "SMASK", "A", "B", "C", "D", "E", "F"}

const (
	REG_PC = iota
	REG_AC
	REG_SP
	REG_IR
	REG_TIR
	REG_0
	REG_1
	REG_NEG1
	REG_AMASK
	REG_SMASK
	REG_A
	REG_B
	REG_C
	REG_D
	REG_E
	REG_F
)

const (
	HALT = iota
	RUN
)

type mic1 struct {
	Registers [16]uint16
	Memory    [4096]uint16
	MAR       uint16
	MBR       uint16
	ALU       *mic1Alu
	MPC       uint8
	MCC       [256]*instruction

	RD int8
	WR int8

	// MBRS staging
	MBRS uint16
	// MAR staging
	MARS uint16

	State        int8
	DesiredState int8
	Cycles       uint64

	MemSymbols []Symbol

	StateLock     *sync.Mutex
	RegistersLock *sync.Mutex

	StateChange *sync.Cond

	// Breakpoints for PC and MPC
	MPCBR []uint8
	PCBR  []uint16
}

type Symbol struct {
	Name string
	Val  uint16
}

func InitMic1() *mic1 {
	m := &mic1{ALU: &mic1Alu{}, StateLock: &sync.Mutex{}, RegistersLock: &sync.Mutex{}, Cycles: 0, MemSymbols: make([]Symbol, 0)}
	m.StateChange = sync.NewCond(m.StateLock)
	m.DesiredState = HALT
	m.Registers[REG_PC] = 0
	m.Registers[REG_SP] = 4095
	m.Registers[REG_0] = 0
	m.Registers[REG_1] = 1
	m.Registers[REG_NEG1] = 0xFFFF
	m.Registers[REG_AMASK] = 0x0FFF
	m.Registers[REG_SMASK] = 0x00FF

	m.MARS = 0xFFFF

	return m
}

func (m *mic1) ZeroMem() {
	for i := 0; i < len(m.Memory); i++ {
		m.Memory[i] = 0
	}
}

func (m *mic1) ZeroMC() {
	for i := 0; i < len(m.MCC); i++ {
		m.MCC[i] = nil
	}
}

func (m *mic1) Reset() {
	m.DesiredState = HALT
	m.RegistersLock.Lock()
	defer m.RegistersLock.Unlock()
	m.Registers[REG_PC] = 0
	m.Registers[REG_AC] = 0
	m.Registers[REG_SP] = 4095
	m.Registers[REG_IR] = 0
	m.Registers[REG_TIR] = 0
	m.Registers[REG_0] = 0
	m.Registers[REG_1] = 1
	m.Registers[REG_NEG1] = 0xFFFF
	m.Registers[REG_AMASK] = 0x0FFF
	m.Registers[REG_SMASK] = 0x00FF
	m.Registers[REG_A] = 0
	m.Registers[REG_B] = 0
	m.Registers[REG_C] = 0
	m.Registers[REG_D] = 0
	m.Registers[REG_E] = 0
	m.Registers[REG_F] = 0

	m.MARS = 0xFFFF
	m.MBR = 0
	m.MPC = 0
	m.Cycles = 0
}

func (m *mic1) AddMPCBR(br uint8) {
	m.MPCBR = append(m.MPCBR, br)
}

func (m *mic1) LoadMC(mc []uint32) {
	for i, v := range mc {
		ins := Unpack(v)
		m.MCC[i] = &ins
	}
}

func (m *mic1) LoadMem(mem []uint16) {
	for i, v := range mem {
		m.Memory[i] = v
	}
}

func (m *mic1) Run() {
	if m.DesiredState == RUN {
		m.StateChange.L.Lock()
		m.State = RUN
		m.StateChange.Broadcast()
		m.StateChange.L.Unlock()
	}
	for {
		// check if we should run
		shouldRun := m.DesiredState
		if shouldRun == RUN {
			m.Step()
		} else {
			m.StateChange.L.Lock()
			m.State = HALT
			m.StateChange.Broadcast()
			m.StateChange.L.Unlock()
			m.StateChange.Broadcast()
			break
		}
	}
}

/* Executes one microcode cycle */
func (m *mic1) Step() {
	//log.Printf("Read MC instruction at %d", m.MPC)
	m.RegistersLock.Lock()
	defer m.RegistersLock.Unlock()
	ins := m.MCC[m.MPC]
	//log.Printf("%d, %d", ins.RD, ins.WR)
	//ins.RD = 1
	//ins.WR = 1
	// Set ALU's B input
	m.ALU.B = m.Registers[ins.B]
	// Set ALU's A input
	if ins.AMUX == 1 {
		m.ALU.A = m.MBR
	} else {
		m.ALU.A = m.Registers[ins.A]
	}
	// Set ALU function
	m.ALU.F = ins.ALU
	// Set ALU shifter
	m.ALU.S = ins.SH

	/* Sub step 3 */
	if ins.MAR == 1 {
		m.MAR = m.Registers[ins.B]
	}

	m.ALU.Calc()
	/* Sub step 4 */
	if ins.MBR == 1 {
		m.MBR = m.ALU.R
	}
	if ins.ENC == 1 {
		m.Registers[ins.C] = m.ALU.R
	}

	m.RD = ins.RD
	m.WR = ins.WR

	m.MPC++
	switch ins.COND {
	case 1:
		if m.ALU.N == 1 {
			m.MPC = ins.ADDR
		}
	case 2:
		if m.ALU.Z == 1 {
			m.MPC = ins.ADDR
		}
	case 3:
		m.MPC = ins.ADDR
	}

	if m.RD == 1 && m.WR == 1 {
		m.DesiredState = HALT
	} else if m.RD == 1 {
		// check if READ is set
		if m.MARS != 0xFFFF {
			// Cycle 2
			// check if there is an address in the MAR staging
			//log.Printf("RD C2 MARS: %d", m.MARS)
			m.MBR = m.Memory[m.MARS]
			m.MARS = 0xFFFF
		} else {
			// Cycle 1
			// set MAR staging to address to be loaded on cycle 2
			//log.Printf("RD C1")
			// MAR ignores the upper 4 bits
			m.MARS = m.MAR & 0x0FFF
		}
	} else if m.WR == 1 {
		// check if WRITE is set
		if m.MARS != 0xFFFF {
			// Cycle 2
			// write the value in MBR staging to memory
			//log.Printf("WR C2 MARS: %d", m.MARS)
			m.Memory[m.MARS] = m.MBRS
			m.MARS = 0xFFFF
		} else {
			// Cycle 1
			// copy MAR to MAR staging and copy MBR to MBR staging
			m.MBRS = m.MBR
			// MAR ignores the upper 4 bits
			m.MARS = m.MAR & 0x0FFF
		}
	}
	m.Cycles++
	if m.MCC[m.MPC].BR {
		m.DesiredState = HALT
	}
}
