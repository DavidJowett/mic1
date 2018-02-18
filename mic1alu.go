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
