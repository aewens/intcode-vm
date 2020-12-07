package intcode

import (
	"fmt"
	"strings"

	"github.com/aewens/intcode-vm/pkg/shared"
)

type Computer struct {
	Position  int
	RPosition int
	Codes     []int
	Memory    map[int]int
	InBuffer  chan int
	OutBuffer chan int
	InQueue   []int
	OutQueue  []int
	Halted    bool
}

type Opcode struct {
	Value int
	Modes []int
}

func createMemoryMap(codes []int) map[int]int {
	memory := make(map[int]int)
	for c, code := range codes {
		memory[c] = code
	}
	return memory
}

func Parser(program string) []int {
	var codes []int

	instructions := strings.Split(program, ",")
	for _, instruction := range instructions {
		code := shared.StringToInt(instruction)
		codes = append(codes, code)
	}

	return codes
}

func New(program string) *Computer {
	codes := Parser(program)
	memory := createMemoryMap(codes)

	var inBuffer chan int = nil
	var outBuffer chan int = nil

	var inQueue []int = nil
	var outQueue []int = nil

	return &Computer{
		Position:  0,
		RPosition: 0,
		Codes:     codes,
		Memory:    memory,
		InBuffer:  inBuffer,
		OutBuffer: outBuffer,
		InQueue:   inQueue,
		OutQueue:  outQueue,
		Halted:    false,
	}
}

func BufferedNew(program string) *Computer {
	codes := Parser(program)
	memory := createMemoryMap(codes)

	var inQueue []int = nil
	var outQueue []int = nil

	return &Computer{
		Position:  0,
		RPosition: 0,
		Codes:     codes,
		Memory:    memory,
		InBuffer:  make(chan int),
		OutBuffer: make(chan int),
		InQueue:   inQueue,
		OutQueue:  outQueue,
		Halted:    false,
	}
}

func QueuedNew(program string) *Computer {
	codes := Parser(program)
	memory := createMemoryMap(codes)

	var inBuffer chan int = nil
	var outBuffer chan int = nil

	return &Computer{
		Position:  0,
		RPosition: 0,
		Codes:     codes,
		Memory:    memory,
		InBuffer:  inBuffer,
		OutBuffer: outBuffer,
		InQueue:   []int{},
		OutQueue:  []int{},
		Halted:    false,
	}
}

func (computer *Computer) CheckAddress(address int) {
	_, ok := computer.Memory[address]
	if !ok {
		computer.Memory[address] = 0
	}
}

func (computer *Computer) ReadMemory(address int) int {
	computer.CheckAddress(address)
	//fmt.Println("R", address, computer.Memory[address])
	return computer.Memory[address]
}

func (computer *Computer) WriteMemory(address int, value int) {
	computer.CheckAddress(address)
	computer.Memory[address] = value
	//fmt.Println("W", address, computer.Memory[address])
}

func (computer *Computer) Next() {
	computer.Position = computer.Position + 1
}

func (computer *Computer) Jump(address int) {
	computer.CheckAddress(address)
	computer.Position = address
}

func (computer *Computer) RelativeJump(address int) {
	relativeAddress := computer.RPosition + address
	computer.CheckAddress(relativeAddress)
	computer.RPosition = relativeAddress
}

func (computer *Computer) Read() int {
	return computer.ReadMemory(computer.Position)
}

func (computer *Computer) ReadRelative(address int) int {
	return computer.ReadMemory(computer.RPosition + address)
}

func (computer *Computer) ReadNext() int {
	computer.Next()
	return computer.Read()
}

func (computer *Computer) ReadRelativeNext() int {
	address := computer.ReadNext()
	return computer.ReadRelative(address)
}

func (computer *Computer) ReadFromNext() int {
	return computer.ReadMemory(computer.ReadNext())
}

func (computer *Computer) ReadFromRelativeNext() int {
	return computer.ReadMemory(computer.ReadRelativeNext())
}

func (computer *Computer) ReadNextGivenMode(mode int) int {
	switch mode {
	case 0:
		return computer.ReadFromNext()

	case 1:
		return computer.ReadNext()

	case 2:
		return computer.ReadRelativeNext()

	default:
		panic(fmt.Sprintf("Unknown mode: %d", mode))
	}
}

func (computer *Computer) WriteToNext(value int) {
	computer.WriteMemory(computer.ReadNext(), value)
}

func (computer *Computer) WriteToRelative(value int) {
	address := computer.RPosition + computer.ReadNext()
	computer.WriteMemory(address, value)
}

func (computer *Computer) WriteNextGivenMode(mode int, value int) {
	switch mode {
	case 0:
		computer.WriteToNext(value)

	case 1:
		panic("Illegal mode for write operations")

	case 2:
		computer.WriteToRelative(value)

	default:
		panic(fmt.Sprintf("Unknown mode: %d", mode))
	}
}

func (computer *Computer) ReadOpcode(value int) *Opcode {
	modes := []int{}
	if value == 99 {
		return &Opcode{value, modes}
	}

	parsed := shared.IntToString(value)
	parsedSize := len(parsed)
	if parsedSize < 2 {
		parsed = "0" + parsed
		parsedSize = 2
	}

	opcode := shared.StringToInt(parsed[parsedSize-2:])

	params := parsed[:parsedSize-2]
	paramsSize := len(params)

	expectedModes := make(map[int]int)
	expectedModes[1] = 3
	expectedModes[2] = 3
	expectedModes[3] = 1
	expectedModes[4] = 1
	expectedModes[5] = 2
	expectedModes[6] = 2
	expectedModes[7] = 3
	expectedModes[8] = 3
	expectedModes[9] = 1

	expecting, ok := expectedModes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown mode: %d | %d", value, computer.Position))
	}

	// Left-pad the missing zeroes
	if paramsSize < expecting {
		for e := 0; e < expecting-paramsSize; e++ {
			params = "0" + params
		}
		paramsSize = expecting
	}

	for p := 1; p <= paramsSize; p++ {
		param := rune(params[paramsSize-p])
		mode := shared.RuneToInt(param)
		modes = append(modes, mode)
	}

	return &Opcode{opcode, modes}
}

func (computer *Computer) ReadNextOpcode() *Opcode {
	return computer.ReadOpcode(computer.Read())
}

func (computer *Computer) Step() {
	opcode := computer.ReadNextOpcode()
	//fmt.Println(computer.Position, opcode)

	switch opcode.Value {
	case 1: // ADD
		value1 := computer.ReadNextGivenMode(opcode.Modes[0])
		value2 := computer.ReadNextGivenMode(opcode.Modes[1])
		computer.WriteNextGivenMode(opcode.Modes[2], value1+value2)
		computer.Next()

	case 2: // MUL
		value1 := computer.ReadNextGivenMode(opcode.Modes[0])
		value2 := computer.ReadNextGivenMode(opcode.Modes[1])
		computer.WriteNextGivenMode(opcode.Modes[2], value1*value2)
		computer.Next()

	case 3: // GET
		var input int

		if computer.InBuffer == nil && computer.InQueue == nil {
			_, err := fmt.Scan(&input)
			if err != nil {
				panic(err)
			}
		} else if computer.InBuffer != nil {
			buffer, ok := <-computer.InBuffer
			if !ok {
				panic("Input buffer is empty")
			}

			input = buffer
		} else if computer.InQueue != nil {
			input = computer.InQueue[len(computer.InQueue)-1]
			computer.InQueue = computer.InQueue[:len(computer.InQueue)-1]
		}

		computer.WriteNextGivenMode(opcode.Modes[0], input)
		computer.Next()

	case 4: // SHW
		value := computer.ReadNextGivenMode(opcode.Modes[0])
		if computer.OutBuffer == nil && computer.OutQueue == nil {
			fmt.Println(value)
		} else if computer.OutBuffer != nil {
			computer.OutBuffer <- value
		} else if computer.OutQueue != nil {
			computer.OutQueue = append(computer.OutQueue, value)
		}
		computer.Next()

	case 5: // JIT
		check := computer.ReadNextGivenMode(opcode.Modes[0])
		position := computer.ReadNextGivenMode(opcode.Modes[1])
		if check != 0 {
			computer.Jump(position)
		} else {
			computer.Next()
		}

	case 6: // JIF
		check := computer.ReadNextGivenMode(opcode.Modes[0])
		position := computer.ReadNextGivenMode(opcode.Modes[1])
		if check == 0 {
			computer.Jump(position)
		} else {
			computer.Next()
		}

	case 7: // LST
		value1 := computer.ReadNextGivenMode(opcode.Modes[0])
		value2 := computer.ReadNextGivenMode(opcode.Modes[1])
		if value1 < value2 {
			computer.WriteNextGivenMode(opcode.Modes[2], 1)
		} else {
			computer.WriteNextGivenMode(opcode.Modes[2], 0)
		}

		computer.Next()

	case 8: // EQT
		value1 := computer.ReadNextGivenMode(opcode.Modes[0])
		value2 := computer.ReadNextGivenMode(opcode.Modes[1])
		if value1 == value2 {
			computer.WriteNextGivenMode(opcode.Modes[2], 1)
		} else {
			computer.WriteNextGivenMode(opcode.Modes[2], 0)
		}

		computer.Next()

	case 9: // ARP
		position := computer.ReadNextGivenMode(opcode.Modes[0])
		computer.RelativeJump(position)

		computer.Next()

	case 99:
		computer.Halted = true

	default:
		panic(fmt.Sprintf("Invalid opcode: %d", opcode))
	}
}

func (computer *Computer) StepUntil(opcodes ...int) int {
	for {
		nextOpcode := computer.ReadNextOpcode()
		for _, opcode := range opcodes {
			if opcode == nextOpcode.Value {
				return opcode
			}
		}

		computer.Step()
	}
}

func (computer *Computer) Run() map[int]int {
	for {
		computer.Step()

		if computer.Halted {
			break
		}
	}

	return computer.Memory
}

func (computer *Computer) Reset() {
	computer.Position = 0
	computer.RPosition = 0
	computer.Memory = createMemoryMap(computer.Codes)
	computer.Halted = false
}

func (computer *Computer) Load(program string) {
	computer.Codes = Parser(program)
	computer.Reset()
}

func (computer *Computer) Input(value int) {
	if computer.InBuffer == nil && computer.InQueue == nil {
		panic("Cannot input without buffers")
	} else if computer.InBuffer != nil {
		computer.InBuffer <- value
	} else if computer.InQueue != nil {
		computer.InQueue = append(computer.InQueue, value)
	}
}

func (computer *Computer) Output() int {
	if computer.OutBuffer != nil {
		return <-computer.OutBuffer
	}

	if computer.OutQueue != nil {
		output := computer.OutQueue[len(computer.OutQueue)-1]
		computer.OutQueue = computer.OutQueue[:len(computer.OutQueue)-1]
		return output
	}

	panic("Cannot output without buffers")
}

func (computer *Computer) RunAndReset() {
	computer.Run()
	computer.Reset()
}
