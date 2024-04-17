# RandomX (Golang Implementation)

Fork from [git.dero.io/DERO_Foundation/RandomX](https://git.dero.io/DERO_Foundation/RandomX). Also related, their [Analysis of RandomX writeup](https://medium.com/deroproject/analysis-of-randomx-dde9dfe9bbc6).

Original code failed RandomX testcases and was implemented using big.Float.

This package implements RandomX without CGO, using only Golang code, pure float64 ops and two small assembly sections to implement CFROUND modes, with optional soft float implementation.

All test cases pass properly.

Uses minimal Go assembly due to having to set rounding mode natively. Native hard float can be added with supporting rounding mode under _asm_.

JIT is supported on a few platforms but can be hard-disabled via the `disable_jit` build flag, or at runtime.

A pure Golang implementation can be used on platforms without hard float support or via the `purego` build flag manually.

|  Platform   | Supported | Hard Float | SuperScalar JIT |      Notes       |
|:-----------:|:---------:|:----------:|:---------------:|:----------------:|
|   **386**   |     ✅     |     ✅      |        ❌        |                  |
|  **amd64**  |     ✅     |     ✅      |       ✅*        | JIT only on Unix |
|   **arm**   |    ✅*     |     ❌      |        ❌        |                  |
|  **arm64**  |     ✅     |     ✅      |        ❌        |                  |
|  **mips**   |    ✅*     |     ❌      |        ❌        |                  |
| **mips64**  |    ✅*     |     ❌      |        ❌        |                  |
| **riscv64** |    ✅*     |     ❌      |        ❌        |                  |
|  **wasm**   |    ✅*     |     ❌      |        ❌        |                  |

&ast; these platforms only support software floating point / purego and will not be performant.