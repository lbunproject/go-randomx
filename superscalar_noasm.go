//go:build !unix || !amd64 || purego || disable_jit

package randomx

// generateSuperscalarCode
func generateSuperscalarCode(scalarProgram SuperScalarProgram) ProgramFunc {
	return nil
}
