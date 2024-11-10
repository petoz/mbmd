package rs485

import . "github.com/volkszaehler/mbmd/meters"

func init() {
	Register("ORNO1P525", NewORNO1P525Producer)
}

var ops1p525 Opcodes = Opcodes{
	/***
	 * Opcodes for ORNO WE-524 WE-525 and WE-526
	 * https://files.orno.pl/support/Others/ORNO/ORWE525_5908254827846/OR-WE-525_rejestry.pdf
	 */
	Frequency:     0x10A, // 16 bit, Hz
	Voltage:       0x100, // 16 bit, V
	Current:       0x102, // 16 bit, A
	Power:         0x104, // 16 bit, W
	ReactivePower: 0x108, // 16 bit, var
	ApparentPower: 0x106, // 16 bit, va
	Cosphi:        0x10B, // 16 bit,

	Sum:         0x10E, //32 Bit, wh
	ReactiveSum: 0x140, //32 Bit, varh
}

type ORNO1P525Producer struct {
	Opcodes
}

func NewORNO1P525Producer() Producer {
	return &ORNO1P525Producer{Opcodes: ops1p525}
}

// Description implements Producer interface
func (p *ORNO1P525Producer) Description() string {
	return "ORNO WE-525"
}

// snip creates modbus operation
func (p *ORNO1P525Producer) snip(iec Measurement, readlen uint16) Operation {
	return Operation{
		FuncCode: ReadHoldingReg,
		OpCode:   p.Opcode(iec), // adjust according to docs
		ReadLen:  readlen,
		IEC61850: iec,
	}
}

// snip16 creates modbus operation for single register
func (p *ORNO1P525Producer) snip16(iec Measurement, scaler ...float64) Operation {
	snip := p.snip(iec, 1)

	snip.Transform = RTUUint16ToFloat64 // default conversion
	if len(scaler) > 0 {
		snip.Transform = MakeScaledTransform(snip.Transform, scaler[0])
	}

	return snip
}

// snip32 creates modbus operation for double register
func (p *ORNO1P525Producer) snip32(iec Measurement, scaler ...float64) Operation {
	snip := p.snip(iec, 2)

	snip.Transform = RTUUint32ToFloat64 // default conversion
	if len(scaler) > 0 {
		snip.Transform = MakeScaledTransform(snip.Transform, scaler[0])
	}

	return snip
}

func (p *ORNO1P525Producer) Probe() Operation {
	return p.snip32(Voltage, 1)
}

// Produce implements Producer interface
func (p *ORNO1P525Producer) Produce() (res []Operation) {
	for _, op := range []Measurement{
		Power, ApparentPower,
	} {
		res = append(res, p.snip32(op, 1))
	}

	for _, op := range []Measurement{
		ReactivePower,
	} {
		res = append(res, p.snip32(op, 10000000))
	}

	for _, op := range []Measurement{
		Frequency,
	} {
		res = append(res, p.snip16(op, 10))
	}

	for _, op := range []Measurement{
		Voltage, Current,
	} {
		res = append(res, p.snip32(op, 1000))
	}

	for _, op := range []Measurement{
		Cosphi,
	} {
		res = append(res, p.snip16(op, 1000))
	}

	for _, op := range []Measurement{
		ReactiveSum, Sum,
	} {
		res = append(res, p.snip32(op, 100))
	}

	return res
}
