    .global _start

msg:
    .ascii "hello, world"
    .byte 10
    .byte 0

_start:
    movi msg, r1
    call print_by_char

    movi msg, r1
    call print_by_len

    halt

// (r1: *byte): void
print_by_char:
    mov r1, r2 // buf
    movi 1, r0 // SYS_write
    movi 1, r1 // stdout
    movi 1, r3 // len
    movi 0, r4 // null terminator

    jmp print_by_char_check
print_by_char_loop:
    syscall
    add r3, r2
print_by_char_check:
    rdb r2, r5
    cmpb r4, r5
    jne print_by_char_loop

    ret

// (r1: *byte): void
print_by_len:
    call strlen

    mov r1, r2
    mov r0, r3
    movi 1, r1
    movi 1, r0
    syscall

    ret

// (r1: *byte): word
strlen:
    movi 0, r0 // len
    movi 1, r2 // inc
    movi 0, r3 // null terminator

    jmp strlen_check
strlen_loop:
    add r2, r0
strlen_check:
    push r1
    add r0, r1
    rdb r1, r1
    cmp r3, r1
    pop r1
    jne strlen_loop

    ret
