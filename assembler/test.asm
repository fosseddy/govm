msg_1:
    .ascii "hello, from _start"
    .byte 10

msg_2:
    .ascii "hello, from function"
    .byte 10

_start:
    call function

    movi 1, r1
    movi msg_1, r2
    movi 19, r3
    movi 1, r0
    syscall

    halt

function:
    push r0
    push r1
    push r2
    push r3

    movi 1, r1
    movi msg_2, r2
    movi 21, r3
    movi 1, r0
    syscall

    pop r3
    pop r2
    pop r1
    pop r0
    ret

