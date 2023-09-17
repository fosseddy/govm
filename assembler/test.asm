
	.extern even
	.global more

var_1:
	.byte 69
//	.byte 'o'
var_2:
	.word 420
var_3:
//	.string "hello, world"
	.byte 10
//	.string "this is the end"
	.byte 10
	.byte 0
var_4:
//	.string "hello, world"
	.byte 10
	.byte 0
var_5:
	.skip 13

_start:
	mov 0 r1
	mov 60 r0
	mov 'c', r2
	syscall

	halt
