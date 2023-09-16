package main

import (
	"fmt"
)

const (
	halt uint8 = iota

	mov
	movb
	movi
	movze
	movse

	st
	stb
	ld
	ldb

	add
	addb
	sub
	subb

	cmp
	cmpb

	jmp

	syscall
)

type register uint8

const (
	r0 register = iota
	r1
	r2
	r3
	r4
	r5
	r6
	r7
	r8
	r9
	r10
	r11
	r12
	r13

	rsp
	rbp

	rcount
)

func (r register) writeb(val byte) {
	regs[r * 2] = val	
}

func (r register) write(val uint16) {
	addr := r * 2
	regs[addr] = byte(val)
	regs[addr + 1] = byte(val >> 8)
}

func (r register) readb() byte {
	return regs[r * 2]
}

func (r register) read() uint16 {
	addr := r * 2
	lsb := uint16(regs[addr])
	msb := uint16(regs[addr + 1])
	return msb << 8 | lsb
}

type memory [1<<16]byte

func (mem *memory) writeb(addr uint16, val byte) {
	mem[addr] = val
}

func (mem *memory) write(addr uint16, val uint16) {
	mem.writeb(addr, byte(val))
	mem.writeb(addr + 1, byte(val >> 8))
}

func (mem *memory) readb(addr uint16) byte {
	return mem[addr]
}

func (mem *memory) read(addr uint16) uint16 {
	lsb := uint16(mem.readb(addr))
	msb := uint16(mem.readb(addr + 1))
	return msb << 8 | lsb
}

var rom memory
var ram memory
var regs [rcount*2]byte

var ip uint16
var flags uint8

func init() {
	maxrcount := 1 << 4
	if int(rcount) > maxrcount {
		panic(fmt.Sprintf("register count %d is more than max register count %d\n", rcount, maxrcount))
	}
}

func main() {
	var i uint16 = 0

	rom.writeb(i, cmp)
	i++
	rom.writeb(i, byte(r0 << 4 | r1))
	i++

	rom.writeb(i, jmp)
	i++
	rom.writeb(i, 16)
	i++
	rom.write(i, 420)
	i += 2

	rom.writeb(i, halt)
	i++

	r0.write(25)
	r1.write(26)

	ip = 0

	halted := false
	for !halted {
		fmt.Println("IP:   ", ip)
		fmt.Printf("Flags: %04b\n", flags)
		fmt.Println("Regs:", regs)
		fmt.Println("ROM: ", rom[:32])
		//fmt.Printf("  %08b\n", rom[:i])

		fmt.Println("RAM: ", ram[:32])
		//fmt.Printf("  %08b\n", ram[:i])
		
		fmt.Println()

		op := rom.readb(ip)
		ip++

		switch op {
		case halt:
			halted = true

		case mov:
			src, dst := getRegs(rom.readb(ip))
			ip++
			dst.write(src.read())
		case movb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			dst.writeb(src.readb())
		case movi:
			src := register(rom.readb(ip))
			ip++
			imm := rom.read(ip)
			ip += 2
			src.write(imm)
		case movze:
			src, dst := getRegs(rom.readb(ip))
			ip++
			dst.write(uint16(src.readb()))
		case movse:
			src, dst := getRegs(rom.readb(ip))
			ip++
			b := src.readb()
			v := uint16(b)
			if b >> 7 & 0b1 == 1 {
				mask := ^uint16(0)
				v = mask << 8 | v
			}
			dst.write(v)

		case st:
			src, dst := getRegs(rom.readb(ip))
			ip++
			ram.write(dst.read(), src.read())
		case stb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			ram.writeb(dst.read(), src.readb())
		case ld:
			src, dst := getRegs(rom.readb(ip))
			ip++
			dst.write(ram.read(src.read()))
		case ldb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			dst.writeb(ram.readb(src.read()))

		case add:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.read(), src.read()
			v := a + b
			as, bs, vs := int16(a), int16(b), int16(v)
			if v == 0 {
				flags |= 0b0001
			}
			if v < a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs > 0 && vs <= 0) || (as < 0 && bs < 0 && vs >= 0) {
				flags |= 0b1000
			}
			dst.write(v)
		case addb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.readb(), src.readb()
			v := a + b
			as, bs, vs := int8(a), int8(b), int8(v)
			if v == 0 {
				flags |= 0b0001
			}
			if v < a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs > 0 && vs <= 0) || (as < 0 && bs < 0 && vs >= 0) {
				flags |= 0b1000
			}
			dst.writeb(v)
		case sub:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.read(), src.read()
			v := a - b
			as, bs, vs := int16(a), int16(b), int16(v)
			println(a,b,v)
			println(as,bs,vs)
			if v == 0 {
				flags |= 0b0001
			}
			if v > a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= 0b1000
			}
			dst.write(v)
		case subb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.readb(), src.readb()
			v := a - b
			as, bs, vs := int8(a), int8(b), int8(v)
			if v == 0 {
				flags |= 0b0001
			}
			if v > a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= 0b1000
			}
			dst.writeb(v)

		case cmp:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.read(), src.read()
			v := a - b
			as, bs, vs := int16(a), int16(b), int16(v)
			if v == 0 {
				flags |= 0b0001
			}
			if v > a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= 0b1000
			}
		case cmpb:
			src, dst := getRegs(rom.readb(ip))
			ip++
			a, b := dst.readb(), src.readb()
			v := a - b
			as, bs, vs := int8(a), int8(b), int8(v)
			if v == 0 {
				flags |= 0b0001
			}
			if v > a {
				flags |= 0b0010
			}
			if vs < 0 {
				flags |= 0b0100
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= 0b1000
			}

		case jmp:
			br := rom.readb(ip)
			ip++
			addr := rom.read(ip)
			ip += 2

			zf := flags & 0b0001
			cf := flags >> 1 & 0b0001
			sf := flags >> 2 & 0b0001
			of := flags >> 3 & 0b0001

			setAddr := false

			switch br {
			case 0:
				setAddr = true
			case 1: // je
				setAddr = zf == 1
			case 2: // jne
				setAddr = not(zf) == 1
			case 3: // jc
				setAddr = cf == 1
			case 4: // jnc
				setAddr = not(cf) == 1
			case 5: // js
				setAddr = sf == 1
			case 6: // jns
				setAddr = not(sf) == 1
			case 7: // jo
				setAddr = of == 1
			case 8: // jno
				setAddr = not(of) == 1
			case 9: // ja
				setAddr = (not(cf) & not(zf)) == 1
			case 10: // jae
				setAddr = not(cf) == 1
			case 11: // jb
				setAddr = cf == 1
			case 12: // jbe
				setAddr = (cf | zf) == 1
			case 13: // jg
				setAddr = (not(sf ^ of) & not(zf)) == 1
			case 14: // jge
				setAddr = not(sf ^ of) == 1
			case 15: // jl
				setAddr = (sf ^ of) == 1
			case 16: // jle
				setAddr = ((sf ^ of) | zf) == 1
			default:
				panic(fmt.Sprintf("unknown jmp branch %d\n", br))
			}

			if setAddr {
				ip = addr
			}

		case syscall:
			panic("not implemented")
		default:
			panic(fmt.Sprintf("unknown op %d\n", op))
		}
	}
}

func getRegs(b byte) (register, register) {
	src := register(b >> 4 & 0b1111)
	dst := register(b & 0b1111)
	return src, dst
}

func not(b uint8) uint8 {
	if b == 1 {
		return 0
	}

	return 1
}
