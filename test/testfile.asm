global _start
_start:
	mov rax, 'C'
	push rax			;; Stack position: 1
	push rax			;; Stack position: 2
	mov rax, 'b'
	push rax			;; Stack position: 3
	push rax			;; Stack position: 4
	mov rax, 'c'
	push rax			;; Stack position: 5
	push rax			;; Stack position: 6
	mov rax, 'e'
	push rax			;; Stack position: 7
	pop rax
	mov QWORD [rsp + 32], rax
	mov rax, 60
	mov rdi, 0
	syscall
