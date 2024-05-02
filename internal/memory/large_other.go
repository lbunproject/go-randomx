//go:build openbsd || netbsd || dragonfly || darwin || ios || !unix || purego

package memory

var LargePageNoMemoryErr error

// NewLargePageAllocator Not supported in platform
func NewLargePageAllocator() Allocator {
	return nil
}
