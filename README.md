# RandomX (Golang Implementation)
RandomX is a proof-of-work (PoW) algorithm that is optimized for general-purpose CPUs.
RandomX uses random code execution (hence the name) together with several memory-hard techniques to minimize the efficiency advantage of specialized hardware.

---

Fork from [git.dero.io/DERO_Foundation/RandomX](https://git.dero.io/DERO_Foundation/RandomX). Also related, their [Analysis of RandomX writeup](https://medium.com/deroproject/analysis-of-randomx-dde9dfe9bbc6).

Original code failed RandomX testcases and was implemented using big.Float.

---

This package implements RandomX without CGO, using only Golang code, native float64 ops, some assembly, but with optional soft float _purego_ implementation.

All test cases pass properly.

Supports Full mode and Light mode.

For the C++ implementation and design of RandomX, see [github.com/tevador/RandomX](https://github.com/tevador/RandomX)

|        Feature        |    386     |     amd64      |  arm   |   arm64    |  mips  | mips64 | riscv64 |  wasm  |
|:---------------------:|:----------:|:--------------:|:------:|:----------:|:------:|:------:|:-------:|:------:|
|        purego         |     ✅      |       ✅        |   ✅    |     ✅      |   ✅    |   ✅    |    ✅    |   ✅    |
|       Full Mode       |     ❌      |       ✅        |   ❌    |     ✅      |   ❌    |   ✅    |    ✅    |   ❌    |
|   Float Operations    |     hw     |     **hw**     |  soft  |   **hw**   |  soft  |  soft  |  soft   |  soft  |
|    AES Operations     |    soft    |     **hw**     |  soft  |    soft    |  soft  |  soft  |  soft   |  soft  |
| Superscalar Execution |   native   | **native+jit** | native |   native   | native | native | native  | native |
|     VM Execution      | **native** | **native+jit** |  soft  | **native** |  soft  |  soft  |  soft   |  soft  |


A pure Golang implementation can be used on platforms without hard float support or via the `purego` build flag manually.

Any platform with no hard float support or when enabled manually will use soft float, using [softfloat64](https://git.gammaspectra.live/P2Pool/softfloat64). This will be very slow.

Native hard float can be added with supporting rounding mode under _asm_.

JIT only supported under Unix systems (Linux, *BSD, macOS), and can be hard-disabled via the `disable_jit` build flag, or at runtime.