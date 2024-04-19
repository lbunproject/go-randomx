//go:build !unix || !amd64 || purego || disable_jit

package randomx

func (f SuperScalarProgramFunc) Execute(rf uintptr) {

}

// generateSuperscalarCode
func generateSuperscalarCode(scalarProgram SuperScalarProgram) SuperScalarProgramFunc {
	return nil
}
