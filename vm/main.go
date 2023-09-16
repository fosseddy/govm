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

	push
	pop

	call
	ret

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

func (r register) storeb(val byte) {
	regs[r * 2] = val	
}

func (r register) loadb() byte {
	return regs[r * 2]
}

func (r register) store(val uint16) {
	addr := r * 2
	regs[addr] = byte(val)
	regs[addr + 1] = byte(val >> 8)
}

func (r register) load() uint16 {
	addr := r * 2
	lsb := uint16(regs[addr])
	msb := uint16(regs[addr + 1])
	return msb << 8 | lsb
}

const (
	zflag uint8 = 0b0001 << iota
	cflag
	sflag
	oflag
)

type instmem [1<<16]byte
type mainmem [1<<16]byte

func (mem *instmem) storeb(addr uint16, val byte) {
	mem[addr] = val
}

func (mem *instmem) loadb() byte {
	b := mem[ip]
	ip++
	return b
}

func (mem *instmem) store(addr uint16, val uint16) {
	mem.storeb(addr, byte(val))
	mem.storeb(addr + 1, byte(val >> 8))
}

func (mem *instmem) load() uint16 {
	lsb := uint16(mem.loadb())
	msb := uint16(mem.loadb())
	return msb << 8 | lsb
}

func (mem *mainmem) storeb(addr uint16, val byte) {
	mem[addr] = val
}

func (mem *mainmem) loadb(addr uint16) byte {
	return mem[addr]
}

func (mem *mainmem) store(addr uint16, val uint16) {
	mem.storeb(addr, byte(val))
	mem.storeb(addr + 1, byte(val >> 8))
}

func (mem *mainmem) load(addr uint16) uint16 {
	lsb := uint16(mem.loadb(addr))
	msb := uint16(mem.loadb(addr + 1))
	return msb << 8 | lsb
}

func (mem *mainmem) push(val uint16) {
	sp := rsp.load() - 2
	mem.store(sp, val)
	rsp.store(sp)
}

func (mem *mainmem) pop() uint16 {
	sp := rsp.load()
	v := mem.load(sp)
	sp += 2
	rsp.store(sp)
	return v
}

var rom instmem
var ram mainmem
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
	ip = 0
	rsp.store(0)

	halted := false
	for !halted {
		op := rom.loadb()

		switch op {
		case halt:
			halted = true

		case mov:
			src, dst := getRegs(rom.loadb())
			dst.store(src.load())
		case movb:
			src, dst := getRegs(rom.loadb())
			dst.storeb(src.loadb())
		case movi:
			src := register(rom.loadb())
			imm := rom.load()
			src.store(imm)
		case movze:
			src, dst := getRegs(rom.loadb())
			dst.store(uint16(src.loadb()))
		case movse:
			src, dst := getRegs(rom.loadb())
			b := src.loadb()
			v := uint16(b)
			if b >> 7 == 1 {
				ones := ^uint16(0)
				v = ones << 8 | v
			}
			dst.store(v)

		case st:
			src, dst := getRegs(rom.loadb())
			ram.store(dst.load(), src.load())
		case stb:
			src, dst := getRegs(rom.loadb())
			ram.storeb(dst.load(), src.loadb())
		case ld:
			src, dst := getRegs(rom.loadb())
			dst.store(ram.load(src.load()))
		case ldb:
			src, dst := getRegs(rom.loadb())
			dst.storeb(ram.loadb(src.load()))

		case add:
			src, dst := getRegs(rom.loadb())
			a, b := dst.load(), src.load()
			v := a + b
			as, bs, vs := int16(a), int16(b), int16(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v < a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs > 0 && vs <= 0) || (as < 0 && bs < 0 && vs >= 0) {
				flags |= oflag
			}
			dst.store(v)
		case addb:
			src, dst := getRegs(rom.loadb())
			a, b := dst.loadb(), src.loadb()
			v := a + b
			as, bs, vs := int8(a), int8(b), int8(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v < a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs > 0 && vs <= 0) || (as < 0 && bs < 0 && vs >= 0) {
				flags |= oflag
			}
			dst.storeb(v)
		case sub:
			src, dst := getRegs(rom.loadb())
			a, b := dst.load(), src.load()
			v := a - b
			as, bs, vs := int16(a), int16(b), int16(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v > a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= oflag
			}
			dst.store(v)
		case subb:
			src, dst := getRegs(rom.loadb())
			a, b := dst.loadb(), src.loadb()
			v := a - b
			as, bs, vs := int8(a), int8(b), int8(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v > a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= oflag
			}
			dst.storeb(v)

		case cmp:
			src, dst := getRegs(rom.loadb())
			a, b := dst.load(), src.load()
			v := a - b
			as, bs, vs := int16(a), int16(b), int16(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v > a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= oflag
			}
		case cmpb:
			src, dst := getRegs(rom.loadb())
			a, b := dst.loadb(), src.loadb()
			v := a - b
			as, bs, vs := int8(a), int8(b), int8(v)
			flags = 0
			if v == 0 {
				flags |= zflag
			}
			if v > a {
				flags |= cflag
			}
			if vs < 0 {
				flags |= sflag
			}
			if (as > 0 && bs < 0 && vs <= 0) || (as < 0 && bs > 0 && vs >= 0) {
				flags |= oflag
			}

		case jmp:
			branch := rom.loadb()
			addr := rom.load()

			zf := flags & 0b1
			cf := flags >> 1 & 0b1
			sf := flags >> 2 & 0b1
			of := flags >> 3 & 0b1

			setAddr := false

			switch branch {
			case 0: // jmp
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
				panic(fmt.Sprintf("unknown jmp branch %d\n", branch))
			}

			if setAddr {
				ip = addr
			}

		case push:
			src := register(rom.loadb())
			ram.push(src.load())
		case pop:
			dst := register(rom.loadb())
			dst.store(ram.pop())

		case call:
			addr := rom.load()
			ram.push(addr)
			ip = addr
		case ret:
			ip = ram.pop()

		case syscall:
			panic("not implemented")
		default:
			panic(fmt.Sprintf("unknown op %d\n", op))
		}

		fmt.Println("IP:   ", ip)
		fmt.Printf("Flags: %04b\n", flags)
		fmt.Println("Regs: ", regs)
		fmt.Println("ROM:  ", rom[:32])
		fmt.Println("RAM:  ", ram[:32])
		fmt.Println("Stack:", ram[len(ram) - 32:])
		
		fmt.Println()
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
