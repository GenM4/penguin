global _start
_start:
	mov rax, 11
	push rax			;; Stack position: 1
	mov rax, 4
	push rax			;; Stack position: 2
	mov rax, 23
	push rax			;; Stack position: 3
	push QWORD [rsp + 8]			;; Stack position: 4
	pop rax
	push QWORD [rsp + 0]			;; Stack position: 4
	pop rbx
	add rax, rbx
	push rax			;; Stack position: 4
	pop rax
	mov QWORD [rsp + 16], rax
	push QWORD [rsp + 16]			;; Stack position: 4
	pop rax
	mov rbx, 1
	add rax, rbx
	push rax			;; Stack position: 4
	pop rax
	mov QWORD [rsp + 16], rax
	push QWORD [rsp + 16]			;; Stack position: 4
	pop rax
	mov rbx, 1
	add rax, rbx
	push rax			;; Stack position: 4
	pop rax
	mov QWORD [rsp + 16], rax
	push QWORD [rsp + 16]			;; Stack position: 4
	mov rax, 60
	pop rdi
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
