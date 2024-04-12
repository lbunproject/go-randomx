//go:build !unix || !amd64 || disable_jit

package randomx

// generateSuperscalarCode
func generateSuperscalarCode(scalarProgram SuperScalarProgram) ProgramFunc {
	return nil
}
