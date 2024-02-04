global _start
_start:
	mov rax, 22
	mov rbx, 11
	sub rax, rbx
	push rax
	mov rax, 22
	mov rbx, 1
	div rbx
	push rax
	mov rax, 2
	mov rbx, 6
	mul rbx
	push rax
	pop rax, 
	pop rbx, 
	add rax, rbx
	push rax
	pop rax, 
	pop rbx, 
	add rax, rbx
	push rax
	mov rax, 60
	pop rdi
	syscall
	mov rax, 4
	mov rbx, 2
	div rbx
	push rax
	pop rax, 
	mov rbx, 3
	add rax, rbx
	push rax
	pop rax, 
	mov rbx, 6
	add rax, rbx
	push rax
	mov rax, 60
	pop rdi
	syscall
	mov rax, 33
	mov rbx, 1
	mul rbx
	push rax
	mov rax, 3
	pop rbx, 
	sub rax, rbx
	push rax
	mov rax, 60
	pop rdi
	syscall
	mov rax, 44
	mov rbx, 11
	div rbx
	push rax
	mov rax, 60
	pop rdi
	syscall
	mov rax, 11
	push rax
	mov rax, 60
	pop rdi
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
