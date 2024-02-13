global _start
_start:
	mov rax, 51h
	push rax			;; Stack position: 1
	push QWORD [rsp + 0]			;; Stack position: 2
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
