global _start
_start:
	mov rax, 'C'
	push rax			;; Stack position: 1
	push rax			;; Stack position: 2
	mov rax, 'H'
	push rax			;; Stack position: 3
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 'e'
	push rax			;; Stack position: 4
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 'l'
	push rax			;; Stack position: 5
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, ' '
	push rax			;; Stack position: 6
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 'l'
	push rax			;; Stack position: 7
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 'o'
	push rax			;; Stack position: 8
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 10
	push rax			;; Stack position: 9
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	push QWORD [rsp + 56]			;; Stack position: 10
	mov rax, 1
	mov rdi, 1
	mov rsi, rsp
	mov rdx, 1
	syscall
	mov rax, 60
	mov rdi, 0
	syscall
