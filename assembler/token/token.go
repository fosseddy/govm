package token

import "fmt"

type Kind int

const (
	EOF Kind = iota
	LF

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
	Ascii
	Skip

	Halt
	Mov
	Movb
	Movze
	Movse
	Wr
	Wrb
	Rd
	Rdb
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

	tok_reg_begin
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
	tok_reg_end
)

func (k Kind) IsRegister() bool {
	return k > tok_reg_begin && k < tok_reg_end
}

func (k Kind) String() string {
	switch k {
	case EOF:
		return "<end of file>"
	case LF:
		return "<line feed>"

	case Num:
		return "number"
	case Sym:
		return "symbol"
	case Str:
		return "string"
	case Char:
		return "character"

	case Colon:
		return ":"
	case Comma:
		return ","
	case Dot:
		return "."

	case Extern:
		return "extern"
	case Global:
		return "global"
	case Byte:
		return "byte"
	case Word:
		return "word"
	case Ascii:
		return "ascii"
	case Skip:
		return "skip"

	case Halt:
		return "halt"
	case Mov:
		return "mov"
	case Movb:
		return "movb"
	case Movze:
		return "movze"
	case Movse:
		return "movse"
	case Wr:
		return "wr"
	case Wrb:
		return "wrb"
	case Rd:
		return "rd"
	case Rdb:
		return "rdb"
	case Add:
		return "add"
	case Addb:
		return "addb"
	case Sub:
		return "sub"
	case Subb:
		return "subb"
	case Cmp:
		return "cmp"
	case Cmpb:
		return "cmpb"
	case Jmp:
		return "jmp"
	case Jz:
		return "jz"
	case Je:
		return "je"
	case Jnz:
		return "jnz"
	case Jne:
		return "jne"
	case Jc:
		return "jc"
	case Jb:
		return "jb"
	case Jnc:
		return "jnc"
	case Jae:
		return "jae"
	case Js:
		return "js"
	case Jns:
		return "jns"
	case Jo:
		return "jo"
	case Jno:
		return "jno"
	case Jbe:
		return "jbe"
	case Ja:
		return "ja"
	case Jl:
		return "jl"
	case Jge:
		return "jge"
	case Jle:
		return "jle"
	case Jg:
		return "jg"
	case Push:
		return "push"
	case Pop:
		return "pop"
	case Call:
		return "call"
	case Ret:
		return "ret"
	case Syscall:
		return "syscall"

	case R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12, R13, Rsp, Rbp:
		return "register"
	}

	panic("unreachable")
}

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
	"ascii": Ascii,
	"skip": Skip,

	"halt": Halt,
	"mov": Mov,
	"movb": Movb,
	"movze": Movze,
	"movse": Movse,
	"wr": Wr,
	"wrb": Wrb,
	"rd": Rd,
	"rdb": Rdb,
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
