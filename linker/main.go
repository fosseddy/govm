/*
Object file

Header
  nsyms - 2 bytes
  nrels - 2 bytes
  ncode - 2 bytes

Code ncode bytes

Symbols nsyms bytes
  kind   - 1 byte
  idx    - 2 bytes
  addr   - 2 bytes
  nlabel - 2 bytes
  label  - nlabel bytes

Relocations nrels bytes
  loc    - 2 bytes
  symidx - 2 bytes
*/

package main

import (
	"fmt"
	"os"
	"bytes"
	"encoding/binary"
)

type header struct {
	nsyms uint16
	nrels uint16
	ncode uint16
}

type module struct {
	idx int
	header
	code []byte
	locals []symbol
	rels []relocation
}

const (
	symlocal uint8 = iota
	symglobal
	symextern
)

type symbol struct {
	kind uint8
	idx uint16
	addr uint16
	label string
}

type relocation struct {
	loc uint16
	symidx uint16
}

type gsymbol struct {
	modidx int
	symidx uint16
}

var globals = map[string]gsymbol{}
var modules []module

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Provide object file[s] to link")
		os.Exit(1)
	}

	modules = make([]module, len(os.Args) - 1)
	off := 0
	
	for i, arg := range os.Args[1:] {
		f, err := os.Open(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var mod module
		mod.idx = i
		binary.Read(f, binary.LittleEndian, &mod.header.nsyms)
		binary.Read(f, binary.LittleEndian, &mod.header.nrels)
		binary.Read(f, binary.LittleEndian, &mod.header.ncode)

		mod.code = make([]byte, mod.header.ncode)
		binary.Read(f, binary.LittleEndian, mod.code)

		mod.locals = make([]symbol, mod.header.nsyms)
		for i := uint16(0); i < mod.header.nsyms; i++ {
			var s symbol
			binary.Read(f, binary.LittleEndian, &s.kind)
			binary.Read(f, binary.LittleEndian, &s.idx)
			binary.Read(f, binary.LittleEndian, &s.addr)

			addr := int(s.addr) + off
			if addr > int(^uint16(0)) {
				fmt.Fprintln(os.Stderr, "memory address overflow")
				os.Exit(1)
			}
			s.addr = uint16(addr)

			var len uint16
			binary.Read(f, binary.LittleEndian, &len)

			label := make([]byte, len)
			binary.Read(f, binary.LittleEndian, label)

			s.label = string(label)
			mod.locals[s.idx] = s

			if s.kind == symglobal {
				if _, ok := globals[s.label]; ok {
					fmt.Fprintf(os.Stderr, "global symbol %s already defined\n", s.label)
					os.Exit(1)
				}
				globals[s.label] = gsymbol{mod.idx, s.idx}
			}
		}

		mod.rels = make([]relocation, mod.header.nrels)
		for i := uint16(0); i < mod.header.nrels; i++ {
			var rel relocation
			binary.Read(f, binary.LittleEndian, &rel.loc)
			binary.Read(f, binary.LittleEndian, &rel.symidx)
			mod.rels[i] = rel
		}

		modules[mod.idx] = mod
		off += int(mod.header.ncode)
		f.Close()
	}

	for _, mod := range modules {
		for _, rel := range mod.rels {
			sym := mod.locals[rel.symidx]
			addr := sym.addr
			if sym.kind == symextern {
				gsym, ok := globals[sym.label]
				if !ok {
					fmt.Fprintf(os.Stderr, "symbol %s is not defined\n", sym.label)
					os.Exit(1)
				}
				addr = modules[gsym.modidx].locals[gsym.symidx].addr
			}
			// art: bro, i just want to replace 2 bytes from rel.loc, wtf is this
			buf := bytes.NewBuffer(mod.code[rel.loc-1:rel.loc])
			binary.Write(buf, binary.LittleEndian, addr)
		}
	}

	out, _ := os.Create("out.vm")

	_start_addr, ok := globals["_start"]
	if !ok {
		fmt.Fprintln(os.Stderr, "_start entry point is not defined")
		os.Exit(1)
	}

	binary.Write(out, binary.LittleEndian, modules[_start_addr.modidx].locals[_start_addr.symidx].addr)

	code := make([]byte, 0, 1024*5)
	for _, mod := range modules {
		code = append(code, mod.code...)
	}

	binary.Write(out, binary.LittleEndian, uint16(len(code)))
	binary.Write(out, binary.LittleEndian, code)

	out.Close()
}
