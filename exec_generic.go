//go:build !unix || disable_jit || purego

package randomx

func (f SuperScalarProgramFunc) Close() error {
	return nil
}
