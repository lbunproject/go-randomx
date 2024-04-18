//go:build amd64 && !purego

package asm

func Cpuid(op uint32) (eax, ebx, ecx, edx uint32)
func Cpuidex(op, op2 uint32) (eax, ebx, ecx, edx uint32)
func Xgetbv(index uint32) (eax, edx uint32)
