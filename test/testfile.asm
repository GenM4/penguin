global _start
_f:
	push rbp			;; Local Stack position: 1
	mov rbp, rsp
	mov QWORD [rbp + 16], rdi
	mov QWORD [rbp + 24], rsi
	mov QWORD [rbp + 32], rdx
	mov rax, QWORD [rbp + 16]
	mov rbx, QWORD [rbp + 24]
	mul rbx
	push rax			;; Local Stack position: 5
	pop rax
	mov rbx, QWORD [rbp + 32]
	mul rbx
	push rax			;; Local Stack position: 5
	pop rax
	pop rbp
	ret
_start:
	push rbp			;; Local Stack position: 1
	mov rbp, rsp
	mov rax, 2h
	mov rdi, rax
	mov rax, 3h
	mov rsi, rax
	mov rax, 4h
	mov rdx, rax
	call _f
	mov QWORD [rbp + 16], rax
	mov rax, QWORD [rbp + 16]
	mov rbx, 2h
	sub rax, rbx
	push rax			;; Local Stack position: 3
	pop rax
	mov QWORD [rbp + 24], rax
	mov rax, 'H'
	push rax			;; Local Stack position: 4
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	pop rax
	mov rdi, 0h
	mov rax, 60
	syscall
	mov rdi, QWORD [rbp + 24]
	mov rax, 60
	syscall
	pop rbp
	ret
	mov rax, 60
	mov rdi, 0
	syscall
