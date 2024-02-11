global _start
_start:
	mov rax, 22
	push rax	;;1
	mov rax, 30
	push rax	;;2
	mov rax, 10
	mov QWORD [rsp + 8], rax
	push QWORD [rsp + 0]	;;3
	pop rax
	push QWORD [rsp + 8]	;;3
	pop rbx
	add rax, rbx
	push rax	;;3
	pop rax
	mov rbx, 22
	add rax, rbx
	push rax	;;3
	mov rax, 60
	pop rdi
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
