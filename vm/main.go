package main

import (
	"fmt"
	"os"
	"encoding/binary"
	_syscall "syscall"
)

const (
	halt uint8 = iota

	mov
	movb
	movi
	movze
	movse

	wr
	wrb
	rd
	rdb

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

func (r register) writeb(val byte) {
	regs[r * 2] = val	
}

func (r register) readb() byte {
	return regs[r * 2]
}

func (r register) write(val uint16) {
	addr := r * 2
	regs[addr] = byte(val)
	regs[addr + 1] = byte(val >> 8)
}

func (r register) read() uint16 {
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

type mainmem [1<<16]byte

func (mem *mainmem) writeb(addr uint16, val byte) {
	mem[addr] = val
}

func (mem *mainmem) readb(addr uint16) byte {
	return mem[addr]
}

func (mem *mainmem) write(addr uint16, val uint16) {
	mem.writeb(addr, byte(val))
	mem.writeb(addr + 1, byte(val >> 8))
}

func (mem *mainmem) read(addr uint16) uint16 {
	lsb := uint16(mem.readb(addr))
	msb := uint16(mem.readb(addr + 1))
	return msb << 8 | lsb
}

func (mem *mainmem) push(val uint16) {
	sp := rsp.read() - 2
	mem.write(sp, val)
	rsp.write(sp)
}

func (mem *mainmem) pop() uint16 {
	sp := rsp.read()
	v := mem.read(sp)
	sp += 2
	rsp.write(sp)
	return v
}

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
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Provide file to execute")
		os.Exit(1)
	}

	halted := false
	for !halted {
		op := ram.readb(ip)
		ip++

		switch op {
		case halt:
			halted = true

		case mov:
			src, dst := getRegs(ram.readb(ip))
			ip++
			dst.write(src.read())
		case movb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			dst.writeb(src.readb())
		case movi:
			src := register(ram.readb(ip))
			ip++
			imm := ram.read(ip)
			ip += 2
			src.write(imm)
		case movze:
			src, dst := getRegs(ram.readb(ip))
			ip++
			dst.write(uint16(src.readb()))
		case movse:
			src, dst := getRegs(ram.readb(ip))
			ip++
			b := src.readb()
			v := uint16(b)
			if b >> 7 == 1 {
				ones := ^uint16(0)
				v = ones << 8 | v
			}
			dst.write(v)

		case wr:
			src, dst := getRegs(ram.readb(ip))
			ip++
			ram.write(dst.read(), src.read())
		case wrb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			ram.writeb(dst.read(), src.readb())
		case rd:
			src, dst := getRegs(ram.readb(ip))
			ip++
			dst.write(ram.read(src.read()))
		case rdb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			dst.writeb(ram.readb(src.read()))

		case add:
			src, dst := getRegs(ram.readb(ip))
			ip++
			a, b := dst.read(), src.read()
			dst.write(a + b)
			setFlags(uint(a), uint(b), 16)
		case addb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			a, b := dst.readb(), src.readb()
			dst.writeb(a + b)
			setFlags(uint(a), uint(b), 8)
		case sub:
			src, dst := getRegs(ram.readb(ip))
			ip++
			a, b := dst.read(), src.read()
			dst.write(a - b)
			setFlags(uint(a), ^uint(b) + 1, 16)
		case subb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			a, b := dst.readb(), src.readb()
			dst.writeb(a - b)
			setFlags(uint(a), ^uint(b) + 1, 8)

		case cmp:
			src, dst := getRegs(ram.readb(ip))
			ip++
			setFlags(uint(dst.read()), ^uint(src.read()) + 1, 16)
		case cmpb:
			src, dst := getRegs(ram.readb(ip))
			ip++
			setFlags(uint(dst.readb()), ^uint(src.readb()) + 1, 8)

		case jmp:
			branch := ram.readb(ip)
			ip++
			addr := ram.read(ip)
			ip += 2

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
			src := register(ram.readb(ip))
			ip++
			ram.push(src.read())
		case pop:
			dst := register(ram.readb(ip))
			ip++
			dst.write(ram.pop())

		case call:
			addr := ram.read(ip)
			ip += 2
			ram.push(ip)
			ip = addr
		case ret:
			ip = ram.pop()

		case syscall:
			k := r0.read()
			switch k {
			case 1:
				fd := r1.read()
				ptr := r2.read()
				len := r3.read()
				_syscall.Write(int(fd), ram[ptr:ptr+len])
			default:
				panic("syscall kind is not implemented")
			}
		default:
			panic(fmt.Sprintf("unknown op %d\n", op))
		}

		//fmt.Println("IP:   ", ip)
		//fmt.Printf("oscz:  %04b\n", flags)
		//fmt.Println("Regs: ", regs)
		//fmt.Println("RAM:  ", ram[:80])
		//fmt.Println("Stack:", ram[len(ram) - 32:])
		//fmt.Println()
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
