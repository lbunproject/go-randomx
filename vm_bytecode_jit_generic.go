//go:build !unix || !amd64 || disable_jit || purego

package randomx

func (c *ByteCode) generateCode(program []byte) {

}

func (f VMProgramFunc) Execute(rf *RegisterFile, pad *ScratchPad, eMask [2]uint64) {

}
