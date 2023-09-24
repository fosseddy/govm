     .global _start

_start:
    call some_fn
    halt

some_fn:
    push rbp
    mov rsp, rbp

    movi 15, r1
    sub r1, rsp
    // buf    = rsp-13
    // buflen = rsp-15

    // buflen = 13
    mov rbp, r1
    movi 15, r2
    sub r2, r1
    movi 13, r2
    wr r2, r1

    mov rbp, r1
    movi 13, r2
    sub r2, r1
    call copymsg

    movi 1, r1  // fd
    mov rbp, r2 // buf
    movi 13, r3
    sub r3, r2
    mov rbp, r3 // buflen
    movi 15, r4
    sub r4, r3
    rd r3, r3
    movi 1, r0  // write
    syscall

    mov rbp, rsp
    pop rbp
    ret

msg:
    .ascii "hello, world"
    .byte 10

// (dst: *byte): void
copymsg:
    movi 0, r2  // index
    movi 13, r3 // len
    movi 1, r4  // inc

    jmp copymsg_test
copymsg_loop:
    movi msg, r5
    add r2, r5
    rdb r5, r5
    mov r1, r6
    add r2, r6
    wrb r5, r6
    add r4, r2
copymsg_test:
    cmp r3, r2
    jl copymsg_loop

    ret
