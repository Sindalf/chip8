package main

import (
	//"fmt"
	"math/rand"
	"io/ioutil" 
	"github.com/nsf/termbox-go"
	"time"
)

/*
0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
0x200-0xFFF - Program ROM and work RAM
*/
var memory [4096]byte

var opcode uint16
var registers [16]byte // uint8
var pc uint16
var I uint16
var gfx [64*32]byte
var delay_timer uint8
var sound_timer uint8
var stack = make([]uint16, 0)
var sp int8 // actually pretty pointless
var draw_flag bool = false
var key byte

// Thanks ejholmes
var font_set = []byte {
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}


var keyMap = map[rune]byte{
	'1': 0x01, '2': 0x02, '3': 0x03, '4': 0x0C,
	'q': 0x04, 'w': 0x05, 'e': 0x06, 'r': 0x0D,
	'a': 0x07, 's': 0x08, 'd': 0x09, 'f': 0x0E,
	'z': 0x0A, 'x': 0x00, 'c': 0x0B, 'v': 0x0F,
}

func getKey() (byte) {
	event := termbox.PollEvent()
	////fmt.Println(event)
	return keyMap[event.Ch]
}

func push(a []uint16, x uint16) []uint16 {
	a = append(a, x)
	//////fmt.Println(&a)
	return a
}

func pop(a []uint16) (uint16, []uint16) {
	x := a[len(a)-1]
	a = a[:len(a)-1]
	return x,a
}

/*
nnn or addr - A 12-bit value, the lowest 12 bits of the instruction | 0x0FFF
n or nibble - A 4-bit value, the lowest 4 bits of the instruction 
x - A 4-bit value, the lower 4 bits of the high byte of the instruction
y - A 4-bit value, the upper 4 bits of the low byte of the instruction
kk or byte - An 8-bit value, the lowest 8 bits of the instruction
*/
func emulatecycle() {
	opcode = uint16(memory[pc]) << 8 | uint16(memory[pc+1])
	var x uint8 = uint8(opcode & 0x0F00 >> 8)
	var y uint8 = uint8(opcode & 0x00F0 >> 4)
	//var kk uint8 = uint8(opcode & 0x00FF)
	kk := byte(opcode)
	//fmt.Printf("X: %b, Y: %b, op: %X, Fop: %X, 8op: %X, pc: %X\n", x, y, opcode & 0xF000, opcode & 0x00FF, opcode & 0x000F, pc)
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
			case 0x00E0: //  00E0 - CLS
				clearScreen()
				pc += 2
			case 0x00EE: //  00EE - RET
				pc,stack = pop(stack)
				sp -= 1
				pc += 2
			default:
				//fmt.Println("unknown opcode ", opcode)
		}
		case 0x1000:  // 1nnn - JP addr
		
			pc = opcode & 0x0FFF
			//skip = true
		case 0x2000: // 2nnn - CALL addr
			sp += 1
			stack = push(stack, pc)
			pc = opcode & 0x0FFF
		//	skip = true
		case 0x3000: //  3xkk - SE Vx, byte
			pc += 2
			if registers[x] == kk {
				pc += 2
			} 
		case 0x4000: //  4xkk - SNE Vx, byte
			pc += 2
			if registers[x] != kk {
				pc += 2
			}
		case 0x5000:	// 5xy0 - SE Vx, Vy
			switch opcode & 0xF00F {
				case 0x5000: 
					pc += 2
					if registers[x] == registers[y] {
						pc += 2
					}
				default:
					//fmt.Println("unknown opcode 0x5000")
			}
		case 0x6000: //  6xkk - LD Vx, byte
			pc += 2
			registers[x] = kk
		case 0x7000: //  7xkk - ADD Vx, byte
			pc += 2
			registers[x] += kk
		case 0x8000:  //  requires switch for arguments
			switch opcode & 0x000F {
				case 0x0000: // 8xy0 - LD Vx, Vy
					registers[x] = registers[y]
					pc += 2
				case 0x0001: //  8xy1 - OR Vx, Vy
					registers[x] |= registers[y]
					pc += 2
				case 0x0002: //  8xy2 - AND Vx, Vy
					registers[x] &= registers[y]
					pc += 2
				case 0x0003: //  8xy3 - XOR Vx, Vy
					registers[x] ^= registers[y]
					pc += 2
				case 0x0004: //  8xy4 - ADD Vx, Vy
					r := uint16(registers[x]) + uint16(registers[y])
					if r > 0xFF {
						registers[0x000F] = 1
					} else {
						registers[0x000F] = 0
					}
					registers[x] = uint8(r)
					pc += 2
				case 0x0005: //  8xy5 - SUB Vx, Vy
					if registers[x] > registers[y] {
						registers[0x000F] = 1
					} else {
						registers[0x000F] = 0
					}
					registers[x] -= registers[y]
					pc += 2
				case 0x0006: // 8xy6 - SHR Vx {, Vy}
					if (registers[x] & 0x0001) == 0x0001 {
						registers[0x000F] = 1
					}  else {
						registers[0x000F] = 0
					}
					registers[x] /= 2
					pc += 2
				case 0x0007: // 8xy7 - SUBN Vx, Vy
					if registers[y] > registers[x] {
						registers[0x000F] = 1
					} else {
						registers[0x000F] = 0
					}
					registers[x] = registers[y] - registers[x]
					pc += 2
				case 0x000E: //  8xyE - SHL Vx {, Vy}
					if (registers[x] & 0x0080) == 0x0080 {
						registers[0x000F] = 1
					} else {
						registers[0x000F] = 0
					}
					registers[x] *= 2
					pc += 2
						
				default: 
					//fmt.Printf("Invalid opcode within 0x8000? %X\n", opcode)
			}
		case 0x9000: // 9xy0 - SNE Vx, Vy
			switch opcode & 0x000F {
				case 0x0000:
					pc += 2
					if registers[x] != registers[y] {
						pc += 2
					}
				default:
					//fmt.Println("unknown op code 0x9000")
			}
		case 0xA000: // Annn - LD I, addr
			I = opcode & 0x0FFF
			pc += 2
		case 0xB000: //  Bnnn - JP V0, addr
			pc = (opcode & 0x0FFF) + uint16(registers[0])
		case 0xC000: // Cxkk - RND Vx, byte
			var r uint8 = uint8(rand.Intn(255))
			registers[x] = r & kk
			pc += 2
		case 0xD000: // Dxyn - DRW Vx, Vy, nibble
			var x uint8 = registers[(opcode & 0x0F00) >> 8]
			var y uint8 = registers[(opcode & 0x00F0) >> 4]
			n := opcode & 0x000F
			sprite := memory[I:I+n]
			registers[0x000F] = 0
			//var height uint8 = uint8(opcode & 0x000F)
			for yline := uint8(0); yline < uint8(len(sprite)); yline++  {
				r := sprite[yline]
				for xline := uint8(0); xline < 8; xline++ {
					on := (r & byte(0x80 >> xline)) == byte(0x80 >> xline)
					v := byte(0)
					if on {
						v = 1
					}
						xg := uint16(x) + uint16(xline)
						yg := uint16(y) + uint16(yline)
				
						if(gfx[xg + yg * 64] == 1)  {
							registers[0xF] = 1 // collision detected
						}
						gfx[xg + yg * 64] ^= v
				}
			}
			draw_flag = true
			pc += 2
		case 0xE000:
			switch opcode & 0x00FF {
				case 0x009E:
					pc += 2
					if registers[x] == key {
						pc+=2
					}
					key = 0
				case 0x00A1:
					pc += 2
					if registers[x] != key {
						pc+=2
					}
					key = 0
				default:
					//fmt.Println("unknown opcode in 0xE000")
			}
		case 0xF000:
			switch opcode & 0x00FF {
				case 0x0007:
					registers[x] = delay_timer
					pc += 2
				case 0x000A:
					registers[x] = getKey()
					pc += 2
				case 0x0015:
					delay_timer = registers[x]
					pc += 2
				case 0x0018:
					sound_timer = registers[x]
					pc += 2
				case 0x001E:
					I += uint16(registers[x])
					pc += 2
				case 0x0029:
					I = uint16(registers[x]) * 5 // font_set starts at memory address 0
					pc += 2
				case 0x0033:
					memory[I] = (registers[x] / 100)
					memory[I+1] = (registers[x] / 10) % 10
					memory[I+2] = (registers[x] % 100) % 10
					pc += 2
				case 0x0055:
					for k := uint16(0); k <= uint16(x); k++ {
						memory[I+k] = registers[k]
					}
					pc += 2
				case 0x0065:
					for k := uint16(0); k <= uint16(x); k++ {
						registers[k] = memory[I+k]
					}
					pc += 2
				default:
					//fmt.Printf("Invalid opcode at 0xF000, %X", opcode & 0x00FF)
				}
				
		default:
			//fmt.Printf("Invalid opcode? %X\n", opcode)
	
	}
}

func draw() {
	w, h := termbox.Size()
	h = 32
	w = 64
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	gfx_count := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if gfx[gfx_count] == 1 {
				termbox.SetCell(x, y, ' ', termbox.ColorWhite, termbox.ColorWhite)
			}
			gfx_count++
		}
	}
	termbox.Flush()
}

func clearScreen() {
	for i := 0; i < len(gfx); i++  {
		gfx[i] = 0
	}
}

func cpu_init(data []byte) {
	pc = 0x200
	
	for i := 0; i < 80; i++ {
		memory[i] = font_set[i]
	}
	 
	for i := 0; i < len(data); i++ {
		memory[i + 0x200] = data[i]
	}
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	data,err := ioutil.ReadFile("INVADERS")
	termbox.SetInputMode(termbox.InputEsc)
	cpu_init(data)
	tick := time.Tick(time.Second / time.Duration(60))
	event_queue := make(chan termbox.Event)
        go func() {
                for {
                        event_queue <- termbox.PollEvent()
                }
        }()
		
	for {
		select {
		case <- tick:
			emulatecycle()
			if draw_flag == true {
				draw()
				draw_flag = false
			}
			if delay_timer > 0 {
				delay_timer--
			}
			if sound_timer > 0 {
				sound_timer--
			}
			select {
				case ev := <- event_queue:
					if ev.Ch != 0 {
						key = keyMap[ev.Ch]
					}
				default:
			}
		}
	}
}
