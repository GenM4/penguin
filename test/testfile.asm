global _start
_f:
	push rbp			;; Local Stack position: 1
	mov rbp, rsp
	mov QWORD [rbp + 16], rdi
	mov rax, 3h
	mov QWORD [rbp + 24], rax
	mov rax, QWORD [rbp + 16]
	mov rbx, QWORD [rbp + 24]
	add rax, rbx
	push rax			;; Local Stack position: 4
	pop rax
	mov rbx, 5h
	add rax, rbx
	push rax			;; Local Stack position: 4
	pop rax
	pop rbp
	ret
_start:
	push rbp			;; Local Stack position: 1
	mov rbp, rsp
	mov rax, 5h
	mov QWORD [rbp + 16], rax
	mov rax, QWORD [rbp + 16]
	mov rbx, 5h
	add rax, rbx
	push rax			;; Local Stack position: 3
	pop rax
	mov rdi, rax
	call _f
	mov QWORD [rbp + 24], rax
	mov rdi, QWORD [rbp + 24]
	mov rax, 60
	syscall
	pop rbp
	ret
	mov rax, 'H'
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
