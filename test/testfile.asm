global _start
_start:
	mov rax, 22
	push rax	;;1
	mov rax, 30
	push rax	;;2
	mov rax, 19
	push rax	;;3
	mov rax, 33
	push rax	;;4
	push QWORD [rsp + 16]	;;5
	pop rax
	push QWORD [rsp + 24]	;;5
	pop rbx
	add rax, rbx
	push rax	;;5
	pop rax
	mov rbx, 22
	add rax, rbx
	push rax	;;5
	mov rax, 60
	pop rdi
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
