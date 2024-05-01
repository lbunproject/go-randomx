/*
Copyright (c) 2019 DERO Foundation. All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package aes

import (
	"git.gammaspectra.live/P2Pool/go-randomx/v3/internal/keys"
)

var fillAes4Rx4Keys0 = [4][4]uint32{
	keys.AesGenerator4R_Keys[0],
	keys.AesGenerator4R_Keys[0],
	keys.AesGenerator4R_Keys[4],
	keys.AesGenerator4R_Keys[4],
}
var fillAes4Rx4Keys1 = [4][4]uint32{
	keys.AesGenerator4R_Keys[1],
	keys.AesGenerator4R_Keys[1],
	keys.AesGenerator4R_Keys[5],
	keys.AesGenerator4R_Keys[5],
}
var fillAes4Rx4Keys2 = [4][4]uint32{
	keys.AesGenerator4R_Keys[2],
	keys.AesGenerator4R_Keys[2],
	keys.AesGenerator4R_Keys[6],
	keys.AesGenerator4R_Keys[6],
}
var fillAes4Rx4Keys3 = [4][4]uint32{
	keys.AesGenerator4R_Keys[3],
	keys.AesGenerator4R_Keys[3],
	keys.AesGenerator4R_Keys[7],
	keys.AesGenerator4R_Keys[7],
}
