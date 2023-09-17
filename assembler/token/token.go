package token

import "fmt"

type Kind int

const (
	Eof Kind = iota

	Num
	Sym
	Str
	Char

	Colon
	Comma
	Dot

	Extern
	Global

	Byte
	Word
	String
	Skip

	Halt

	Mov
	Movb
	Movze
	Movse

	St
	Stb
	Ld
	Ldb

	Add
	Addb
	Sub
	Subb

	Cmp
	Cmpb

	Jmp
	Jz
	Je
	Jnz
	Jne
	Jc
	Jb
	Jnc
	Jae
	Js
	Jns
	Jo
	Jno
	Jbe
	Ja
	Jl
	Jge
	Jle
	Jg

	Push
	Pop

	Call
	Ret

	Syscall

	R0
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10
	R11
	R12
	R13
	Rsp
	Rbp
)

type Position struct {
	File string
	Line int
}

func (pos Position) String() string {
	return fmt.Sprintf("%s:%d", pos.File, pos.Line)
}

type Token struct {
	Kind Kind
	Lex string
	Value int
	Pos Position	
}

var keywords = map[string]Kind {
	"extern": Extern,
	"global": Global,

	"byte": Byte,
	"word": Word,
	"string": String,
	"skip": Skip,

	"halt": Halt,

	"mov": Mov,
	"movb": Movb,
	"movze": Movze,
	"movse": Movse,

	"st": St,
	"stb": Stb,
	"ld": Ld,
	"ldb": Ldb,

	"add": Add,
	"addb": Addb,
	"sub": Sub,
	"subb": Subb,

	"cmp": Cmp,
	"cmpb": Cmpb,

	"jmp": Jmp,
	"jz": Jz,
	"je": Je,
	"jnz": Jnz,
	"jne": Jne,
	"jc": Jc,
	"jb": Jb,
	"jnc": Jnc,
	"jae": Jae,
	"js": Js,
	"jns": Jns,
	"jo": Jo,
	"jno": Jno,
	"jbe": Jbe,
	"ja": Ja,
	"jl": Jl,
	"jge": Jge,
	"jle": Jle,
	"jg": Jg,

	"push": Push,
	"pop": Pop,

	"call": Call,
	"ret": Ret,

	"syscall": Syscall,

	"r0": R0,
	"r1": R1,
	"r2": R2,
	"r3": R3,
	"r4": R4,
	"r5": R5,
	"r6": R6,
	"r7": R7,
	"r8": R8,
	"r9": R9,
	"r10": R10,
	"r11": R11,
	"r12": R12,
	"r13": R13,
	"rsp": Rsp,
	"rbp": Rbp,
}

func LookupKeyword(lex string) Kind {
	if kind, ok := keywords[lex]; ok {
		return kind
	}
	return Sym
}
