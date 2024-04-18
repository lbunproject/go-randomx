//go:build !unix || disable_jit || purego

package randomx

func (f SuperScalarProgramFunc) Close() error {
	return nil
}

func (f VMProgramFunc) Close() error {
	return nil
}

func mapProgram(program []byte, size int) []byte {
	return nil
}

func mapProgramRW(execFunc []byte) {

}

func mapProgramRX(execFunc []byte) {

}

// mapProgramRWX insecure!
func mapProgramRWX(execFunc []byte) {

}
