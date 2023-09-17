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

type condflag uint8

const (
	zflag condflag = 0b0001 << iota
	cflag
	sflag
	oflag
)

func (f condflag) set() {
	flags |= uint8(f)
}

type instmem [1<<16]byte

func (mem *instmem) storeb(val byte) {
	mem[ip] = val
	ip++
}

func (mem *instmem) loadb() byte {
	b := mem[ip]
	ip++
	return b
}

func (mem *instmem) store(val uint16) {
	mem.storeb(byte(val))
	mem.storeb(byte(val >> 8))
}

func (mem *instmem) load() uint16 {
	lsb := uint16(mem.loadb())
	msb := uint16(mem.loadb())
	return msb << 8 | lsb
}

type mainmem [1<<16]byte

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
	rom.storeb(movi)
	rom.storeb(byte(r0))
	rom.store(127)

	rom.storeb(movi)
	rom.storeb(byte(r1))
	rom.store(1)

	rom.storeb(addb)
	rom.storeb(byte(r1 << 4 | r0))

	rom.storeb(halt)

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
			dst.store(a + b)
			setFlags(uint(a), uint(b), 16)
		case addb:
			src, dst := getRegs(rom.loadb())
			a, b := dst.loadb(), src.loadb()
			dst.storeb(a + b)
			setFlags(uint(a), uint(b), 8)
		case sub:
			src, dst := getRegs(rom.loadb())
			a, b := dst.load(), src.load()
			dst.store(a - b)
			setFlags(uint(a), ^uint(b) + 1, 16)
		case subb:
			src, dst := getRegs(rom.loadb())
			a, b := dst.loadb(), src.loadb()
			dst.storeb(a - b)
			setFlags(uint(a), ^uint(b) + 1, 8)

		case cmp:
			src, dst := getRegs(rom.loadb())
			setFlags(uint(dst.load()), ^uint(src.load()) + 1, 16)
		case cmpb:
			src, dst := getRegs(rom.loadb())
			setFlags(uint(dst.loadb()), ^uint(src.loadb()) + 1, 8)

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
			case 1: // jz, je
				setAddr = zf == 1
			case 2: // jnz, jne
				setAddr = zf == 0
			case 3: // jc, jb
				setAddr = cf == 1
			case 4: // jnc, jae
				setAddr = cf == 0
			case 5: // js
				setAddr = sf == 1
			case 6: // jns
				setAddr = sf == 0
			case 7: // jo
				setAddr = of == 1
			case 8: // jno
				setAddr = of == 0
			case 9: // jbe
				setAddr = cf | zf == 1
			case 10: // ja
				setAddr = cf | zf == 0
			case 11: // jl
				setAddr = sf ^ of == 1
			case 12: // jge
				setAddr = sf ^ of == 0
			case 13: // jle
				setAddr = (sf ^ of) | zf == 1
			case 14: // jg
				setAddr = (sf ^ of) | zf == 0
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
		fmt.Printf("oscz:  %04b\n", flags)
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

func setFlags(a, b uint, size int) {
	v := a + b
	if v == 0 {
		zflag.set()
	}

	carry := v >> size & 0b1
	if carry == 1 {
		cflag.set()
	}

	sign := v >> (size - 1) & 0b1
	if sign == 1 {
		sflag.set()
	}
	
	as := a >> (size - 1) & 0b1
	bs := b >> (size - 1) & 0b1
	if (as == 0 && bs == 0 && sign == 1) || (as == 1 && bs == 1 && sign == 0) {
		oflag.set()
	}
}
