msg:
    .ascii "hello, world"
    .byte 10

_start:
    mov 1, r1
    mov msg, r2
    mov 13, r3
    mov 1, r0
    syscall
    halt
