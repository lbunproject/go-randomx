//go:build !unix || disable_jit || purego

package randomx

func (f ProgramFunc) Execute(rl *RegisterLine) {

}

func (f ProgramFunc) Close() error {
	return nil
}
