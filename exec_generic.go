//go:build !unix || disable_jit || purego

package randomx

func (f ProgramFunc) Execute(v uintptr) {

}

func (f ProgramFunc) Close() error {
	return nil
}
