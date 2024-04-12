//go:build !unix || disable_jit

package randomx

func (f ProgramFunc) Execute(rl *RegisterLine) {

}

func (f ProgramFunc) Close() error {

}
