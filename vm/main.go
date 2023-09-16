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
	inc
	incb
	dec
	decb

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

func init() {
	maxrcount := 1 << 4
	if int(rcount) > maxrcount {
		panic(fmt.Sprintf("register count %d is more than max register count %d\n", rcount, maxrcount))
	}
}

func main() {
	var i uint16 = 0

	rom.writeb(i, ldb)
	i++
	rom.writeb(i, byte(r0 << 4 | r1))
	i++

	rom.writeb(i, halt)
	i++

	r0.write(0)
	ram.write(0, 42069)

	ip = 0

	halted := false
	for !halted {

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
			panic("not implemented")
		case addb:
			panic("not implemented")
		case sub:
			panic("not implemented")
		case subb:
			panic("not implemented")
		case inc:
			panic("not implemented")
		case incb:
			panic("not implemented")
		case dec:
			panic("not implemented")
		case decb:
			panic("not implemented")

		case cmp:
			panic("not implemented")
		case cmpb:
			panic("not implemented")

		case jmp:
			panic("not implemented")

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
