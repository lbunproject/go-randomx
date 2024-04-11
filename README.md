# RandomX (Golang Implementation)

Fork from [git.dero.io/DERO_Foundation/RandomX](https://git.dero.io/DERO_Foundation/RandomX). Also related, their [Analysis of RandomX writeup](https://medium.com/deroproject/analysis-of-randomx-dde9dfe9bbc6).

Original code failed RandomX testcases and was implemented using big.Float.

This package implements RandomX without CGO, using only Golang code, pure float64 ops and two small assembly sections to implement CFROUND modes.

All test cases pass properly.

Supports `386` `amd64` `arm64` platforms due to rounding mode set via assembly. More can be added with supporting rounding mode under _fpu_.