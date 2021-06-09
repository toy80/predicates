// Package predicates is implements arbitrary precision floating-point arithmetic and
// fast robust geometric predicates. ported form C code "predicates.c"
package predicates

// 移植方式是用GCC的预处理器展开C代码里的宏, 然后手工改为Go代码:
//  cpp ./predicates.c.txt
// 很多地方是手工编辑的, 我暂时没时间一一验证, 很可能会有疏漏. 如果您发现BUG, 欢迎提交issue或pr.

import (
	"fmt"
	"math"
	"math/rand"
	"unsafe"
)

// Float is floating-point number type
type Float = float32

// Size of the Float type
const floatSize = unsafe.Sizeof(*(*Float)(nil))

func random() int32 {
	return rand.Int31()
}

// # 432 "./predicates.c.txt"

var (
	splitter Float
	epsilon  Float

	resulterrbound                           Float
	ccwerrboundA, ccwerrboundB, ccwerrboundC Float
	o3derrboundA, o3derrboundB, o3derrboundC Float
	iccerrboundA, iccerrboundB, iccerrboundC Float
	isperrboundA, isperrboundB, isperrboundC Float
)

func doubleToString(number float64) (s string) {
	no := math.Float64bits(number)
	sign := no & 0x8000000000000000
	expo := (no >> 52) & 0x7ff
	exponent := int(expo) - 1023
	if sign != 0 {
		s = "-"
	} else {
		s = "+"
	}
	if exponent == -1023 {
		s += "0.0000000000000000000000000000000000000000000000000000_     (   )"
	} else {
		s += "1."
		bottomi := -1
		for i := 0; i < 52; i++ {
			if (no & 0x0008000000000000) != 0 {
				s += "1"
				bottomi = i
			} else {
				s += "0"
			}
			no <<= 1
		}
		s += fmt.Sprintf("_%d  (%d)", exponent, exponent-1-bottomi)
	}
	return
}

func floatToString(number float32) (s string) {
	no := math.Float32bits(number)
	sign := no & 0x80000000
	expo := (no >> 23) & 0xff
	exponent := int(expo) - 127
	if sign != 0 {
		s = ("-")
	} else {
		s = (" ")
	}
	if exponent == -127 {
		s += ("0.00000000000000000000000_     (   )")
	} else {
		s += ("1.")
		bottomi := -1
		for i := 0; i < 23; i++ {
			if (no & 0x00400000) != 0 {
				s += ("1")
				bottomi = i
			} else {
				s += ("0")
			}
			no <<= 1
		}
		s += fmt.Sprintf("_%3d  (%3d)", exponent, exponent-1-bottomi)
	}
	return
}

func realToString(x Float) string {
	if unsafe.Sizeof(x) == 4 {
		return floatToString(float32(x))
	} else {
		return doubleToString(float64(x))
	}
}

func expansionToString(elen int, e *Float) (s string) {
	for i := elen - 1; i >= 0; i-- {
		s += realToString(*(*Float)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + floatSize*uintptr(i))))
		if i > 0 {
			s += " +\n"
		} else {
			s += "\n"
		}
	}
	return
}

func narrowRealRand() (x Float) {
	if unsafe.Sizeof(x) == 8 {
		return Float(narrowDoubleRand())
	}
	return Float(narrowFloatRand())
}

func realRand() (x Float) {
	if unsafe.Sizeof(x) == 8 {
		return Float(doubleRand())
	}
	return Float(floatRand())
}

// # 567 "./predicates.c.txt"
func doubleRand() float64 {
	var result float64
	var expo float64
	var a, b, c int32
	var i int32

	a = random()
	b = random()
	c = random()
	result = (float64)(a-1073741824)*8388608.0 + (float64)(b>>8)
	for i, expo = 512, 2; i <= 131072; i, expo = i*2, expo*expo {
		if (c & i) != 0 {
			result = result * expo
		}
	}
	return result
}

// # 593 "./predicates.c.txt"
func narrowDoubleRand() float64 {
	var result float64
	var expo float64
	var a, b, c int32
	var i int32

	a = random()
	b = random()
	c = random()
	result = (float64)(a-1073741824)*8388608.0 + (float64)(b>>8)
	for i, expo = 512, 2; i <= 2048; i, expo = i*2, expo*expo {
		if (c & i) != 0 {
			result = result * expo
		}
	}
	return result
}

func uniformDoubleRand() float64 {
	var result float64
	var a, b int32

	a = random()
	b = random()
	result = (float64)(a-1073741824)*8388608.0 + (float64)(b>>8)
	return result
}

// # 636 "./predicates.c.txt"
func floatRand() float32 {
	var result float32
	var expo float32
	var a, c int32
	var i int32

	a = random()
	c = random()
	result = (float32)((a - 1073741824) >> 6)
	for i, expo = 512, 2; i <= 16384; i, expo = i*2, expo*expo {
		if (c & i) != 0 {
			result = result * expo
		}
	}
	return result
}

// # 661 "./predicates.c.txt"
func narrowFloatRand() float32 {
	var result float32
	var expo float32
	var a, c int32
	var i int32

	a = random()
	c = random()
	result = (float32)((a - 1073741824) >> 6)
	for i, expo = 512, 2; i <= 2048; i, expo = i*2, expo*expo {
		if (c & i) != 0 {
			result = result * expo
		}
	}
	return result
}

func uniformFloatRand() float32 {
	var result float32
	var a int32 = random()
	result = (float32)((a - 1073741824) >> 6)
	return result
}

// # 714 "./predicates.c.txt"
func init() {
	var half Float
	var check, lastcheck Float
	var every_other bool

	every_other = true
	half = 0.5
	epsilon = 1.0
	splitter = 1.0
	check = 1.0

	for {
		lastcheck = check
		epsilon = epsilon * half
		if every_other {
			splitter = splitter * 2.0
		}
		every_other = !every_other
		check = 1.0 + epsilon
		if !((check != 1.0) && (check != lastcheck)) {
			break
		}
	}
	splitter = splitter + 1.0

	resulterrbound = (3.0 + 8.0*epsilon) * epsilon
	ccwerrboundA = (3.0 + 16.0*epsilon) * epsilon
	ccwerrboundB = (2.0 + 12.0*epsilon) * epsilon
	ccwerrboundC = (9.0 + 64.0*epsilon) * epsilon * epsilon
	o3derrboundA = (7.0 + 56.0*epsilon) * epsilon
	o3derrboundB = (3.0 + 28.0*epsilon) * epsilon
	o3derrboundC = (26.0 + 288.0*epsilon) * epsilon * epsilon
	iccerrboundA = (10.0 + 96.0*epsilon) * epsilon
	iccerrboundB = (4.0 + 48.0*epsilon) * epsilon
	iccerrboundC = (44.0 + 576.0*epsilon) * epsilon * epsilon
	isperrboundA = (16.0 + 224.0*epsilon) * epsilon
	isperrboundB = (5.0 + 72.0*epsilon) * epsilon
	isperrboundC = (71.0 + 1408.0*epsilon) * epsilon * epsilon

	ensureOrient2dWorks()
}

func isSamePred(a, b Float) bool {
	return a == b || a > 0 && b > 0 || a < 0 && b < 0
}

func ensureOrient2dWorks() {
	var te, tn, tf Float
	pa := [2]Float{0, 0}
	pb := [2]Float{2, 3}
	var pc [2]Float
	if unsafe.Sizeof(*(*Float)(nil)) == 8 {
		pc = [2]Float{2e+08, 2.9999999989999914e+08}
	} else {
		pc = [2]Float{20000, 29999.896}
	}
	te = Orient2dExact(pa, pb, pc)
	tn = Orient2d(pa, pb, pc)
	tf = Orient2dFast(pa, pb, pc)

	if !isSamePred(te, tn) || isSamePred(te, tf) {
		panic("init predicates failed!")
	}

}

// # 770 "./predicates.c.txt"
func GrowExpansion(elen int, e *Float, b Float, h *Float) int {
	var Q Float
	var Qnew Float
	var eindex int
	var enow Float
	var bvirt Float
	var avirt, bround, around Float

	Q = b
	for eindex = 0; eindex < elen; eindex++ {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex)))) // enow = e[eindex]
		Qnew = (Float)(Q + enow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = enow - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex)))) = around + bround //h[eindex] = around + bround
		Q = Qnew
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex)))) = Q // h[eindex] = Q
	return eindex + 1
}

// # 803 "./predicates.c.txt"
func GrowExpansionZeroElim(elen int, e *Float, b Float, h *Float) int {
	var Q, hh Float
	var Qnew Float
	var eindex, hindex int
	var enow Float
	var bvirt Float
	var avirt, bround, around Float

	hindex = 0
	Q = b
	for eindex = 0; eindex < elen; eindex++ {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex)))) // enow = e[eindex]
		Qnew = (Float)(Q + enow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = enow - bvirt
		around = Q - avirt
		hh = around + bround
		Q = Qnew
		if hh != 0.0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	if (Q != 0.0) || (hindex == 0) {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
		hindex++
	}
	return hindex
}

// # 841 "./predicates.c.txt"
func ExpansionSum(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q Float
	var Qnew Float
	var findex, hindex, hlast int
	var hnow Float
	var bvirt Float
	var avirt, bround, around Float

	Q = *f
	for hindex = 0; hindex < elen; hindex++ {
		hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(hindex))))
		Qnew = (Float)(Q + hnow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = hnow - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		Q = Qnew
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
	hlast = hindex
	for findex = 1; findex < flen; findex++ {
		Q = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		for hindex = findex; hindex <= hlast; hindex++ {
			hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex))))
			Qnew = (Float)(Q + hnow)
			bvirt = (Float)(Qnew - Q)
			avirt = Qnew - bvirt
			bround = hnow - bvirt
			around = Q - avirt
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
			Q = Qnew
		}
		hlast++
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hlast)))) = Q // h[hlast] = Q
	}
	return hlast + 1
}

// # 885 "./predicates.c.txt"
func ExpansionSumZeroElim1(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q Float
	var Qnew Float
	var index, findex, hindex, hlast int
	var hnow Float
	var bvirt Float
	var avirt, bround, around Float

	Q = *f
	for hindex = 0; hindex < elen; hindex++ {
		hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(hindex))))
		Qnew = (Float)(Q + hnow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = hnow - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		Q = Qnew
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
	hlast = hindex
	for findex = 1; findex < flen; findex++ {
		Q = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		for hindex = findex; hindex <= hlast; hindex++ {
			hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex))))
			Qnew = (Float)(Q + hnow)
			bvirt = (Float)(Qnew - Q)
			avirt = Qnew - bvirt
			bround = hnow - bvirt
			around = Q - avirt
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
			Q = Qnew
		}
		hlast++
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hlast)))) = Q // h[hlast] = Q
	}
	hindex = -1
	for index = 0; index <= hlast; index++ {
		hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(index))))
		if hnow != 0.0 {
			hindex++
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hnow // h[hindex] = hnow
		}
	}
	if hindex == -1 {
		return 1
	} else {
		return hindex + 1
	}
}

// # 940 "./predicates.c.txt"
func ExpansionSumZeroElim2(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q, hh Float
	var Qnew Float
	var eindex, findex, hindex, hlast int
	var enow Float
	var bvirt Float
	var avirt, bround, around Float

	hindex = 0
	Q = *f
	for eindex = 0; eindex < elen; eindex++ {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		Qnew = (Float)(Q + enow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = enow - bvirt
		around = Q - avirt
		hh = around + bround
		Q = Qnew
		if hh != 0.0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
	hlast = hindex
	for findex = 1; findex < flen; findex++ {
		hindex = 0
		Q = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		for eindex = 0; eindex <= hlast; eindex++ {
			enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(eindex))))
			Qnew = (Float)(Q + enow)
			bvirt = (Float)(Qnew - Q)
			avirt = Qnew - bvirt
			bround = enow - bvirt
			around = Q - avirt
			hh = around + bround
			Q = Qnew
			if hh != 0 {
				*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
				hindex++
			}
		}
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
		hlast = hindex
	}
	return hlast + 1
}

// # 992 "./predicates.c.txt"
func FastExpansionSum(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q Float
	var Qnew Float
	var bvirt Float
	var avirt, bround, around Float
	var eindex, findex, hindex int
	var enow, fnow Float

	enow = *e
	fnow = *f
	// eindex = findex = 0;
	if (fnow > enow) == (fnow > -enow) {
		Q = enow
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		Q = fnow
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	hindex = 0
	if (eindex < elen) && (findex < flen) {
		if (fnow > enow) == (fnow > -enow) {
			Qnew = (Float)(enow + Q)
			bvirt = Qnew - enow
			*h = Q - bvirt // h[0] = Q - bvirt
			eindex++
			enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		} else {
			Qnew = (Float)(fnow + Q)
			bvirt = Qnew - fnow
			*h = Q - bvirt // h[0] = Q - bvirt
			findex++
			fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		}
		Q = Qnew
		hindex = 1
		for (eindex < elen) && (findex < flen) {
			if (fnow > enow) == (fnow > -enow) {
				Qnew = (Float)(Q + enow)
				bvirt = (Float)(Qnew - Q)
				avirt = Qnew - bvirt
				bround = enow - bvirt
				around = Q - avirt
				*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
				eindex++
				enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
			} else {
				Qnew = (Float)(Q + fnow)
				bvirt = (Float)(Qnew - Q)
				avirt = Qnew - bvirt
				bround = fnow - bvirt
				around = Q - avirt
				*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
				findex++
				fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
			}
			Q = Qnew
			hindex++
		}
	}
	for eindex < elen {
		Qnew = (Float)(Q + enow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = enow - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		Q = Qnew
		hindex++
	}
	for findex < flen {
		Qnew = (Float)(Q + fnow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = fnow - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		Q = Qnew
		hindex++
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
	return hindex + 1
}

// # 1065 "./predicates.c.txt"
func FastExpansionSumZeroElim(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q Float
	var Qnew Float
	var hh Float
	var bvirt Float
	var avirt, bround, around Float
	var eindex, findex, hindex int
	var enow, fnow Float

	enow = *e
	fnow = *f
	// eindex = findex = 0;
	if (fnow > enow) == (fnow > -enow) {
		Q = enow
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		Q = fnow
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	hindex = 0
	if (eindex < elen) && (findex < flen) {
		if (fnow > enow) == (fnow > -enow) {
			Qnew = (Float)(enow + Q)
			bvirt = Qnew - enow
			hh = Q - bvirt
			eindex++
			enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		} else {
			Qnew = (Float)(fnow + Q)
			bvirt = Qnew - fnow
			hh = Q - bvirt
			findex++
			fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		}
		Q = Qnew
		if hh != 0.0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
		for (eindex < elen) && (findex < flen) {
			if (fnow > enow) == (fnow > -enow) {
				Qnew = (Float)(Q + enow)
				bvirt = (Float)(Qnew - Q)
				avirt = Qnew - bvirt
				bround = enow - bvirt
				around = Q - avirt
				hh = around + bround
				eindex++
				enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
			} else {
				Qnew = (Float)(Q + fnow)
				bvirt = (Float)(Qnew - Q)
				avirt = Qnew - bvirt
				bround = fnow - bvirt
				around = Q - avirt
				hh = around + bround
				findex++
				fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
			}
			Q = Qnew
			if hh != 0.0 {
				*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
				hindex++
			}
		}
	}
	for eindex < elen {
		Qnew = (Float)(Q + enow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = enow - bvirt
		around = Q - avirt
		hh = around + bround
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		Q = Qnew
		if hh != 0.0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	for findex < flen {
		Qnew = (Float)(Q + fnow)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = fnow - bvirt
		around = Q - avirt
		hh = around + bround
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		Q = Qnew
		if hh != 0.0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	if (Q != 0.0) || (hindex == 0) {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
		hindex++
	}
	return hindex
}

// # 1145 "./predicates.c.txt"
func LinearExpansionSum(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q, q Float
	var Qnew Float
	var R Float
	var bvirt Float
	var avirt, bround, around Float
	var eindex, findex, hindex int
	var enow, fnow Float
	var g0 Float

	enow = *e
	fnow = *f
	// eindex = findex = 0;
	if (fnow > enow) == (fnow > -enow) {
		g0 = enow
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		g0 = fnow
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	if (eindex < elen) && ((findex >= flen) ||
		((fnow > enow) == (fnow > -enow))) {
		Qnew = (Float)(enow + g0)
		bvirt = Qnew - enow
		q = g0 - bvirt
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		Qnew = (Float)(fnow + g0)
		bvirt = Qnew - fnow
		q = g0 - bvirt
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	Q = Qnew
	for hindex = 0; hindex < elen+flen-2; hindex++ {
		if (eindex < elen) && ((findex >= flen) ||
			((fnow > enow) == (fnow > -enow))) {
			R = (Float)(enow + q)
			bvirt = R - enow
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = q - bvirt // h[hindex] = q - bvirt
			eindex++
			enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		} else {
			R = (Float)(fnow + q)
			bvirt = R - fnow
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = q - bvirt // h[hindex] = q - bvirt
			findex++
			fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		}
		Qnew = (Float)(Q + R)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = R - bvirt
		around = Q - avirt
		q = around + bround
		Q = Qnew
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = q   // h[hindex] = q
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex+1)))) = Q // h[hindex+1] = Q
	return hindex + 2
}

// # 1204 "./predicates.c.txt"
func LinearExpansionSumZeroElim(elen int, e *Float, flen int, f *Float, h *Float) int {
	var Q, q, hh Float
	var Qnew Float
	var R Float
	var bvirt Float
	var avirt, bround, around Float
	var eindex, findex, hindex int
	var count int
	var enow, fnow Float
	var g0 Float

	enow = *e
	fnow = *f
	// eindex = findex = 0;
	hindex = 0
	if (fnow > enow) == (fnow > -enow) {
		g0 = enow
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		g0 = fnow
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	if (eindex < elen) && ((findex >= flen) ||
		((fnow > enow) == (fnow > -enow))) {
		Qnew = (Float)(enow + g0)
		bvirt = Qnew - enow
		q = g0 - bvirt
		eindex++
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	} else {
		Qnew = (Float)(fnow + g0)
		bvirt = Qnew - fnow
		q = g0 - bvirt
		findex++
		fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
	}
	Q = Qnew
	for count = 2; count < elen+flen; count++ {
		if (eindex < elen) && ((findex >= flen) || ((fnow > enow) == (fnow > -enow))) {
			R = (Float)(enow + q)
			bvirt = R - enow
			hh = q - bvirt
			eindex++
			enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		} else {
			R = (Float)(fnow + q)
			bvirt = R - fnow
			hh = q - bvirt
			findex++
			fnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(f)) + floatSize*uintptr(findex))))
		}
		Qnew = (Float)(Q + R)
		bvirt = (Float)(Qnew - Q)
		avirt = Qnew - bvirt
		bround = R - bvirt
		around = Q - avirt
		q = around + bround
		Q = Qnew
		if hh != 0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	if q != 0 {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = q // h[hindex] = q
		hindex++
	}
	if (Q != 0.0) || (hindex == 0) {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
		hindex++
	}
	return hindex
}

// # 1273 "./predicates.c.txt"
func ScaleExpansion(elen int, e *Float, b Float, h *Float) int {
	var Q Float
	var sum Float
	var product1 Float
	var product0 Float
	var eindex, hindex int
	var enow Float
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float

	c = (Float)(splitter * b)
	abig = (Float)(c - b)
	bhi = c - abig
	blo = b - bhi
	Q = (Float)((*e) * b)
	c = (Float)(splitter * (*e))
	abig = (Float)(c - (*e))
	ahi = c - abig
	alo = (*e) - ahi
	err1 = Q - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	(*h) = (alo * blo) - err3
	hindex = 1
	for eindex = 1; eindex < elen; eindex++ {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		product1 = (Float)(enow * b)
		c = (Float)(splitter * enow)
		abig = (Float)(c - enow)
		ahi = c - abig
		alo = enow - ahi
		err1 = product1 - (ahi * bhi)
		err2 = err1 - (alo * bhi)
		err3 = err2 - (ahi * blo)
		product0 = (alo * blo) - err3
		sum = (Float)(Q + product0)
		bvirt = (Float)(sum - Q)
		avirt = sum - bvirt
		bround = product0 - bvirt
		around = Q - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		hindex++
		Q = (Float)(product1 + sum)
		bvirt = (Float)(Q - product1)
		avirt = Q - bvirt
		bround = sum - bvirt
		around = product1 - avirt
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = around + bround // h[hindex] = around + bround
		hindex++
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
	return elen + elen
}

// # 1318 "./predicates.c.txt"
func ScaleExpansionZeroElim(elen int, e *Float, b Float, h *Float) int {
	var Q, sum Float
	var hh Float
	var product1 Float
	var product0 Float
	var eindex, hindex int
	var enow Float
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float

	c = (Float)(splitter * b)
	abig = (Float)(c - b)
	bhi = c - abig
	blo = b - bhi
	Q = (Float)((*e) * b)
	c = (Float)(splitter * (*e))
	abig = (Float)(c - (*e))
	ahi = c - abig
	alo = (*e) - ahi
	err1 = Q - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	hh = (alo * blo) - err3
	hindex = 0
	if hh != 0 {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
		hindex++
	}
	for eindex = 1; eindex < elen; eindex++ {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		product1 = (Float)(enow * b)
		c = (Float)(splitter * enow)
		abig = (Float)(c - enow)
		ahi = c - abig
		alo = enow - ahi
		err1 = product1 - (ahi * bhi)
		err2 = err1 - (alo * bhi)
		err3 = err2 - (ahi * blo)
		product0 = (alo * blo) - err3
		sum = (Float)(Q + product0)
		bvirt = (Float)(sum - Q)
		avirt = sum - bvirt
		bround = product0 - bvirt
		around = Q - avirt
		hh = around + bround
		if hh != 0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
		Q = (Float)(product1 + sum)
		bvirt = Q - product1
		hh = sum - bvirt
		if hh != 0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = hh // h[hindex] = hh
			hindex++
		}
	}
	if (Q != 0.0) || (hindex == 0) {
		*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex)))) = Q // h[hindex] = Q
		hindex++
	}
	return hindex
}

// # 1369 "./predicates.c.txt"
func Compress(elen int, e *Float, h *Float) int {
	var Q, q Float
	var Qnew Float
	var eindex, hindex int
	var bvirt Float
	var enow, hnow Float
	var top, bottom int

	bottom = elen - 1
	Q = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(bottom))))
	for eindex = elen - 2; eindex >= 0; eindex-- {
		enow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
		Qnew = (Float)(Q + enow)
		bvirt = Qnew - Q
		q = enow - bvirt
		if q != 0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(bottom)))) = Qnew // h[bottom] = Qnew
			bottom--
			Q = q
		} else {
			Q = Qnew
		}
	}
	top = 0
	for hindex = bottom + 1; hindex < elen; hindex++ {
		hnow = *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(hindex))))
		Qnew = (Float)(hnow + Q)
		bvirt = Qnew - hnow
		q = Q - bvirt
		if q != 0 {
			*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(top)))) = q // h[top] = q
			top++
		}
		Q = Qnew
	}
	*(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(h)) + floatSize*uintptr(top)))) = Q // h[top] = Q
	return top + 1
}

// # 1411 "./predicates.c.txt"
func Estimate(elen int, e *Float) Float {
	var Q Float
	var eindex int

	Q = (*e)
	for eindex = 1; eindex < elen; eindex++ {
		Q = Q + *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(e)) + floatSize*uintptr(eindex))))
	}
	return Q
}

func abs(x Float) Float {
	if x >= 0.0 {
		return x
	}
	return -x
}

// # 1449 "./predicates.c.txt"
func Orient2dFast(pa [2]Float, pb [2]Float, pc [2]Float) Float {
	var acx, bcx, acy, bcy Float

	acx = pa[0] - pc[0]
	bcx = pb[0] - pc[0]
	acy = pa[1] - pc[1]
	bcy = pb[1] - pc[1]
	return acx*bcy - acy*bcx
}

func Orient2dExact(pa [2]Float, pb [2]Float, pc [2]Float) Float {
	var axby1, axcy1, bxcy1, bxay1, cxay1, cxby1 Float
	var axby0, axcy0, bxcy0, bxay0, cxay0, cxby0 Float
	var aterms, bterms, cterms [4]Float
	var aterms3, bterms3, cterms3 Float
	var v [8]Float
	var w [12]Float
	var vlength, wlength int

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	axby1 = (Float)(pa[0] * pb[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = axby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axby0 = (alo * blo) - err3
	axcy1 = (Float)(pa[0] * pc[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = axcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axcy0 = (alo * blo) - err3
	_i = (Float)(axby0 - axcy0)
	bvirt = (Float)(axby0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axcy0
	around = axby0 - avirt
	aterms[0] = around + bround
	_j = (Float)(axby1 + _i)
	bvirt = (Float)(_j - axby1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axby1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - axcy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axcy1
	around = _0 - avirt
	aterms[1] = around + bround
	aterms3 = (Float)(_j + _i)
	bvirt = (Float)(aterms3 - _j)
	avirt = aterms3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	aterms[2] = around + bround

	aterms[3] = aterms3

	bxcy1 = (Float)(pb[0] * pc[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = bxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxcy0 = (alo * blo) - err3
	bxay1 = (Float)(pb[0] * pa[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = bxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxay0 = (alo * blo) - err3
	_i = (Float)(bxcy0 - bxay0)
	bvirt = (Float)(bxcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay0
	around = bxcy0 - avirt
	bterms[0] = around + bround
	_j = (Float)(bxcy1 + _i)
	bvirt = (Float)(_j - bxcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay1
	around = _0 - avirt
	bterms[1] = around + bround
	bterms3 = (Float)(_j + _i)
	bvirt = (Float)(bterms3 - _j)
	avirt = bterms3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bterms[2] = around + bround

	bterms[3] = bterms3

	cxay1 = (Float)(pc[0] * pa[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = cxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxay0 = (alo * blo) - err3
	cxby1 = (Float)(pc[0] * pb[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = cxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxby0 = (alo * blo) - err3
	_i = (Float)(cxay0 - cxby0)
	bvirt = (Float)(cxay0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby0
	around = cxay0 - avirt
	cterms[0] = around + bround
	_j = (Float)(cxay1 + _i)
	bvirt = (Float)(_j - cxay1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cxay1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby1
	around = _0 - avirt
	cterms[1] = around + bround
	cterms3 = (Float)(_j + _i)
	bvirt = (Float)(cterms3 - _j)
	avirt = cterms3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	cterms[2] = around + bround

	cterms[3] = cterms3

	vlength = FastExpansionSumZeroElim(4, &aterms[0], 4, &bterms[0], &v[0])
	wlength = FastExpansionSumZeroElim(vlength, &v[0], 4, &cterms[0], &w[0])

	return w[wlength-1]
}

func Orient2dSlow(pa [2]Float, pb [2]Float, pc [2]Float) Float {
	var acx, acy, bcx, bcy Float
	var acxtail, acytail Float
	var bcxtail, bcytail Float
	var negate, negatetail Float
	var axby, bxay [8]Float
	var axby7, bxay7 Float
	var deter [16]Float
	var deterlen int
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var a0hi, a0lo, a1hi, a1lo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j, _k, _l, _m, _n Float
	var _0, _1, _2 Float

	acx = (Float)(pa[0] - pc[0])
	bvirt = (Float)(pa[0] - acx)
	avirt = acx + bvirt
	bround = bvirt - pc[0]
	around = pa[0] - avirt
	acxtail = around + bround
	acy = (Float)(pa[1] - pc[1])
	bvirt = (Float)(pa[1] - acy)
	avirt = acy + bvirt
	bround = bvirt - pc[1]
	around = pa[1] - avirt
	acytail = around + bround
	bcx = (Float)(pb[0] - pc[0])
	bvirt = (Float)(pb[0] - bcx)
	avirt = bcx + bvirt
	bround = bvirt - pc[0]
	around = pb[0] - avirt
	bcxtail = around + bround
	bcy = (Float)(pb[1] - pc[1])
	bvirt = (Float)(pb[1] - bcy)
	avirt = bcy + bvirt
	bround = bvirt - pc[1]
	around = pb[1] - avirt
	bcytail = around + bround

	c = (Float)(splitter * acxtail)
	abig = (Float)(c - acxtail)
	a0hi = c - abig
	a0lo = acxtail - a0hi
	c = (Float)(splitter * bcytail)
	abig = (Float)(c - bcytail)
	bhi = c - abig
	blo = bcytail - bhi
	_i = (Float)(acxtail * bcytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * acx)
	abig = (Float)(c - acx)
	a1hi = c - abig
	a1lo = acx - a1hi
	_j = (Float)(acx * bcytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * bcy)
	abig = (Float)(c - bcy)
	bhi = c - abig
	blo = bcy - bhi
	_i = (Float)(acxtail * bcy)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(acx * bcy)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axby[5] = around + bround
	axby7 = (Float)(_m + _k)
	bvirt = (Float)(axby7 - _m)
	avirt = axby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axby[6] = around + bround

	axby[7] = axby7
	negate = -acy
	negatetail = -acytail
	c = (Float)(splitter * bcxtail)
	abig = (Float)(c - bcxtail)
	a0hi = c - abig
	a0lo = bcxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(bcxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bcx)
	abig = (Float)(c - bcx)
	a1hi = c - abig
	a1lo = bcx - a1hi
	_j = (Float)(bcx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(bcxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bcx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxay[5] = around + bround
	bxay7 = (Float)(_m + _k)
	bvirt = (Float)(bxay7 - _m)
	avirt = bxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxay[6] = around + bround

	bxay[7] = bxay7

	deterlen = FastExpansionSumZeroElim(8, &axby[0], 8, &bxay[0], &deter[0])

	return deter[deterlen-1]
}

// # 1543 "./predicates.c.txt"
func Orient2dAdapt(pa [2]Float, pb [2]Float, pc [2]Float, detsum Float) Float {
	var acx, acy, bcx, bcy Float
	var acxtail, acytail, bcxtail, bcytail Float
	var detleft, detright Float
	var detlefttail, detrighttail Float
	var det, errbound Float
	var B [4]Float
	var C1 [8]Float
	var C2 [12]Float
	var D [16]Float
	var B3 Float
	var C1length, C2length, Dlength int
	var u [4]Float
	var u3 Float
	var s1, t1 Float
	var s0, t0 Float

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	acx = (Float)(pa[0] - pc[0])
	bcx = (Float)(pb[0] - pc[0])
	acy = (Float)(pa[1] - pc[1])
	bcy = (Float)(pb[1] - pc[1])

	detleft = (Float)(acx * bcy)
	c = (Float)(splitter * acx)
	abig = (Float)(c - acx)
	ahi = c - abig
	alo = acx - ahi
	c = (Float)(splitter * bcy)
	abig = (Float)(c - bcy)
	bhi = c - abig
	blo = bcy - bhi
	err1 = detleft - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	detlefttail = (alo * blo) - err3
	detright = (Float)(acy * bcx)
	c = (Float)(splitter * acy)
	abig = (Float)(c - acy)
	ahi = c - abig
	alo = acy - ahi
	c = (Float)(splitter * bcx)
	abig = (Float)(c - bcx)
	bhi = c - abig
	blo = bcx - bhi
	err1 = detright - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	detrighttail = (alo * blo) - err3

	_i = (Float)(detlefttail - detrighttail)
	bvirt = (Float)(detlefttail - _i)
	avirt = _i + bvirt
	bround = bvirt - detrighttail
	around = detlefttail - avirt
	B[0] = around + bround
	_j = (Float)(detleft + _i)
	bvirt = (Float)(_j - detleft)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = detleft - avirt
	_0 = around + bround
	_i = (Float)(_0 - detright)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - detright
	around = _0 - avirt
	B[1] = around + bround
	B3 = (Float)(_j + _i)
	bvirt = (Float)(B3 - _j)
	avirt = B3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	B[2] = around + bround

	B[3] = B3

	det = Estimate(4, &B[0])
	errbound = ccwerrboundB * detsum
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	bvirt = (Float)(pa[0] - acx)
	avirt = acx + bvirt
	bround = bvirt - pc[0]
	around = pa[0] - avirt
	acxtail = around + bround
	bvirt = (Float)(pb[0] - bcx)
	avirt = bcx + bvirt
	bround = bvirt - pc[0]
	around = pb[0] - avirt
	bcxtail = around + bround
	bvirt = (Float)(pa[1] - acy)
	avirt = acy + bvirt
	bround = bvirt - pc[1]
	around = pa[1] - avirt
	acytail = around + bround
	bvirt = (Float)(pb[1] - bcy)
	avirt = bcy + bvirt
	bround = bvirt - pc[1]
	around = pb[1] - avirt
	bcytail = around + bround

	if (acxtail == 0.0) && (acytail == 0.0) && (bcxtail == 0.0) && (bcytail == 0.0) {
		return det
	}
	errbound = ccwerrboundC*detsum + resulterrbound*abs(det)
	det += (acx*bcytail + bcy*acxtail) -
		(acy*bcxtail + bcx*acytail)
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	s1 = (Float)(acxtail * bcy)
	c = (Float)(splitter * acxtail)
	abig = (Float)(c - acxtail)
	ahi = c - abig
	alo = acxtail - ahi
	c = (Float)(splitter * bcy)
	abig = (Float)(c - bcy)
	bhi = c - abig
	blo = bcy - bhi
	err1 = s1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	s0 = (alo * blo) - err3
	t1 = (Float)(acytail * bcx)
	c = (Float)(splitter * acytail)
	abig = (Float)(c - acytail)
	ahi = c - abig
	alo = acytail - ahi
	c = (Float)(splitter * bcx)
	abig = (Float)(c - bcx)
	bhi = c - abig
	blo = bcx - bhi
	err1 = t1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	t0 = (alo * blo) - err3
	_i = (Float)(s0 - t0)
	bvirt = (Float)(s0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t0
	around = s0 - avirt
	u[0] = around + bround
	_j = (Float)(s1 + _i)
	bvirt = (Float)(_j - s1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = s1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - t1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t1
	around = _0 - avirt
	u[1] = around + bround
	u3 = (Float)(_j + _i)
	bvirt = (Float)(u3 - _j)
	avirt = u3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	u[2] = around + bround
	u[3] = u3
	C1length = FastExpansionSumZeroElim(4, &B[0], 4, &B[0], &C1[0])

	s1 = (Float)(acx * bcytail)
	c = (Float)(splitter * acx)
	abig = (Float)(c - acx)
	ahi = c - abig
	alo = acx - ahi
	c = (Float)(splitter * bcytail)
	abig = (Float)(c - bcytail)
	bhi = c - abig
	blo = bcytail - bhi
	err1 = s1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	s0 = (alo * blo) - err3
	t1 = (Float)(acy * bcxtail)
	c = (Float)(splitter * acy)
	abig = (Float)(c - acy)
	ahi = c - abig
	alo = acy - ahi
	c = (Float)(splitter * bcxtail)
	abig = (Float)(c - bcxtail)
	bhi = c - abig
	blo = bcxtail - bhi
	err1 = t1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	t0 = (alo * blo) - err3
	_i = (Float)(s0 - t0)
	bvirt = (Float)(s0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t0
	around = s0 - avirt
	u[0] = around + bround
	_j = (Float)(s1 + _i)
	bvirt = (Float)(_j - s1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = s1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - t1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t1
	around = _0 - avirt
	u[1] = around + bround
	u3 = (Float)(_j + _i)
	bvirt = (Float)(u3 - _j)
	avirt = u3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	u[2] = around + bround
	u[3] = u3
	C2length = FastExpansionSumZeroElim(C1length, &C1[0], 4, &u[0], &C2[0])

	s1 = (Float)(acxtail * bcytail)
	c = (Float)(splitter * acxtail)
	abig = (Float)(c - acxtail)
	ahi = c - abig
	alo = acxtail - ahi
	c = (Float)(splitter * bcytail)
	abig = (Float)(c - bcytail)
	bhi = c - abig
	blo = bcytail - bhi
	err1 = s1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	s0 = (alo * blo) - err3
	t1 = (Float)(acytail * bcxtail)
	c = (Float)(splitter * acytail)
	abig = (Float)(c - acytail)
	ahi = c - abig
	alo = acytail - ahi
	c = (Float)(splitter * bcxtail)
	abig = (Float)(c - bcxtail)
	bhi = c - abig
	blo = bcxtail - bhi
	err1 = t1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	t0 = (alo * blo) - err3
	_i = (Float)(s0 - t0)
	bvirt = (Float)(s0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t0
	around = s0 - avirt
	u[0] = around + bround
	_j = (Float)(s1 + _i)
	bvirt = (Float)(_j - s1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = s1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - t1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - t1
	around = _0 - avirt
	u[1] = around + bround
	u3 = (Float)(_j + _i)
	bvirt = (Float)(u3 - _j)
	avirt = u3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	u[2] = around + bround
	u[3] = u3
	Dlength = FastExpansionSumZeroElim(C2length, &C2[0], 4, &u[0], &D[0])

	return (D[Dlength-1])
}

/*****************************************************************************/
/*                                                                           */
/*  orient2dfast()   Approximate 2D orientation test.  Nonrobust.            */
/*  orient2dexact()   Exact 2D orientation test.  Robust.                    */
/*  orient2dslow()   Another exact 2D orientation test.  Robust.             */
/*  orient2d()   Adaptive exact 2D orientation test.  Robust.                */
/*                                                                           */
/*               Return a positive value if the points pa, pb, and pc occur  */
/*               in counterclockwise order; a negative value if they occur   */
/*               in clockwise order; and zero if they are collinear.  The    */
/*               result is also a rough approximation of twice the signed    */
/*               area of the triangle defined by the three points.           */
/*                                                                           */
/*  Only the first and last routine should be used; the middle two are for   */
/*  timings.                                                                 */
/*                                                                           */
/*  The last three use exact arithmetic to ensure a correct answer.  The     */
/*  result returned is the determinant of a matrix.  In orient2d() only,     */
/*  this determinant is computed adaptively, in the sense that exact         */
/*  arithmetic is used only to the degree it is needed to ensure that the    */
/*  returned value has the correct sign.  Hence, orient2d() is usually quite */
/*  fast, but will run more slowly when the input points are collinear or    */
/*  nearly so.                                                               */
/*                                                                           */
/*****************************************************************************/
func Orient2d(pa [2]Float, pb [2]Float, pc [2]Float) Float {
	var detleft, detright, det Float
	var detsum, errbound Float

	detleft = (pa[0] - pc[0]) * (pb[1] - pc[1])
	detright = (pa[1] - pc[1]) * (pb[0] - pc[0])
	det = detleft - detright

	if detleft > 0.0 {
		if detright <= 0.0 {
			return det
		} else {
			detsum = detleft + detright
		}
	} else if detleft < 0.0 {
		if detright >= 0.0 {
			return det
		} else {
			detsum = -detleft - detright
		}
	} else {
		return det
	}

	errbound = ccwerrboundA * detsum
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	return Orient2dAdapt(pa, pb, pc, detsum)
}

// # 1685 "./predicates.c.txt"
func Orient3dFast(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float) Float {
	var adx, bdx, cdx Float
	var ady, bdy, cdy Float
	var adz, bdz, cdz Float

	adx = pa[0] - pd[0]
	bdx = pb[0] - pd[0]
	cdx = pc[0] - pd[0]
	ady = pa[1] - pd[1]
	bdy = pb[1] - pd[1]
	cdy = pc[1] - pd[1]
	adz = pa[2] - pd[2]
	bdz = pb[2] - pd[2]
	cdz = pc[2] - pd[2]

	return adx*(bdy*cdz-bdz*cdy) +
		bdx*(cdy*adz-cdz*ady) +
		cdx*(ady*bdz-adz*bdy)
}

func Orient3dExact(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float) Float {
	var axby1, bxcy1, cxdy1, dxay1, axcy1, bxdy1 Float
	var bxay1, cxby1, dxcy1, axdy1, cxay1, dxby1 Float
	var axby0, bxcy0, cxdy0, dxay0, axcy0, bxdy0 Float
	var bxay0, cxby0, dxcy0, axdy0, cxay0, dxby0 Float
	var ab, bc, cd, da, ac, bd [4]Float
	var temp8 [8]Float
	var templen int
	var abc, bcd, cda, dab [12]Float
	var abclen, bcdlen, cdalen, dablen int
	var adet, bdet, cdet, ddet [24]Float
	var alen, blen, clen, dlen int
	var abdet, cddet [48]Float
	var ablen, cdlen int
	var deter [96]Float
	var deterlen int
	var i int

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	axby1 = (Float)(pa[0] * pb[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = axby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axby0 = (alo * blo) - err3
	bxay1 = (Float)(pb[0] * pa[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = bxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxay0 = (alo * blo) - err3
	_i = (Float)(axby0 - bxay0)
	bvirt = (Float)(axby0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay0
	around = axby0 - avirt
	ab[0] = around + bround
	_j = (Float)(axby1 + _i)
	bvirt = (Float)(_j - axby1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axby1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay1
	around = _0 - avirt
	ab[1] = around + bround
	ab[3] = (Float)(_j + _i)
	bvirt = (Float)(ab[3] - _j)
	avirt = ab[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround

	bxcy1 = (Float)(pb[0] * pc[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = bxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxcy0 = (alo * blo) - err3
	cxby1 = (Float)(pc[0] * pb[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = cxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxby0 = (alo * blo) - err3
	_i = (Float)(bxcy0 - cxby0)
	bvirt = (Float)(bxcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby0
	around = bxcy0 - avirt
	bc[0] = around + bround
	_j = (Float)(bxcy1 + _i)
	bvirt = (Float)(_j - bxcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby1
	around = _0 - avirt
	bc[1] = around + bround
	bc[3] = (Float)(_j + _i)
	bvirt = (Float)(bc[3] - _j)
	avirt = bc[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround

	cxdy1 = (Float)(pc[0] * pd[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = cxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxdy0 = (alo * blo) - err3
	dxcy1 = (Float)(pd[0] * pc[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = dxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxcy0 = (alo * blo) - err3
	_i = (Float)(cxdy0 - dxcy0)
	bvirt = (Float)(cxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy0
	around = cxdy0 - avirt
	cd[0] = around + bround
	_j = (Float)(cxdy1 + _i)
	bvirt = (Float)(_j - cxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxcy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy1
	around = _0 - avirt
	cd[1] = around + bround
	cd[3] = (Float)(_j + _i)
	bvirt = (Float)(cd[3] - _j)
	avirt = cd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	cd[2] = around + bround

	dxay1 = (Float)(pd[0] * pa[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = dxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxay0 = (alo * blo) - err3
	axdy1 = (Float)(pa[0] * pd[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = axdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axdy0 = (alo * blo) - err3
	_i = (Float)(dxay0 - axdy0)
	bvirt = (Float)(dxay0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy0
	around = dxay0 - avirt
	da[0] = around + bround
	_j = (Float)(dxay1 + _i)
	bvirt = (Float)(_j - dxay1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = dxay1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - axdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy1
	around = _0 - avirt
	da[1] = around + bround
	da[3] = (Float)(_j + _i)
	bvirt = (Float)(da[3] - _j)
	avirt = da[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	da[2] = around + bround

	axcy1 = (Float)(pa[0] * pc[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = axcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axcy0 = (alo * blo) - err3
	cxay1 = (Float)(pc[0] * pa[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = cxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxay0 = (alo * blo) - err3
	_i = (Float)(axcy0 - cxay0)
	bvirt = (Float)(axcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay0
	around = axcy0 - avirt
	ac[0] = around + bround
	_j = (Float)(axcy1 + _i)
	bvirt = (Float)(_j - axcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay1
	around = _0 - avirt
	ac[1] = around + bround
	ac[3] = (Float)(_j + _i)
	bvirt = (Float)(ac[3] - _j)
	avirt = ac[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ac[2] = around + bround

	bxdy1 = (Float)(pb[0] * pd[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = bxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxdy0 = (alo * blo) - err3
	dxby1 = (Float)(pd[0] * pb[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = dxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxby0 = (alo * blo) - err3
	_i = (Float)(bxdy0 - dxby0)
	bvirt = (Float)(bxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby0
	around = bxdy0 - avirt
	bd[0] = around + bround
	_j = (Float)(bxdy1 + _i)
	bvirt = (Float)(_j - bxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby1
	around = _0 - avirt
	bd[1] = around + bround
	bd[3] = (Float)(_j + _i)
	bvirt = (Float)(bd[3] - _j)
	avirt = bd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bd[2] = around + bround

	templen = FastExpansionSumZeroElim(4, &cd[0], 4, &da[0], &temp8[0])
	cdalen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &ac[0], &cda[0])
	templen = FastExpansionSumZeroElim(4, &da[0], 4, &ab[0], &temp8[0])
	dablen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &bd[0], &dab[0])
	for i = 0; i < 4; i++ {
		bd[i] = -bd[i]
		ac[i] = -ac[i]
	}
	templen = FastExpansionSumZeroElim(4, &ab[0], 4, &bc[0], &temp8[0])
	abclen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &ac[0], &abc[0])
	templen = FastExpansionSumZeroElim(4, &bc[0], 4, &cd[0], &temp8[0])
	bcdlen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &bd[0], &bcd[0])

	alen = ScaleExpansionZeroElim(bcdlen, &bcd[0], pa[2], &adet[0])
	blen = ScaleExpansionZeroElim(cdalen, &cda[0], -pb[2], &bdet[0])
	clen = ScaleExpansionZeroElim(dablen, &dab[0], pc[2], &cdet[0])
	dlen = ScaleExpansionZeroElim(abclen, &abc[0], -pd[2], &ddet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	cdlen = FastExpansionSumZeroElim(clen, &cdet[0], dlen, &ddet[0], &cddet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], cdlen, &cddet[0], &deter[0])

	return deter[deterlen-1]
}

func Orient3dSlow(pa, pb, pc, pd [3]Float) Float {
	var adx, ady, adz, bdx, bdy, bdz, cdx, cdy, cdz Float
	var adxtail, adytail, adztail Float
	var bdxtail, bdytail, bdztail Float
	var cdxtail, cdytail, cdztail Float
	var negate, negatetail Float
	var axby7, bxcy7, axcy7, bxay7, cxby7, cxay7 Float
	var axby, bxcy, axcy, bxay, cxby, cxay [8]Float
	var temp16 [16]Float
	var temp32, temp32t [32]Float
	var temp16len, temp32len, temp32tlen int
	var adet, bdet, cdet [64]Float
	var alen, blen, clen int
	var abdet [128]Float
	var ablen int
	var deter [192]Float
	var deterlen int
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var a0hi, a0lo, a1hi, a1lo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j, _k, _l, _m, _n Float
	var _0, _1, _2 Float

	adx = (Float)(pa[0] - pd[0])
	bvirt = (Float)(pa[0] - adx)
	avirt = adx + bvirt
	bround = bvirt - pd[0]
	around = pa[0] - avirt
	adxtail = around + bround
	ady = (Float)(pa[1] - pd[1])
	bvirt = (Float)(pa[1] - ady)
	avirt = ady + bvirt
	bround = bvirt - pd[1]
	around = pa[1] - avirt
	adytail = around + bround
	adz = (Float)(pa[2] - pd[2])
	bvirt = (Float)(pa[2] - adz)
	avirt = adz + bvirt
	bround = bvirt - pd[2]
	around = pa[2] - avirt
	adztail = around + bround
	bdx = (Float)(pb[0] - pd[0])
	bvirt = (Float)(pb[0] - bdx)
	avirt = bdx + bvirt
	bround = bvirt - pd[0]
	around = pb[0] - avirt
	bdxtail = around + bround
	bdy = (Float)(pb[1] - pd[1])
	bvirt = (Float)(pb[1] - bdy)
	avirt = bdy + bvirt
	bround = bvirt - pd[1]
	around = pb[1] - avirt
	bdytail = around + bround
	bdz = (Float)(pb[2] - pd[2])
	bvirt = (Float)(pb[2] - bdz)
	avirt = bdz + bvirt
	bround = bvirt - pd[2]
	around = pb[2] - avirt
	bdztail = around + bround
	cdx = (Float)(pc[0] - pd[0])
	bvirt = (Float)(pc[0] - cdx)
	avirt = cdx + bvirt
	bround = bvirt - pd[0]
	around = pc[0] - avirt
	cdxtail = around + bround
	cdy = (Float)(pc[1] - pd[1])
	bvirt = (Float)(pc[1] - cdy)
	avirt = cdy + bvirt
	bround = bvirt - pd[1]
	around = pc[1] - avirt
	cdytail = around + bround
	cdz = (Float)(pc[2] - pd[2])
	bvirt = (Float)(pc[2] - cdz)
	avirt = cdz + bvirt
	bround = bvirt - pd[2]
	around = pc[2] - avirt
	cdztail = around + bround

	c = (Float)(splitter * adxtail)
	abig = (Float)(c - adxtail)
	a0hi = c - abig
	a0lo = adxtail - a0hi
	c = (Float)(splitter * bdytail)
	abig = (Float)(c - bdytail)
	bhi = c - abig
	blo = bdytail - bhi
	_i = (Float)(adxtail * bdytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	a1hi = c - abig
	a1lo = adx - a1hi
	_j = (Float)(adx * bdytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	_i = (Float)(adxtail * bdy)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(adx * bdy)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axby[5] = around + bround
	axby7 = (Float)(_m + _k)
	bvirt = (Float)(axby7 - _m)
	avirt = axby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axby[6] = around + bround

	axby[7] = axby7
	negate = -ady
	negatetail = -adytail
	c = (Float)(splitter * bdxtail)
	abig = (Float)(c - bdxtail)
	a0hi = c - abig
	a0lo = bdxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(bdxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	a1hi = c - abig
	a1lo = bdx - a1hi
	_j = (Float)(bdx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(bdxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bdx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxay[5] = around + bround
	bxay7 = (Float)(_m + _k)
	bvirt = (Float)(bxay7 - _m)
	avirt = bxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxay[6] = around + bround

	bxay[7] = bxay7
	c = (Float)(splitter * bdxtail)
	abig = (Float)(c - bdxtail)
	a0hi = c - abig
	a0lo = bdxtail - a0hi
	c = (Float)(splitter * cdytail)
	abig = (Float)(c - cdytail)
	bhi = c - abig
	blo = cdytail - bhi
	_i = (Float)(bdxtail * cdytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	a1hi = c - abig
	a1lo = bdx - a1hi
	_j = (Float)(bdx * cdytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	_i = (Float)(bdxtail * cdy)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bdx * cdy)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxcy[5] = around + bround
	bxcy7 = (Float)(_m + _k)
	bvirt = (Float)(bxcy7 - _m)
	avirt = bxcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxcy[6] = around + bround

	bxcy[7] = bxcy7
	negate = -bdy
	negatetail = -bdytail
	c = (Float)(splitter * cdxtail)
	abig = (Float)(c - cdxtail)
	a0hi = c - abig
	a0lo = cdxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(cdxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	a1hi = c - abig
	a1lo = cdx - a1hi
	_j = (Float)(cdx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(cdxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cdx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxby[5] = around + bround
	cxby7 = (Float)(_m + _k)
	bvirt = (Float)(cxby7 - _m)
	avirt = cxby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxby[6] = around + bround

	cxby[7] = cxby7
	c = (Float)(splitter * cdxtail)
	abig = (Float)(c - cdxtail)
	a0hi = c - abig
	a0lo = cdxtail - a0hi
	c = (Float)(splitter * adytail)
	abig = (Float)(c - adytail)
	bhi = c - abig
	blo = adytail - bhi
	_i = (Float)(cdxtail * adytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	a1hi = c - abig
	a1lo = cdx - a1hi
	_j = (Float)(cdx * adytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	_i = (Float)(cdxtail * ady)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cdx * ady)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxay[5] = around + bround
	cxay7 = (Float)(_m + _k)
	bvirt = (Float)(cxay7 - _m)
	avirt = cxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxay[6] = around + bround

	cxay[7] = cxay7
	negate = -cdy
	negatetail = -cdytail
	c = (Float)(splitter * adxtail)
	abig = (Float)(c - adxtail)
	a0hi = c - abig
	a0lo = adxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(adxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	a1hi = c - abig
	a1lo = adx - a1hi
	_j = (Float)(adx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(adxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(adx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axcy[5] = around + bround
	axcy7 = (Float)(_m + _k)
	bvirt = (Float)(axcy7 - _m)
	avirt = axcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axcy[6] = around + bround

	axcy[7] = axcy7

	temp16len = FastExpansionSumZeroElim(8, &bxcy[0], 8, &cxby[0], &temp16[0])
	temp32len = ScaleExpansionZeroElim(temp16len, &temp16[0], adz, &temp32[0])
	temp32tlen = ScaleExpansionZeroElim(temp16len, &temp16[0], adztail, &temp32t[0])
	alen = FastExpansionSumZeroElim(temp32len, &temp32[0], temp32tlen, &temp32t[0], &adet[0])

	temp16len = FastExpansionSumZeroElim(8, &cxay[0], 8, &axcy[0], &temp16[0])
	temp32len = ScaleExpansionZeroElim(temp16len, &temp16[0], bdz, &temp32[0])
	temp32tlen = ScaleExpansionZeroElim(temp16len, &temp16[0], bdztail, &temp32t[0])
	blen = FastExpansionSumZeroElim(temp32len, &temp32[0], temp32tlen, &temp32t[0], &bdet[0])

	temp16len = FastExpansionSumZeroElim(8, &axby[0], 8, &bxay[0], &temp16[0])
	temp32len = ScaleExpansionZeroElim(temp16len, &temp16[0], cdz, &temp32[0])
	temp32tlen = ScaleExpansionZeroElim(temp16len, &temp16[0], cdztail, &temp32t[0])
	clen = FastExpansionSumZeroElim(temp32len, &temp32[0], temp32tlen, &temp32t[0], &cdet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], clen, &cdet[0], &deter[0])

	return deter[deterlen-1]
}

// # 1877 "./predicates.c.txt"
func Orient3dAdapt(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float, permanent Float) Float {
	var adx, bdx, cdx, ady, bdy, cdy, adz, bdz, cdz Float
	var det, errbound Float

	var bdxcdy1, cdxbdy1, cdxady1, adxcdy1, adxbdy1, bdxady1 Float
	var bdxcdy0, cdxbdy0, cdxady0, adxcdy0, adxbdy0, bdxady0 Float
	var bc, ca, ab [4]Float
	var bc3, ca3, ab3 Float
	var adet, bdet, cdet [8]Float
	var alen, blen, clen int
	var abdet [16]Float
	var ablen int
	var finnow, finother, finswap *Float
	var fin1, fin2 [192]Float
	var finlength int

	var adxtail, bdxtail, cdxtail Float
	var adytail, bdytail, cdytail Float
	var adztail, bdztail, cdztail Float
	var at_blarge, at_clarge Float
	var bt_clarge, bt_alarge Float
	var ct_alarge, ct_blarge Float
	var at_b, at_c, bt_c, bt_a, ct_a, ct_b [4]Float
	var at_blen, at_clen, bt_clen, bt_alen, ct_alen, ct_blen int
	var bdxt_cdy1, cdxt_bdy1, cdxt_ady1 Float
	var adxt_cdy1, adxt_bdy1, bdxt_ady1 Float
	var bdxt_cdy0, cdxt_bdy0, cdxt_ady0 Float
	var adxt_cdy0, adxt_bdy0, bdxt_ady0 Float
	var bdyt_cdx1, cdyt_bdx1, cdyt_adx1 Float
	var adyt_cdx1, adyt_bdx1, bdyt_adx1 Float
	var bdyt_cdx0, cdyt_bdx0, cdyt_adx0 Float
	var adyt_cdx0, adyt_bdx0, bdyt_adx0 Float
	var bct, cat, abt [8]Float
	var bctlen, catlen, abtlen int
	var bdxt_cdyt1, cdxt_bdyt1, cdxt_adyt1 Float
	var adxt_cdyt1, adxt_bdyt1, bdxt_adyt1 Float
	var bdxt_cdyt0, cdxt_bdyt0, cdxt_adyt0 Float
	var adxt_cdyt0, adxt_bdyt0, bdxt_adyt0 Float
	var u [4]Float
	var v [12]Float
	var w [16]Float
	var u3 Float
	var vlength, wlength int
	var negate Float

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j, _k Float
	var _0 Float

	adx = (Float)(pa[0] - pd[0])
	bdx = (Float)(pb[0] - pd[0])
	cdx = (Float)(pc[0] - pd[0])
	ady = (Float)(pa[1] - pd[1])
	bdy = (Float)(pb[1] - pd[1])
	cdy = (Float)(pc[1] - pd[1])
	adz = (Float)(pa[2] - pd[2])
	bdz = (Float)(pb[2] - pd[2])
	cdz = (Float)(pc[2] - pd[2])

	bdxcdy1 = (Float)(bdx * cdy)
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	ahi = c - abig
	alo = bdx - ahi
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	err1 = bdxcdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bdxcdy0 = (alo * blo) - err3
	cdxbdy1 = (Float)(cdx * bdy)
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	ahi = c - abig
	alo = cdx - ahi
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	err1 = cdxbdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cdxbdy0 = (alo * blo) - err3
	_i = (Float)(bdxcdy0 - cdxbdy0)
	bvirt = (Float)(bdxcdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cdxbdy0
	around = bdxcdy0 - avirt
	bc[0] = around + bround
	_j = (Float)(bdxcdy1 + _i)
	bvirt = (Float)(_j - bdxcdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bdxcdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cdxbdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cdxbdy1
	around = _0 - avirt
	bc[1] = around + bround
	bc3 = (Float)(_j + _i)
	bvirt = (Float)(bc3 - _j)
	avirt = bc3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround
	bc[3] = bc3
	alen = ScaleExpansionZeroElim(4, &bc[0], adz, &adet[0])

	cdxady1 = (Float)(cdx * ady)
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	ahi = c - abig
	alo = cdx - ahi
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	err1 = cdxady1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cdxady0 = (alo * blo) - err3
	adxcdy1 = (Float)(adx * cdy)
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	ahi = c - abig
	alo = adx - ahi
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	err1 = adxcdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	adxcdy0 = (alo * blo) - err3
	_i = (Float)(cdxady0 - adxcdy0)
	bvirt = (Float)(cdxady0 - _i)
	avirt = _i + bvirt
	bround = bvirt - adxcdy0
	around = cdxady0 - avirt
	ca[0] = around + bround
	_j = (Float)(cdxady1 + _i)
	bvirt = (Float)(_j - cdxady1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cdxady1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - adxcdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - adxcdy1
	around = _0 - avirt
	ca[1] = around + bround
	ca3 = (Float)(_j + _i)
	bvirt = (Float)(ca3 - _j)
	avirt = ca3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ca[2] = around + bround
	ca[3] = ca3
	blen = ScaleExpansionZeroElim(4, &ca[0], bdz, &bdet[0])

	adxbdy1 = (Float)(adx * bdy)
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	ahi = c - abig
	alo = adx - ahi
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	err1 = adxbdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	adxbdy0 = (alo * blo) - err3
	bdxady1 = (Float)(bdx * ady)
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	ahi = c - abig
	alo = bdx - ahi
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	err1 = bdxady1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bdxady0 = (alo * blo) - err3
	_i = (Float)(adxbdy0 - bdxady0)
	bvirt = (Float)(adxbdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bdxady0
	around = adxbdy0 - avirt
	ab[0] = around + bround
	_j = (Float)(adxbdy1 + _i)
	bvirt = (Float)(_j - adxbdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = adxbdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bdxady1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bdxady1
	around = _0 - avirt
	ab[1] = around + bround
	ab3 = (Float)(_j + _i)
	bvirt = (Float)(ab3 - _j)
	avirt = ab3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround
	ab[3] = ab3
	clen = ScaleExpansionZeroElim(4, &ab[0], cdz, &cdet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	finlength = FastExpansionSumZeroElim(ablen, &abdet[0], clen, &cdet[0], &fin1[0])

	det = Estimate(finlength, &fin1[0])
	errbound = o3derrboundB * permanent
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	bvirt = (Float)(pa[0] - adx)
	avirt = adx + bvirt
	bround = bvirt - pd[0]
	around = pa[0] - avirt
	adxtail = around + bround
	bvirt = (Float)(pb[0] - bdx)
	avirt = bdx + bvirt
	bround = bvirt - pd[0]
	around = pb[0] - avirt
	bdxtail = around + bround
	bvirt = (Float)(pc[0] - cdx)
	avirt = cdx + bvirt
	bround = bvirt - pd[0]
	around = pc[0] - avirt
	cdxtail = around + bround
	bvirt = (Float)(pa[1] - ady)
	avirt = ady + bvirt
	bround = bvirt - pd[1]
	around = pa[1] - avirt
	adytail = around + bround
	bvirt = (Float)(pb[1] - bdy)
	avirt = bdy + bvirt
	bround = bvirt - pd[1]
	around = pb[1] - avirt
	bdytail = around + bround
	bvirt = (Float)(pc[1] - cdy)
	avirt = cdy + bvirt
	bround = bvirt - pd[1]
	around = pc[1] - avirt
	cdytail = around + bround
	bvirt = (Float)(pa[2] - adz)
	avirt = adz + bvirt
	bround = bvirt - pd[2]
	around = pa[2] - avirt
	adztail = around + bround
	bvirt = (Float)(pb[2] - bdz)
	avirt = bdz + bvirt
	bround = bvirt - pd[2]
	around = pb[2] - avirt
	bdztail = around + bround
	bvirt = (Float)(pc[2] - cdz)
	avirt = cdz + bvirt
	bround = bvirt - pd[2]
	around = pc[2] - avirt
	cdztail = around + bround

	if (adxtail == 0.0) && (bdxtail == 0.0) && (cdxtail == 0.0) &&
		(adytail == 0.0) && (bdytail == 0.0) && (cdytail == 0.0) &&
		(adztail == 0.0) && (bdztail == 0.0) && (cdztail == 0.0) {
		return det
	}

	errbound = o3derrboundC*permanent + resulterrbound*abs(det)
	det += (adz*((bdx*cdytail+cdy*bdxtail)-
		(bdy*cdxtail+cdx*bdytail)) +
		adztail*(bdx*cdy-bdy*cdx)) +
		(bdz*((cdx*adytail+ady*cdxtail)-
			(cdy*adxtail+adx*cdytail)) +
			bdztail*(cdx*ady-cdy*adx)) +
		(cdz*((adx*bdytail+bdy*adxtail)-
			(ady*bdxtail+bdx*adytail)) +
			cdztail*(adx*bdy-ady*bdx))
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	finnow = &fin1[0]
	finother = &fin2[0]

	if adxtail == 0.0 {
		if adytail == 0.0 {
			at_b[0] = 0.0
			at_blen = 1
			at_c[0] = 0.0
			at_clen = 1
		} else {
			negate = -adytail
			at_blarge = (Float)(negate * bdx)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			bhi = c - abig
			blo = bdx - bhi
			err1 = at_blarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			at_b[0] = (alo * blo) - err3
			at_b[1] = at_blarge
			at_blen = 2
			at_clarge = (Float)(adytail * cdx)
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			ahi = c - abig
			alo = adytail - ahi
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			bhi = c - abig
			blo = cdx - bhi
			err1 = at_clarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			at_c[0] = (alo * blo) - err3
			at_c[1] = at_clarge
			at_clen = 2
		}
	} else {
		if adytail == 0.0 {
			at_blarge = (Float)(adxtail * bdy)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * bdy)
			abig = (Float)(c - bdy)
			bhi = c - abig
			blo = bdy - bhi
			err1 = at_blarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			at_b[0] = (alo * blo) - err3
			at_b[1] = at_blarge
			at_blen = 2
			negate = -adxtail
			at_clarge = (Float)(negate * cdy)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * cdy)
			abig = (Float)(c - cdy)
			bhi = c - abig
			blo = cdy - bhi
			err1 = at_clarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			at_c[0] = (alo * blo) - err3
			at_c[1] = at_clarge
			at_clen = 2
		} else {
			adxt_bdy1 = (Float)(adxtail * bdy)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * bdy)
			abig = (Float)(c - bdy)
			bhi = c - abig
			blo = bdy - bhi
			err1 = adxt_bdy1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adxt_bdy0 = (alo * blo) - err3
			adyt_bdx1 = (Float)(adytail * bdx)
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			ahi = c - abig
			alo = adytail - ahi
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			bhi = c - abig
			blo = bdx - bhi
			err1 = adyt_bdx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adyt_bdx0 = (alo * blo) - err3
			_i = (Float)(adxt_bdy0 - adyt_bdx0)
			bvirt = (Float)(adxt_bdy0 - _i)
			avirt = _i + bvirt
			bround = bvirt - adyt_bdx0
			around = adxt_bdy0 - avirt
			at_b[0] = around + bround
			_j = (Float)(adxt_bdy1 + _i)
			bvirt = (Float)(_j - adxt_bdy1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = adxt_bdy1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - adyt_bdx1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - adyt_bdx1
			around = _0 - avirt
			at_b[1] = around + bround
			at_blarge = (Float)(_j + _i)
			bvirt = (Float)(at_blarge - _j)
			avirt = at_blarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			at_b[2] = around + bround

			at_b[3] = at_blarge
			at_blen = 4
			adyt_cdx1 = (Float)(adytail * cdx)
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			ahi = c - abig
			alo = adytail - ahi
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			bhi = c - abig
			blo = cdx - bhi
			err1 = adyt_cdx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adyt_cdx0 = (alo * blo) - err3
			adxt_cdy1 = (Float)(adxtail * cdy)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * cdy)
			abig = (Float)(c - cdy)
			bhi = c - abig
			blo = cdy - bhi
			err1 = adxt_cdy1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adxt_cdy0 = (alo * blo) - err3
			_i = (Float)(adyt_cdx0 - adxt_cdy0)
			bvirt = (Float)(adyt_cdx0 - _i)
			avirt = _i + bvirt
			bround = bvirt - adxt_cdy0
			around = adyt_cdx0 - avirt
			at_c[0] = around + bround
			_j = (Float)(adyt_cdx1 + _i)
			bvirt = (Float)(_j - adyt_cdx1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = adyt_cdx1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - adxt_cdy1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - adxt_cdy1
			around = _0 - avirt
			at_c[1] = around + bround
			at_clarge = (Float)(_j + _i)
			bvirt = (Float)(at_clarge - _j)
			avirt = at_clarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			at_c[2] = around + bround

			at_c[3] = at_clarge
			at_clen = 4
		}
	}
	if bdxtail == 0.0 {
		if bdytail == 0.0 {
			bt_c[0] = 0.0
			bt_clen = 1
			bt_a[0] = 0.0
			bt_alen = 1
		} else {
			negate = -bdytail
			bt_clarge = (Float)(negate * cdx)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			bhi = c - abig
			blo = cdx - bhi
			err1 = bt_clarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bt_c[0] = (alo * blo) - err3
			bt_c[1] = bt_clarge
			bt_clen = 2
			bt_alarge = (Float)(bdytail * adx)
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			ahi = c - abig
			alo = bdytail - ahi
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			bhi = c - abig
			blo = adx - bhi
			err1 = bt_alarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bt_a[0] = (alo * blo) - err3
			bt_a[1] = bt_alarge
			bt_alen = 2
		}
	} else {
		if bdytail == 0.0 {
			bt_clarge = (Float)(bdxtail * cdy)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * cdy)
			abig = (Float)(c - cdy)
			bhi = c - abig
			blo = cdy - bhi
			err1 = bt_clarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bt_c[0] = (alo * blo) - err3
			bt_c[1] = bt_clarge
			bt_clen = 2
			negate = -bdxtail
			bt_alarge = (Float)(negate * ady)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * ady)
			abig = (Float)(c - ady)
			bhi = c - abig
			blo = ady - bhi
			err1 = bt_alarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bt_a[0] = (alo * blo) - err3
			bt_a[1] = bt_alarge
			bt_alen = 2
		} else {
			bdxt_cdy1 = (Float)(bdxtail * cdy)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * cdy)
			abig = (Float)(c - cdy)
			bhi = c - abig
			blo = cdy - bhi
			err1 = bdxt_cdy1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdxt_cdy0 = (alo * blo) - err3
			bdyt_cdx1 = (Float)(bdytail * cdx)
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			ahi = c - abig
			alo = bdytail - ahi
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			bhi = c - abig
			blo = cdx - bhi
			err1 = bdyt_cdx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdyt_cdx0 = (alo * blo) - err3
			_i = (Float)(bdxt_cdy0 - bdyt_cdx0)
			bvirt = (Float)(bdxt_cdy0 - _i)
			avirt = _i + bvirt
			bround = bvirt - bdyt_cdx0
			around = bdxt_cdy0 - avirt
			bt_c[0] = around + bround
			_j = (Float)(bdxt_cdy1 + _i)
			bvirt = (Float)(_j - bdxt_cdy1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = bdxt_cdy1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - bdyt_cdx1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - bdyt_cdx1
			around = _0 - avirt
			bt_c[1] = around + bround
			bt_clarge = (Float)(_j + _i)
			bvirt = (Float)(bt_clarge - _j)
			avirt = bt_clarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			bt_c[2] = around + bround

			bt_c[3] = bt_clarge
			bt_clen = 4
			bdyt_adx1 = (Float)(bdytail * adx)
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			ahi = c - abig
			alo = bdytail - ahi
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			bhi = c - abig
			blo = adx - bhi
			err1 = bdyt_adx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdyt_adx0 = (alo * blo) - err3
			bdxt_ady1 = (Float)(bdxtail * ady)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * ady)
			abig = (Float)(c - ady)
			bhi = c - abig
			blo = ady - bhi
			err1 = bdxt_ady1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdxt_ady0 = (alo * blo) - err3
			_i = (Float)(bdyt_adx0 - bdxt_ady0)
			bvirt = (Float)(bdyt_adx0 - _i)
			avirt = _i + bvirt
			bround = bvirt - bdxt_ady0
			around = bdyt_adx0 - avirt
			bt_a[0] = around + bround
			_j = (Float)(bdyt_adx1 + _i)
			bvirt = (Float)(_j - bdyt_adx1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = bdyt_adx1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - bdxt_ady1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - bdxt_ady1
			around = _0 - avirt
			bt_a[1] = around + bround
			bt_alarge = (Float)(_j + _i)
			bvirt = (Float)(bt_alarge - _j)
			avirt = bt_alarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			bt_a[2] = around + bround

			bt_a[3] = bt_alarge
			bt_alen = 4
		}
	}
	if cdxtail == 0.0 {
		if cdytail == 0.0 {
			ct_a[0] = 0.0
			ct_alen = 1
			ct_b[0] = 0.0
			ct_blen = 1
		} else {
			negate = -cdytail
			ct_alarge = (Float)(negate * adx)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			bhi = c - abig
			blo = adx - bhi
			err1 = ct_alarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ct_a[0] = (alo * blo) - err3
			ct_a[1] = ct_alarge
			ct_alen = 2
			ct_blarge = (Float)(cdytail * bdx)
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			ahi = c - abig
			alo = cdytail - ahi
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			bhi = c - abig
			blo = bdx - bhi
			err1 = ct_blarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ct_b[0] = (alo * blo) - err3
			ct_b[1] = ct_blarge
			ct_blen = 2
		}
	} else {
		if cdytail == 0.0 {
			ct_alarge = (Float)(cdxtail * ady)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * ady)
			abig = (Float)(c - ady)
			bhi = c - abig
			blo = ady - bhi
			err1 = ct_alarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ct_a[0] = (alo * blo) - err3
			ct_a[1] = ct_alarge
			ct_alen = 2
			negate = -cdxtail
			ct_blarge = (Float)(negate * bdy)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * bdy)
			abig = (Float)(c - bdy)
			bhi = c - abig
			blo = bdy - bhi
			err1 = ct_blarge - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ct_b[0] = (alo * blo) - err3
			ct_b[1] = ct_blarge
			ct_blen = 2
		} else {
			cdxt_ady1 = (Float)(cdxtail * ady)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * ady)
			abig = (Float)(c - ady)
			bhi = c - abig
			blo = ady - bhi
			err1 = cdxt_ady1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdxt_ady0 = (alo * blo) - err3
			cdyt_adx1 = (Float)(cdytail * adx)
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			ahi = c - abig
			alo = cdytail - ahi
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			bhi = c - abig
			blo = adx - bhi
			err1 = cdyt_adx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdyt_adx0 = (alo * blo) - err3
			_i = (Float)(cdxt_ady0 - cdyt_adx0)
			bvirt = (Float)(cdxt_ady0 - _i)
			avirt = _i + bvirt
			bround = bvirt - cdyt_adx0
			around = cdxt_ady0 - avirt
			ct_a[0] = around + bround
			_j = (Float)(cdxt_ady1 + _i)
			bvirt = (Float)(_j - cdxt_ady1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = cdxt_ady1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - cdyt_adx1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - cdyt_adx1
			around = _0 - avirt
			ct_a[1] = around + bround
			ct_alarge = (Float)(_j + _i)
			bvirt = (Float)(ct_alarge - _j)
			avirt = ct_alarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			ct_a[2] = around + bround

			ct_a[3] = ct_alarge
			ct_alen = 4
			cdyt_bdx1 = (Float)(cdytail * bdx)
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			ahi = c - abig
			alo = cdytail - ahi
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			bhi = c - abig
			blo = bdx - bhi
			err1 = cdyt_bdx1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdyt_bdx0 = (alo * blo) - err3
			cdxt_bdy1 = (Float)(cdxtail * bdy)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * bdy)
			abig = (Float)(c - bdy)
			bhi = c - abig
			blo = bdy - bhi
			err1 = cdxt_bdy1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdxt_bdy0 = (alo * blo) - err3
			_i = (Float)(cdyt_bdx0 - cdxt_bdy0)
			bvirt = (Float)(cdyt_bdx0 - _i)
			avirt = _i + bvirt
			bround = bvirt - cdxt_bdy0
			around = cdyt_bdx0 - avirt
			ct_b[0] = around + bround
			_j = (Float)(cdyt_bdx1 + _i)
			bvirt = (Float)(_j - cdyt_bdx1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = cdyt_bdx1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - cdxt_bdy1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - cdxt_bdy1
			around = _0 - avirt
			ct_b[1] = around + bround
			ct_blarge = (Float)(_j + _i)
			bvirt = (Float)(ct_blarge - _j)
			avirt = ct_blarge - bvirt
			bround = _i - bvirt
			around = _j - avirt
			ct_b[2] = around + bround

			ct_b[3] = ct_blarge
			ct_blen = 4
		}
	}

	bctlen = FastExpansionSumZeroElim(bt_clen, &bt_c[0], ct_blen, &ct_b[0], &bct[0])
	wlength = ScaleExpansionZeroElim(bctlen, &bct[0], adz, &w[0])
	finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
	finswap = finnow
	finnow = finother
	finother = finswap

	catlen = FastExpansionSumZeroElim(ct_alen, &ct_a[0], at_clen, &at_c[0], &cat[0])
	wlength = ScaleExpansionZeroElim(catlen, &cat[0], bdz, &w[0])
	finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
	finswap = finnow
	finnow = finother
	finother = finswap

	abtlen = FastExpansionSumZeroElim(at_blen, &at_b[0], bt_alen, &bt_a[0], &abt[0])
	wlength = ScaleExpansionZeroElim(abtlen, &abt[0], cdz, &w[0])
	finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
	finswap = finnow
	finnow = finother
	finother = finswap

	if adztail != 0.0 {
		vlength = ScaleExpansionZeroElim(4, &bc[0], adztail, &v[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, vlength, &v[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if bdztail != 0.0 {
		vlength = ScaleExpansionZeroElim(4, &ca[0], bdztail, &v[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, vlength, &v[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if cdztail != 0.0 {
		vlength = ScaleExpansionZeroElim(4, &ab[0], cdztail, &v[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, vlength, &v[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}

	if adxtail != 0.0 {
		if bdytail != 0.0 {
			adxt_bdyt1 = (Float)(adxtail * bdytail)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			bhi = c - abig
			blo = bdytail - bhi
			err1 = adxt_bdyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adxt_bdyt0 = (alo * blo) - err3
			c = (Float)(splitter * cdz)
			abig = (Float)(c - cdz)
			bhi = c - abig
			blo = cdz - bhi
			_i = (Float)(adxt_bdyt0 * cdz)
			c = (Float)(splitter * adxt_bdyt0)
			abig = (Float)(c - adxt_bdyt0)
			ahi = c - abig
			alo = adxt_bdyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(adxt_bdyt1 * cdz)
			c = (Float)(splitter * adxt_bdyt1)
			abig = (Float)(c - adxt_bdyt1)
			ahi = c - abig
			alo = adxt_bdyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if cdztail != 0.0 {
				c = (Float)(splitter * cdztail)
				abig = (Float)(c - cdztail)
				bhi = c - abig
				blo = cdztail - bhi
				_i = (Float)(adxt_bdyt0 * cdztail)
				c = (Float)(splitter * adxt_bdyt0)
				abig = (Float)(c - adxt_bdyt0)
				ahi = c - abig
				alo = adxt_bdyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(adxt_bdyt1 * cdztail)
				c = (Float)(splitter * adxt_bdyt1)
				abig = (Float)(c - adxt_bdyt1)
				ahi = c - abig
				alo = adxt_bdyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
		if cdytail != 0.0 {
			negate = -adxtail
			adxt_cdyt1 = (Float)(negate * cdytail)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			bhi = c - abig
			blo = cdytail - bhi
			err1 = adxt_cdyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			adxt_cdyt0 = (alo * blo) - err3
			c = (Float)(splitter * bdz)
			abig = (Float)(c - bdz)
			bhi = c - abig
			blo = bdz - bhi
			_i = (Float)(adxt_cdyt0 * bdz)
			c = (Float)(splitter * adxt_cdyt0)
			abig = (Float)(c - adxt_cdyt0)
			ahi = c - abig
			alo = adxt_cdyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(adxt_cdyt1 * bdz)
			c = (Float)(splitter * adxt_cdyt1)
			abig = (Float)(c - adxt_cdyt1)
			ahi = c - abig
			alo = adxt_cdyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if bdztail != 0.0 {
				c = (Float)(splitter * bdztail)
				abig = (Float)(c - bdztail)
				bhi = c - abig
				blo = bdztail - bhi
				_i = (Float)(adxt_cdyt0 * bdztail)
				c = (Float)(splitter * adxt_cdyt0)
				abig = (Float)(c - adxt_cdyt0)
				ahi = c - abig
				alo = adxt_cdyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(adxt_cdyt1 * bdztail)
				c = (Float)(splitter * adxt_cdyt1)
				abig = (Float)(c - adxt_cdyt1)
				ahi = c - abig
				alo = adxt_cdyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
	}
	if bdxtail != 0.0 {
		if cdytail != 0.0 {
			bdxt_cdyt1 = (Float)(bdxtail * cdytail)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			bhi = c - abig
			blo = cdytail - bhi
			err1 = bdxt_cdyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdxt_cdyt0 = (alo * blo) - err3
			c = (Float)(splitter * adz)
			abig = (Float)(c - adz)
			bhi = c - abig
			blo = adz - bhi
			_i = (Float)(bdxt_cdyt0 * adz)
			c = (Float)(splitter * bdxt_cdyt0)
			abig = (Float)(c - bdxt_cdyt0)
			ahi = c - abig
			alo = bdxt_cdyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(bdxt_cdyt1 * adz)
			c = (Float)(splitter * bdxt_cdyt1)
			abig = (Float)(c - bdxt_cdyt1)
			ahi = c - abig
			alo = bdxt_cdyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if adztail != 0.0 {
				c = (Float)(splitter * adztail)
				abig = (Float)(c - adztail)
				bhi = c - abig
				blo = adztail - bhi
				_i = (Float)(bdxt_cdyt0 * adztail)
				c = (Float)(splitter * bdxt_cdyt0)
				abig = (Float)(c - bdxt_cdyt0)
				ahi = c - abig
				alo = bdxt_cdyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(bdxt_cdyt1 * adztail)
				c = (Float)(splitter * bdxt_cdyt1)
				abig = (Float)(c - bdxt_cdyt1)
				ahi = c - abig
				alo = bdxt_cdyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
		if adytail != 0.0 {
			negate = -bdxtail
			bdxt_adyt1 = (Float)(negate * adytail)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			bhi = c - abig
			blo = adytail - bhi
			err1 = bdxt_adyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			bdxt_adyt0 = (alo * blo) - err3
			c = (Float)(splitter * cdz)
			abig = (Float)(c - cdz)
			bhi = c - abig
			blo = cdz - bhi
			_i = (Float)(bdxt_adyt0 * cdz)
			c = (Float)(splitter * bdxt_adyt0)
			abig = (Float)(c - bdxt_adyt0)
			ahi = c - abig
			alo = bdxt_adyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(bdxt_adyt1 * cdz)
			c = (Float)(splitter * bdxt_adyt1)
			abig = (Float)(c - bdxt_adyt1)
			ahi = c - abig
			alo = bdxt_adyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if cdztail != 0.0 {
				c = (Float)(splitter * cdztail)
				abig = (Float)(c - cdztail)
				bhi = c - abig
				blo = cdztail - bhi
				_i = (Float)(bdxt_adyt0 * cdztail)
				c = (Float)(splitter * bdxt_adyt0)
				abig = (Float)(c - bdxt_adyt0)
				ahi = c - abig
				alo = bdxt_adyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(bdxt_adyt1 * cdztail)
				c = (Float)(splitter * bdxt_adyt1)
				abig = (Float)(c - bdxt_adyt1)
				ahi = c - abig
				alo = bdxt_adyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
	}
	if cdxtail != 0.0 {
		if adytail != 0.0 {
			cdxt_adyt1 = (Float)(cdxtail * adytail)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			bhi = c - abig
			blo = adytail - bhi
			err1 = cdxt_adyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdxt_adyt0 = (alo * blo) - err3
			c = (Float)(splitter * bdz)
			abig = (Float)(c - bdz)
			bhi = c - abig
			blo = bdz - bhi
			_i = (Float)(cdxt_adyt0 * bdz)
			c = (Float)(splitter * cdxt_adyt0)
			abig = (Float)(c - cdxt_adyt0)
			ahi = c - abig
			alo = cdxt_adyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(cdxt_adyt1 * bdz)
			c = (Float)(splitter * cdxt_adyt1)
			abig = (Float)(c - cdxt_adyt1)
			ahi = c - abig
			alo = cdxt_adyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if bdztail != 0.0 {
				c = (Float)(splitter * bdztail)
				abig = (Float)(c - bdztail)
				bhi = c - abig
				blo = bdztail - bhi
				_i = (Float)(cdxt_adyt0 * bdztail)
				c = (Float)(splitter * cdxt_adyt0)
				abig = (Float)(c - cdxt_adyt0)
				ahi = c - abig
				alo = cdxt_adyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(cdxt_adyt1 * bdztail)
				c = (Float)(splitter * cdxt_adyt1)
				abig = (Float)(c - cdxt_adyt1)
				ahi = c - abig
				alo = cdxt_adyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
		if bdytail != 0.0 {
			negate = -cdxtail
			cdxt_bdyt1 = (Float)(negate * bdytail)
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			ahi = c - abig
			alo = negate - ahi
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			bhi = c - abig
			blo = bdytail - bhi
			err1 = cdxt_bdyt1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			cdxt_bdyt0 = (alo * blo) - err3
			c = (Float)(splitter * adz)
			abig = (Float)(c - adz)
			bhi = c - abig
			blo = adz - bhi
			_i = (Float)(cdxt_bdyt0 * adz)
			c = (Float)(splitter * cdxt_bdyt0)
			abig = (Float)(c - cdxt_bdyt0)
			ahi = c - abig
			alo = cdxt_bdyt0 - ahi
			err1 = _i - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			u[0] = (alo * blo) - err3
			_j = (Float)(cdxt_bdyt1 * adz)
			c = (Float)(splitter * cdxt_bdyt1)
			abig = (Float)(c - cdxt_bdyt1)
			ahi = c - abig
			alo = cdxt_bdyt1 - ahi
			err1 = _j - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			_0 = (alo * blo) - err3
			_k = (Float)(_i + _0)
			bvirt = (Float)(_k - _i)
			avirt = _k - bvirt
			bround = _0 - bvirt
			around = _i - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _k)
			bvirt = u3 - _j
			u[2] = _k - bvirt
			u[3] = u3
			finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if adztail != 0.0 {
				c = (Float)(splitter * adztail)
				abig = (Float)(c - adztail)
				bhi = c - abig
				blo = adztail - bhi
				_i = (Float)(cdxt_bdyt0 * adztail)
				c = (Float)(splitter * cdxt_bdyt0)
				abig = (Float)(c - cdxt_bdyt0)
				ahi = c - abig
				alo = cdxt_bdyt0 - ahi
				err1 = _i - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				u[0] = (alo * blo) - err3
				_j = (Float)(cdxt_bdyt1 * adztail)
				c = (Float)(splitter * cdxt_bdyt1)
				abig = (Float)(c - cdxt_bdyt1)
				ahi = c - abig
				alo = cdxt_bdyt1 - ahi
				err1 = _j - (ahi * bhi)
				err2 = err1 - (alo * bhi)
				err3 = err2 - (ahi * blo)
				_0 = (alo * blo) - err3
				_k = (Float)(_i + _0)
				bvirt = (Float)(_k - _i)
				avirt = _k - bvirt
				bround = _0 - bvirt
				around = _i - avirt
				u[1] = around + bround
				u3 = (Float)(_j + _k)
				bvirt = u3 - _j
				u[2] = _k - bvirt
				u[3] = u3
				finlength = FastExpansionSumZeroElim(finlength, finnow, 4, &u[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
		}
	}

	if adztail != 0.0 {
		wlength = ScaleExpansionZeroElim(bctlen, &bct[0], adztail, &w[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if bdztail != 0.0 {
		wlength = ScaleExpansionZeroElim(catlen, &cat[0], bdztail, &w[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if cdztail != 0.0 {
		wlength = ScaleExpansionZeroElim(abtlen, &abt[0], cdztail, &w[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, wlength, &w[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}

	return *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(finnow)) + floatSize*uintptr(finlength-1)))) // finnow[finlength-1]
}

/*****************************************************************************/
/*                                                                           */
/*  orient3dfast()   Approximate 3D orientation test.  Nonrobust.            */
/*  orient3dexact()   Exact 3D orientation test.  Robust.                    */
/*  orient3dslow()   Another exact 3D orientation test.  Robust.             */
/*  orient3d()   Adaptive exact 3D orientation test.  Robust.                */
/*                                                                           */
/*               Return a positive value if the point pd lies below the      */
/*               plane passing through pa, pb, and pc; "below" is defined so */
/*               that pa, pb, and pc appear in counterclockwise order when   */
/*               viewed from above the plane.  Returns a negative value if   */
/*               pd lies above the plane.  Returns zero if the points are    */
/*               coplanar.  The result is also a rough approximation of six  */
/*               times the signed volume of the tetrahedron defined by the   */
/*               four points.                                                */
/*                                                                           */
/*  Only the first and last routine should be used; the middle two are for   */
/*  timings.                                                                 */
/*                                                                           */
/*  The last three use exact arithmetic to ensure a correct answer.  The     */
/*  result returned is the determinant of a matrix.  In orient3d() only,     */
/*  this determinant is computed adaptively, in the sense that exact         */
/*  arithmetic is used only to the degree it is needed to ensure that the    */
/*  returned value has the correct sign.  Hence, orient3d() is usually quite */
/*  fast, but will run more slowly when the input points are coplanar or     */
/*  nearly so.                                                               */
/*                                                                           */
/*****************************************************************************/
func Orient3d(pa, pb, pc, pd [3]Float) Float {
	var adx, bdx, cdx, ady, bdy, cdy, adz, bdz, cdz Float
	var bdxcdy, cdxbdy, cdxady, adxcdy, adxbdy, bdxady Float
	var det Float
	var permanent, errbound Float

	adx = pa[0] - pd[0]
	bdx = pb[0] - pd[0]
	cdx = pc[0] - pd[0]
	ady = pa[1] - pd[1]
	bdy = pb[1] - pd[1]
	cdy = pc[1] - pd[1]
	adz = pa[2] - pd[2]
	bdz = pb[2] - pd[2]
	cdz = pc[2] - pd[2]

	bdxcdy = bdx * cdy
	cdxbdy = cdx * bdy

	cdxady = cdx * ady
	adxcdy = adx * cdy

	adxbdy = adx * bdy
	bdxady = bdx * ady

	det = adz*(bdxcdy-cdxbdy) +
		bdz*(cdxady-adxcdy) +
		cdz*(adxbdy-bdxady)

	permanent = (abs(bdxcdy)+abs(cdxbdy))*abs(adz) +
		(abs(cdxady)+abs(adxcdy))*abs(bdz) +
		(abs(adxbdy)+abs(bdxady))*abs(cdz)

	errbound = o3derrboundA * permanent
	if (det > errbound) || (-det > errbound) {
		return det
	}

	return Orient3dAdapt(pa, pb, pc, pd, permanent)
}

// # 2344 "./predicates.c.txt"
func IncircleFast(pa [2]Float, pb [2]Float, pc [2]Float, pd [2]Float) Float {
	var adx, ady, bdx, bdy, cdx, cdy Float
	var abdet, bcdet, cadet Float
	var alift, blift, clift Float

	adx = pa[0] - pd[0]
	ady = pa[1] - pd[1]
	bdx = pb[0] - pd[0]
	bdy = pb[1] - pd[1]
	cdx = pc[0] - pd[0]
	cdy = pc[1] - pd[1]

	abdet = adx*bdy - bdx*ady
	bcdet = bdx*cdy - cdx*bdy
	cadet = cdx*ady - adx*cdy
	alift = adx*adx + ady*ady
	blift = bdx*bdx + bdy*bdy
	clift = cdx*cdx + cdy*cdy

	return alift*bcdet + blift*cadet + clift*abdet
}

func IncircleExact(pa [2]Float, pb [2]Float, pc [2]Float, pd [2]Float) Float {
	var axby1, bxcy1, cxdy1, dxay1, axcy1, bxdy1 Float
	var bxay1, cxby1, dxcy1, axdy1, cxay1, dxby1 Float
	var axby0, bxcy0, cxdy0, dxay0, axcy0, bxdy0 Float
	var bxay0, cxby0, dxcy0, axdy0, cxay0, dxby0 Float
	var ab [4]Float
	var bc [4]Float
	var cd [4]Float
	var da [4]Float
	var ac [4]Float
	var bd [4]Float
	var temp8 [8]Float
	var templen int
	var abc [12]Float
	var bcd [12]Float
	var cda [12]Float
	var dab [12]Float
	var abclen, bcdlen, cdalen, dablen int
	var det24x [24]Float
	var det24y [24]Float
	var det48x [48]Float
	var det48y [48]Float
	var xlen, ylen int
	var adet [96]Float
	var bdet [96]Float
	var cdet [96]Float
	var ddet [96]Float
	var alen, blen, clen, dlen int
	var abdet [192]Float
	var cddet [192]Float
	var ablen, cdlen int
	var deter [384]Float
	var deterlen int
	var i int

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	axby1 = (Float)(pa[0] * pb[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = axby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axby0 = (alo * blo) - err3
	bxay1 = (Float)(pb[0] * pa[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = bxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxay0 = (alo * blo) - err3
	_i = (Float)(axby0 - bxay0)
	bvirt = (Float)(axby0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay0
	around = axby0 - avirt
	ab[0] = around + bround
	_j = (Float)(axby1 + _i)
	bvirt = (Float)(_j - axby1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axby1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay1
	around = _0 - avirt
	ab[1] = around + bround
	ab[3] = (Float)(_j + _i)
	bvirt = (Float)(ab[3] - _j)
	avirt = ab[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround

	bxcy1 = (Float)(pb[0] * pc[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = bxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxcy0 = (alo * blo) - err3
	cxby1 = (Float)(pc[0] * pb[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = cxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxby0 = (alo * blo) - err3
	_i = (Float)(bxcy0 - cxby0)
	bvirt = (Float)(bxcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby0
	around = bxcy0 - avirt
	bc[0] = around + bround
	_j = (Float)(bxcy1 + _i)
	bvirt = (Float)(_j - bxcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby1
	around = _0 - avirt
	bc[1] = around + bround
	bc[3] = (Float)(_j + _i)
	bvirt = (Float)(bc[3] - _j)
	avirt = bc[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround

	cxdy1 = (Float)(pc[0] * pd[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = cxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxdy0 = (alo * blo) - err3
	dxcy1 = (Float)(pd[0] * pc[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = dxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxcy0 = (alo * blo) - err3
	_i = (Float)(cxdy0 - dxcy0)
	bvirt = (Float)(cxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy0
	around = cxdy0 - avirt
	cd[0] = around + bround
	_j = (Float)(cxdy1 + _i)
	bvirt = (Float)(_j - cxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxcy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy1
	around = _0 - avirt
	cd[1] = around + bround
	cd[3] = (Float)(_j + _i)
	bvirt = (Float)(cd[3] - _j)
	avirt = cd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	cd[2] = around + bround

	dxay1 = (Float)(pd[0] * pa[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = dxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxay0 = (alo * blo) - err3
	axdy1 = (Float)(pa[0] * pd[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = axdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axdy0 = (alo * blo) - err3
	_i = (Float)(dxay0 - axdy0)
	bvirt = (Float)(dxay0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy0
	around = dxay0 - avirt
	da[0] = around + bround
	_j = (Float)(dxay1 + _i)
	bvirt = (Float)(_j - dxay1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = dxay1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - axdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy1
	around = _0 - avirt
	da[1] = around + bround
	da[3] = (Float)(_j + _i)
	bvirt = (Float)(da[3] - _j)
	avirt = da[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	da[2] = around + bround

	axcy1 = (Float)(pa[0] * pc[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = axcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axcy0 = (alo * blo) - err3
	cxay1 = (Float)(pc[0] * pa[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = cxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxay0 = (alo * blo) - err3
	_i = (Float)(axcy0 - cxay0)
	bvirt = (Float)(axcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay0
	around = axcy0 - avirt
	ac[0] = around + bround
	_j = (Float)(axcy1 + _i)
	bvirt = (Float)(_j - axcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay1
	around = _0 - avirt
	ac[1] = around + bround
	ac[3] = (Float)(_j + _i)
	bvirt = (Float)(ac[3] - _j)
	avirt = ac[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ac[2] = around + bround

	bxdy1 = (Float)(pb[0] * pd[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = bxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxdy0 = (alo * blo) - err3
	dxby1 = (Float)(pd[0] * pb[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = dxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxby0 = (alo * blo) - err3
	_i = (Float)(bxdy0 - dxby0)
	bvirt = (Float)(bxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby0
	around = bxdy0 - avirt
	bd[0] = around + bround
	_j = (Float)(bxdy1 + _i)
	bvirt = (Float)(_j - bxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby1
	around = _0 - avirt
	bd[1] = around + bround
	bd[3] = (Float)(_j + _i)
	bvirt = (Float)(bd[3] - _j)
	avirt = bd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bd[2] = around + bround

	templen = FastExpansionSumZeroElim(4, &cd[0], 4, &da[0], &temp8[0])
	cdalen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &ac[0], &cda[0])
	templen = FastExpansionSumZeroElim(4, &da[0], 4, &ab[0], &temp8[0])
	dablen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &bd[0], &dab[0])
	for i = 0; i < 4; i++ {
		bd[i] = -bd[i]
		ac[i] = -ac[i]
	}
	templen = FastExpansionSumZeroElim(4, &ab[0], 4, &bc[0], &temp8[0])
	abclen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &ac[0], &abc[0])
	templen = FastExpansionSumZeroElim(4, &bc[0], 4, &cd[0], &temp8[0])
	bcdlen = FastExpansionSumZeroElim(templen, &temp8[0], 4, &bd[0], &bcd[0])

	xlen = ScaleExpansionZeroElim(bcdlen, &bcd[0], pa[0], &det24x[0])
	xlen = ScaleExpansionZeroElim(xlen, &det24x[0], pa[0], &det48x[0])
	ylen = ScaleExpansionZeroElim(bcdlen, &bcd[0], pa[1], &det24y[0])
	ylen = ScaleExpansionZeroElim(ylen, &det24y[0], pa[1], &det48y[0])
	alen = FastExpansionSumZeroElim(xlen, &det48x[0], ylen, &det48y[0], &adet[0])

	xlen = ScaleExpansionZeroElim(cdalen, &cda[0], pb[0], &det24x[0])
	xlen = ScaleExpansionZeroElim(xlen, &det24x[0], -pb[0], &det48x[0])
	ylen = ScaleExpansionZeroElim(cdalen, &cda[0], pb[1], &det24y[0])
	ylen = ScaleExpansionZeroElim(ylen, &det24y[0], -pb[1], &det48y[0])
	blen = FastExpansionSumZeroElim(xlen, &det48x[0], ylen, &det48y[0], &bdet[0])

	xlen = ScaleExpansionZeroElim(dablen, &dab[0], pc[0], &det24x[0])
	xlen = ScaleExpansionZeroElim(xlen, &det24x[0], pc[0], &det48x[0])
	ylen = ScaleExpansionZeroElim(dablen, &dab[0], pc[1], &det24y[0])
	ylen = ScaleExpansionZeroElim(ylen, &det24y[0], pc[1], &det48y[0])
	clen = FastExpansionSumZeroElim(xlen, &det48x[0], ylen, &det48y[0], &cdet[0])

	xlen = ScaleExpansionZeroElim(abclen, &abc[0], pd[0], &det24x[0])
	xlen = ScaleExpansionZeroElim(xlen, &det24x[0], -pd[0], &det48x[0])
	ylen = ScaleExpansionZeroElim(abclen, &abc[0], pd[1], &det24y[0])
	ylen = ScaleExpansionZeroElim(ylen, &det24y[0], -pd[1], &det48y[0])
	dlen = FastExpansionSumZeroElim(xlen, &det48x[0], ylen, &det48y[0], &ddet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	cdlen = FastExpansionSumZeroElim(clen, &cdet[0], dlen, &ddet[0], &cddet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], cdlen, &cddet[0], &deter[0])

	return deter[deterlen-1]
}

func IncircleSlow(pa, pb, pc, pd [2]Float) Float {
	var adx, bdx, cdx, ady, bdy, cdy Float
	var adxtail, bdxtail, cdxtail Float
	var adytail, bdytail, cdytail Float
	var negate, negatetail Float
	var axby7, bxcy7, axcy7, bxay7, cxby7, cxay7 Float
	var axby, bxcy, axcy, bxay, cxby, cxay [8]Float
	var temp16 [16]Float
	var temp16len int
	var detx [32]Float
	var detxx [64]Float
	var detxt [32]Float
	var detxxt [64]Float
	var detxtxt [64]Float
	var xlen, xxlen, xtlen, xxtlen, xtxtlen int
	var x1 [128]Float
	var x2 [192]Float
	var x1len, x2len int
	var dety [32]Float
	var detyy [64]Float
	var detyt [32]Float
	var detyyt [64]Float
	var detytyt [64]Float
	var ylen, yylen, ytlen, yytlen, ytytlen int
	var y1 [128]Float
	var y2 [192]Float
	var y1len, y2len int
	var adet [384]Float
	var bdet [384]Float
	var cdet [384]Float
	var abdet [768]Float
	var deter [1152]Float
	var alen, blen, clen, ablen, deterlen int
	var i int
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var a0hi, a0lo, a1hi, a1lo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j, _k, _l, _m, _n Float
	var _0, _1, _2 Float

	adx = (Float)(pa[0] - pd[0])
	bvirt = (Float)(pa[0] - adx)
	avirt = adx + bvirt
	bround = bvirt - pd[0]
	around = pa[0] - avirt
	adxtail = around + bround
	ady = (Float)(pa[1] - pd[1])
	bvirt = (Float)(pa[1] - ady)
	avirt = ady + bvirt
	bround = bvirt - pd[1]
	around = pa[1] - avirt
	adytail = around + bround
	bdx = (Float)(pb[0] - pd[0])
	bvirt = (Float)(pb[0] - bdx)
	avirt = bdx + bvirt
	bround = bvirt - pd[0]
	around = pb[0] - avirt
	bdxtail = around + bround
	bdy = (Float)(pb[1] - pd[1])
	bvirt = (Float)(pb[1] - bdy)
	avirt = bdy + bvirt
	bround = bvirt - pd[1]
	around = pb[1] - avirt
	bdytail = around + bround
	cdx = (Float)(pc[0] - pd[0])
	bvirt = (Float)(pc[0] - cdx)
	avirt = cdx + bvirt
	bround = bvirt - pd[0]
	around = pc[0] - avirt
	cdxtail = around + bround
	cdy = (Float)(pc[1] - pd[1])
	bvirt = (Float)(pc[1] - cdy)
	avirt = cdy + bvirt
	bround = bvirt - pd[1]
	around = pc[1] - avirt
	cdytail = around + bround

	c = (Float)(splitter * adxtail)
	abig = (Float)(c - adxtail)
	a0hi = c - abig
	a0lo = adxtail - a0hi
	c = (Float)(splitter * bdytail)
	abig = (Float)(c - bdytail)
	bhi = c - abig
	blo = bdytail - bhi
	_i = (Float)(adxtail * bdytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	a1hi = c - abig
	a1lo = adx - a1hi
	_j = (Float)(adx * bdytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	_i = (Float)(adxtail * bdy)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(adx * bdy)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axby[5] = around + bround
	axby7 = (Float)(_m + _k)
	bvirt = (Float)(axby7 - _m)
	avirt = axby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axby[6] = around + bround

	axby[7] = axby7
	negate = -ady
	negatetail = -adytail
	c = (Float)(splitter * bdxtail)
	abig = (Float)(c - bdxtail)
	a0hi = c - abig
	a0lo = bdxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(bdxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	a1hi = c - abig
	a1lo = bdx - a1hi
	_j = (Float)(bdx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(bdxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bdx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxay[5] = around + bround
	bxay7 = (Float)(_m + _k)
	bvirt = (Float)(bxay7 - _m)
	avirt = bxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxay[6] = around + bround

	bxay[7] = bxay7
	c = (Float)(splitter * bdxtail)
	abig = (Float)(c - bdxtail)
	a0hi = c - abig
	a0lo = bdxtail - a0hi
	c = (Float)(splitter * cdytail)
	abig = (Float)(c - cdytail)
	bhi = c - abig
	blo = cdytail - bhi
	_i = (Float)(bdxtail * cdytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	a1hi = c - abig
	a1lo = bdx - a1hi
	_j = (Float)(bdx * cdytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	_i = (Float)(bdxtail * cdy)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bdx * cdy)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxcy[5] = around + bround
	bxcy7 = (Float)(_m + _k)
	bvirt = (Float)(bxcy7 - _m)
	avirt = bxcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxcy[6] = around + bround

	bxcy[7] = bxcy7
	negate = -bdy
	negatetail = -bdytail
	c = (Float)(splitter * cdxtail)
	abig = (Float)(c - cdxtail)
	a0hi = c - abig
	a0lo = cdxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(cdxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	a1hi = c - abig
	a1lo = cdx - a1hi
	_j = (Float)(cdx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(cdxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cdx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxby[5] = around + bround
	cxby7 = (Float)(_m + _k)
	bvirt = (Float)(cxby7 - _m)
	avirt = cxby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxby[6] = around + bround

	cxby[7] = cxby7
	c = (Float)(splitter * cdxtail)
	abig = (Float)(c - cdxtail)
	a0hi = c - abig
	a0lo = cdxtail - a0hi
	c = (Float)(splitter * adytail)
	abig = (Float)(c - adytail)
	bhi = c - abig
	blo = adytail - bhi
	_i = (Float)(cdxtail * adytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	a1hi = c - abig
	a1lo = cdx - a1hi
	_j = (Float)(cdx * adytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	_i = (Float)(cdxtail * ady)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cdx * ady)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxay[5] = around + bround
	cxay7 = (Float)(_m + _k)
	bvirt = (Float)(cxay7 - _m)
	avirt = cxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxay[6] = around + bround

	cxay[7] = cxay7
	negate = -cdy
	negatetail = -cdytail
	c = (Float)(splitter * adxtail)
	abig = (Float)(c - adxtail)
	a0hi = c - abig
	a0lo = adxtail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(adxtail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	a1hi = c - abig
	a1lo = adx - a1hi
	_j = (Float)(adx * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(adxtail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(adx * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axcy[5] = around + bround
	axcy7 = (Float)(_m + _k)
	bvirt = (Float)(axcy7 - _m)
	avirt = axcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axcy[6] = around + bround

	axcy[7] = axcy7

	temp16len = FastExpansionSumZeroElim(8, &bxcy[0], 8, &cxby[0], &temp16[0])

	xlen = ScaleExpansionZeroElim(temp16len, &temp16[0], adx, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], adx, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp16len, &temp16[0], adxtail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], adx, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], adxtail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])

	ylen = ScaleExpansionZeroElim(temp16len, &temp16[0], ady, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], ady, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp16len, &temp16[0], adytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], ady, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], adytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])

	alen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &adet[0])

	temp16len = FastExpansionSumZeroElim(8, &cxay[0], 8, &axcy[0], &temp16[0])

	xlen = ScaleExpansionZeroElim(temp16len, &temp16[0], bdx, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], bdx, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp16len, &temp16[0], bdxtail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], bdx, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], bdxtail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])

	ylen = ScaleExpansionZeroElim(temp16len, &temp16[0], bdy, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], bdy, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp16len, &temp16[0], bdytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], bdy, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], bdytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])

	blen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &bdet[0])

	temp16len = FastExpansionSumZeroElim(8, &axby[0], 8, &bxay[0], &temp16[0])

	xlen = ScaleExpansionZeroElim(temp16len, &temp16[0], cdx, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], cdx, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp16len, &temp16[0], cdxtail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], cdx, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], cdxtail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])

	ylen = ScaleExpansionZeroElim(temp16len, &temp16[0], cdy, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], cdy, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp16len, &temp16[0], cdytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], cdy, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], cdytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])

	clen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &cdet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], clen, &cdet[0], &deter[0])

	return deter[deterlen-1]
}

// # 2622 "./predicates.c.txt"
func IncircleAdapt(pa [2]Float, pb [2]Float, pc [2]Float, pd [2]Float, permanent Float) Float {
	var adx, bdx, cdx, ady, bdy, cdy Float
	var det, errbound Float

	var bdxcdy1, cdxbdy1, cdxady1, adxcdy1, adxbdy1, bdxady1 Float
	var bdxcdy0, cdxbdy0, cdxady0, adxcdy0, adxbdy0, bdxady0 Float
	var bc [4]Float
	var ca [4]Float
	var ab [4]Float
	var bc3, ca3, ab3 Float
	var axbc [8]Float
	var axxbc [16]Float
	var aybc [8]Float
	var ayybc [16]Float
	var adet [32]Float
	var axbclen, axxbclen, aybclen, ayybclen, alen int
	var bxca [8]Float
	var bxxca [16]Float
	var byca [8]Float
	var byyca [16]Float
	var bdet [32]Float
	var bxcalen, bxxcalen, bycalen, byycalen, blen int
	var cxab [8]Float
	var cxxab [16]Float
	var cyab [8]Float
	var cyyab [16]Float
	var cdet [32]Float
	var cxablen, cxxablen, cyablen, cyyablen, clen int
	var abdet [64]Float
	var ablen int
	var fin1 [1152]Float
	var fin2 [1152]Float
	var finnow, finother, finswap *Float
	var finlength int

	var adxtail, bdxtail, cdxtail, adytail, bdytail, cdytail Float
	var adxadx1, adyady1, bdxbdx1, bdybdy1, cdxcdx1, cdycdy1 Float
	var adxadx0, adyady0, bdxbdx0, bdybdy0, cdxcdx0, cdycdy0 Float
	var aa [4]Float
	var bb [4]Float
	var cc [4]Float
	var aa3, bb3, cc3 Float
	var ti1, tj1 Float
	var ti0, tj0 Float
	var u [4]Float
	var v [4]Float
	var u3, v3 Float
	var temp8 [8]Float
	var temp16a [16]Float
	var temp16b [16]Float
	var temp16c [16]Float
	var temp32a [32]Float
	var temp32b [32]Float
	var temp48 [48]Float
	var temp64 [64]Float
	var temp8len, temp16alen, temp16blen, temp16clen int
	var temp32alen, temp32blen, temp48len, temp64len int
	var axtbb [8]Float
	var axtcc [8]Float
	var aytbb [8]Float
	var aytcc [8]Float
	var axtbblen, axtcclen, aytbblen, aytcclen int
	var bxtaa [8]Float
	var bxtcc [8]Float
	var bytaa [8]Float
	var bytcc [8]Float
	var bxtaalen, bxtcclen, bytaalen, bytcclen int
	var cxtaa [8]Float
	var cxtbb [8]Float
	var cytaa [8]Float
	var cytbb [8]Float
	var cxtaalen, cxtbblen, cytaalen, cytbblen int
	var axtbc [8]Float
	var aytbc [8]Float
	var bxtca [8]Float
	var bytca [8]Float
	var cxtab [8]Float
	var cytab [8]Float

	var axtbclen, aytbclen, bxtcalen, bytcalen, cxtablen, cytablen int
	var axtbct [16]Float
	var aytbct [16]Float
	var bxtcat [16]Float
	var bytcat [16]Float
	var cxtabt [16]Float
	var cytabt [16]Float
	var axtbctlen, aytbctlen, bxtcatlen, bytcatlen, cxtabtlen, cytabtlen int
	var axtbctt [8]Float
	var aytbctt [8]Float
	var bxtcatt [8]Float
	var bytcatt [8]Float
	var cxtabtt [8]Float
	var cytabtt [8]Float
	var axtbcttlen, aytbcttlen, bxtcattlen, bytcattlen, cxtabttlen, cytabttlen int
	var abt [8]Float
	var bct [8]Float
	var cat [8]Float
	var abtlen, bctlen, catlen int
	var abtt [4]Float
	var bctt [4]Float
	var catt [4]Float
	var abttlen, bcttlen, cattlen int
	var abtt3, bctt3, catt3 Float
	var negate Float

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	adx = (Float)(pa[0] - pd[0])
	bdx = (Float)(pb[0] - pd[0])
	cdx = (Float)(pc[0] - pd[0])
	ady = (Float)(pa[1] - pd[1])
	bdy = (Float)(pb[1] - pd[1])
	cdy = (Float)(pc[1] - pd[1])

	bdxcdy1 = (Float)(bdx * cdy)
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	ahi = c - abig
	alo = bdx - ahi
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	err1 = bdxcdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bdxcdy0 = (alo * blo) - err3
	cdxbdy1 = (Float)(cdx * bdy)
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	ahi = c - abig
	alo = cdx - ahi
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	err1 = cdxbdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cdxbdy0 = (alo * blo) - err3
	_i = (Float)(bdxcdy0 - cdxbdy0)
	bvirt = (Float)(bdxcdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cdxbdy0
	around = bdxcdy0 - avirt
	bc[0] = around + bround
	_j = (Float)(bdxcdy1 + _i)
	bvirt = (Float)(_j - bdxcdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bdxcdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cdxbdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cdxbdy1
	around = _0 - avirt
	bc[1] = around + bround
	bc3 = (Float)(_j + _i)
	bvirt = (Float)(bc3 - _j)
	avirt = bc3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround
	bc[3] = bc3
	axbclen = ScaleExpansionZeroElim(4, &bc[0], adx, &axbc[0])
	axxbclen = ScaleExpansionZeroElim(axbclen, &axbc[0], adx, &axxbc[0])
	aybclen = ScaleExpansionZeroElim(4, &bc[0], ady, &aybc[0])
	ayybclen = ScaleExpansionZeroElim(aybclen, &aybc[0], ady, &ayybc[0])
	alen = FastExpansionSumZeroElim(axxbclen, &axxbc[0], ayybclen, &ayybc[0], &adet[0])

	cdxady1 = (Float)(cdx * ady)
	c = (Float)(splitter * cdx)
	abig = (Float)(c - cdx)
	ahi = c - abig
	alo = cdx - ahi
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	err1 = cdxady1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cdxady0 = (alo * blo) - err3
	adxcdy1 = (Float)(adx * cdy)
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	ahi = c - abig
	alo = adx - ahi
	c = (Float)(splitter * cdy)
	abig = (Float)(c - cdy)
	bhi = c - abig
	blo = cdy - bhi
	err1 = adxcdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	adxcdy0 = (alo * blo) - err3
	_i = (Float)(cdxady0 - adxcdy0)
	bvirt = (Float)(cdxady0 - _i)
	avirt = _i + bvirt
	bround = bvirt - adxcdy0
	around = cdxady0 - avirt
	ca[0] = around + bround
	_j = (Float)(cdxady1 + _i)
	bvirt = (Float)(_j - cdxady1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cdxady1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - adxcdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - adxcdy1
	around = _0 - avirt
	ca[1] = around + bround
	ca3 = (Float)(_j + _i)
	bvirt = (Float)(ca3 - _j)
	avirt = ca3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ca[2] = around + bround
	ca[3] = ca3
	bxcalen = ScaleExpansionZeroElim(4, &ca[0], bdx, &bxca[0])
	bxxcalen = ScaleExpansionZeroElim(bxcalen, &bxca[0], bdx, &bxxca[0])
	bycalen = ScaleExpansionZeroElim(4, &ca[0], bdy, &byca[0])
	byycalen = ScaleExpansionZeroElim(bycalen, &byca[0], bdy, &byyca[0])
	blen = FastExpansionSumZeroElim(bxxcalen, &bxxca[0], byycalen, &byyca[0], &bdet[0])

	adxbdy1 = (Float)(adx * bdy)
	c = (Float)(splitter * adx)
	abig = (Float)(c - adx)
	ahi = c - abig
	alo = adx - ahi
	c = (Float)(splitter * bdy)
	abig = (Float)(c - bdy)
	bhi = c - abig
	blo = bdy - bhi
	err1 = adxbdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	adxbdy0 = (alo * blo) - err3
	bdxady1 = (Float)(bdx * ady)
	c = (Float)(splitter * bdx)
	abig = (Float)(c - bdx)
	ahi = c - abig
	alo = bdx - ahi
	c = (Float)(splitter * ady)
	abig = (Float)(c - ady)
	bhi = c - abig
	blo = ady - bhi
	err1 = bdxady1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bdxady0 = (alo * blo) - err3
	_i = (Float)(adxbdy0 - bdxady0)
	bvirt = (Float)(adxbdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bdxady0
	around = adxbdy0 - avirt
	ab[0] = around + bround
	_j = (Float)(adxbdy1 + _i)
	bvirt = (Float)(_j - adxbdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = adxbdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bdxady1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bdxady1
	around = _0 - avirt
	ab[1] = around + bround
	ab3 = (Float)(_j + _i)
	bvirt = (Float)(ab3 - _j)
	avirt = ab3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround
	ab[3] = ab3
	cxablen = ScaleExpansionZeroElim(4, &ab[0], cdx, &cxab[0])
	cxxablen = ScaleExpansionZeroElim(cxablen, &cxab[0], cdx, &cxxab[0])
	cyablen = ScaleExpansionZeroElim(4, &ab[0], cdy, &cyab[0])
	cyyablen = ScaleExpansionZeroElim(cyablen, &cyab[0], cdy, &cyyab[0])
	clen = FastExpansionSumZeroElim(cxxablen, &cxxab[0], cyyablen, &cyyab[0], &cdet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	finlength = FastExpansionSumZeroElim(ablen, &abdet[0], clen, &cdet[0], &fin1[0])

	det = Estimate(finlength, &fin1[0])
	errbound = iccerrboundB * permanent
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	bvirt = (Float)(pa[0] - adx)
	avirt = adx + bvirt
	bround = bvirt - pd[0]
	around = pa[0] - avirt
	adxtail = around + bround
	bvirt = (Float)(pa[1] - ady)
	avirt = ady + bvirt
	bround = bvirt - pd[1]
	around = pa[1] - avirt
	adytail = around + bround
	bvirt = (Float)(pb[0] - bdx)
	avirt = bdx + bvirt
	bround = bvirt - pd[0]
	around = pb[0] - avirt
	bdxtail = around + bround
	bvirt = (Float)(pb[1] - bdy)
	avirt = bdy + bvirt
	bround = bvirt - pd[1]
	around = pb[1] - avirt
	bdytail = around + bround
	bvirt = (Float)(pc[0] - cdx)
	avirt = cdx + bvirt
	bround = bvirt - pd[0]
	around = pc[0] - avirt
	cdxtail = around + bround
	bvirt = (Float)(pc[1] - cdy)
	avirt = cdy + bvirt
	bround = bvirt - pd[1]
	around = pc[1] - avirt
	cdytail = around + bround
	if (adxtail == 0.0) && (bdxtail == 0.0) && (cdxtail == 0.0) &&
		(adytail == 0.0) && (bdytail == 0.0) && (cdytail == 0.0) {
		return det
	}

	errbound = iccerrboundC*permanent + resulterrbound*abs(det)
	det += ((adx*adx+ady*ady)*((bdx*cdytail+cdy*bdxtail)-
		(bdy*cdxtail+cdx*bdytail)) +
		2.0*(adx*adxtail+ady*adytail)*(bdx*cdy-bdy*cdx)) +
		((bdx*bdx+bdy*bdy)*((cdx*adytail+ady*cdxtail)-
			(cdy*adxtail+adx*cdytail)) +
			2.0*(bdx*bdxtail+bdy*bdytail)*(cdx*ady-cdy*adx)) +
		((cdx*cdx+cdy*cdy)*((adx*bdytail+bdy*adxtail)-
			(ady*bdxtail+bdx*adytail)) +
			2.0*(cdx*cdxtail+cdy*cdytail)*(adx*bdy-ady*bdx))
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	finnow = &fin1[0]
	finother = &fin2[0]

	if (bdxtail != 0.0) || (bdytail != 0.0) ||
		(cdxtail != 0.0) || (cdytail != 0.0) {
		adxadx1 = (Float)(adx * adx)
		c = (Float)(splitter * adx)
		abig = (Float)(c - adx)
		ahi = c - abig
		alo = adx - ahi
		err1 = adxadx1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		adxadx0 = (alo * alo) - err3
		adyady1 = (Float)(ady * ady)
		c = (Float)(splitter * ady)
		abig = (Float)(c - ady)
		ahi = c - abig
		alo = ady - ahi
		err1 = adyady1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		adyady0 = (alo * alo) - err3
		_i = (Float)(adxadx0 + adyady0)
		bvirt = (Float)(_i - adxadx0)
		avirt = _i - bvirt
		bround = adyady0 - bvirt
		around = adxadx0 - avirt
		aa[0] = around + bround
		_j = (Float)(adxadx1 + _i)
		bvirt = (Float)(_j - adxadx1)
		avirt = _j - bvirt
		bround = _i - bvirt
		around = adxadx1 - avirt
		_0 = around + bround
		_i = (Float)(_0 + adyady1)
		bvirt = (Float)(_i - _0)
		avirt = _i - bvirt
		bround = adyady1 - bvirt
		around = _0 - avirt
		aa[1] = around + bround
		aa3 = (Float)(_j + _i)
		bvirt = (Float)(aa3 - _j)
		avirt = aa3 - bvirt
		bround = _i - bvirt
		around = _j - avirt
		aa[2] = around + bround
		aa[3] = aa3
	}
	if (cdxtail != 0.0) || (cdytail != 0.0) ||
		(adxtail != 0.0) || (adytail != 0.0) {
		bdxbdx1 = (Float)(bdx * bdx)
		c = (Float)(splitter * bdx)
		abig = (Float)(c - bdx)
		ahi = c - abig
		alo = bdx - ahi
		err1 = bdxbdx1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		bdxbdx0 = (alo * alo) - err3
		bdybdy1 = (Float)(bdy * bdy)
		c = (Float)(splitter * bdy)
		abig = (Float)(c - bdy)
		ahi = c - abig
		alo = bdy - ahi
		err1 = bdybdy1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		bdybdy0 = (alo * alo) - err3
		_i = (Float)(bdxbdx0 + bdybdy0)
		bvirt = (Float)(_i - bdxbdx0)
		avirt = _i - bvirt
		bround = bdybdy0 - bvirt
		around = bdxbdx0 - avirt
		bb[0] = around + bround
		_j = (Float)(bdxbdx1 + _i)
		bvirt = (Float)(_j - bdxbdx1)
		avirt = _j - bvirt
		bround = _i - bvirt
		around = bdxbdx1 - avirt
		_0 = around + bround
		_i = (Float)(_0 + bdybdy1)
		bvirt = (Float)(_i - _0)
		avirt = _i - bvirt
		bround = bdybdy1 - bvirt
		around = _0 - avirt
		bb[1] = around + bround
		bb3 = (Float)(_j + _i)
		bvirt = (Float)(bb3 - _j)
		avirt = bb3 - bvirt
		bround = _i - bvirt
		around = _j - avirt
		bb[2] = around + bround
		bb[3] = bb3
	}
	if (adxtail != 0.0) || (adytail != 0.0) ||
		(bdxtail != 0.0) || (bdytail != 0.0) {
		cdxcdx1 = (Float)(cdx * cdx)
		c = (Float)(splitter * cdx)
		abig = (Float)(c - cdx)
		ahi = c - abig
		alo = cdx - ahi
		err1 = cdxcdx1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		cdxcdx0 = (alo * alo) - err3
		cdycdy1 = (Float)(cdy * cdy)
		c = (Float)(splitter * cdy)
		abig = (Float)(c - cdy)
		ahi = c - abig
		alo = cdy - ahi
		err1 = cdycdy1 - (ahi * ahi)
		err3 = err1 - ((ahi + ahi) * alo)
		cdycdy0 = (alo * alo) - err3
		_i = (Float)(cdxcdx0 + cdycdy0)
		bvirt = (Float)(_i - cdxcdx0)
		avirt = _i - bvirt
		bround = cdycdy0 - bvirt
		around = cdxcdx0 - avirt
		cc[0] = around + bround
		_j = (Float)(cdxcdx1 + _i)
		bvirt = (Float)(_j - cdxcdx1)
		avirt = _j - bvirt
		bround = _i - bvirt
		around = cdxcdx1 - avirt
		_0 = around + bround
		_i = (Float)(_0 + cdycdy1)
		bvirt = (Float)(_i - _0)
		avirt = _i - bvirt
		bround = cdycdy1 - bvirt
		around = _0 - avirt
		cc[1] = around + bround
		cc3 = (Float)(_j + _i)
		bvirt = (Float)(cc3 - _j)
		avirt = cc3 - bvirt
		bround = _i - bvirt
		around = _j - avirt
		cc[2] = around + bround
		cc[3] = cc3
	}

	if adxtail != 0.0 {
		axtbclen = ScaleExpansionZeroElim(4, &bc[0], adxtail, &axtbc[0])
		temp16alen = ScaleExpansionZeroElim(axtbclen, &axtbc[0], 2.0*adx, &temp16a[0])

		axtcclen = ScaleExpansionZeroElim(4, &cc[0], adxtail, &axtcc[0])
		temp16blen = ScaleExpansionZeroElim(axtcclen, &axtcc[0], bdy, &temp16b[0])

		axtbblen = ScaleExpansionZeroElim(4, &bb[0], adxtail, &axtbb[0])
		temp16clen = ScaleExpansionZeroElim(axtbblen, &axtbb[0], -cdy, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if adytail != 0.0 {
		aytbclen = ScaleExpansionZeroElim(4, &bc[0], adytail, &aytbc[0])
		temp16alen = ScaleExpansionZeroElim(aytbclen, &aytbc[0], 2.0*ady, &temp16a[0])

		aytbblen = ScaleExpansionZeroElim(4, &bb[0], adytail, &aytbb[0])
		temp16blen = ScaleExpansionZeroElim(aytbblen, &aytbb[0], cdx, &temp16b[0])

		aytcclen = ScaleExpansionZeroElim(4, &cc[0], adytail, &aytcc[0])
		temp16clen = ScaleExpansionZeroElim(aytcclen, &aytcc[0], -bdx, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if bdxtail != 0.0 {
		bxtcalen = ScaleExpansionZeroElim(4, &ca[0], bdxtail, &bxtca[0])
		temp16alen = ScaleExpansionZeroElim(bxtcalen, &bxtca[0], 2.0*bdx, &temp16a[0])

		bxtaalen = ScaleExpansionZeroElim(4, &aa[0], bdxtail, &bxtaa[0])
		temp16blen = ScaleExpansionZeroElim(bxtaalen, &bxtaa[0], cdy, &temp16b[0])

		bxtcclen = ScaleExpansionZeroElim(4, &cc[0], bdxtail, &bxtcc[0])
		temp16clen = ScaleExpansionZeroElim(bxtcclen, &bxtcc[0], -ady, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if bdytail != 0.0 {
		bytcalen = ScaleExpansionZeroElim(4, &ca[0], bdytail, &bytca[0])
		temp16alen = ScaleExpansionZeroElim(bytcalen, &bytca[0], 2.0*bdy, &temp16a[0])

		bytcclen = ScaleExpansionZeroElim(4, &cc[0], bdytail, &bytcc[0])
		temp16blen = ScaleExpansionZeroElim(bytcclen, &bytcc[0], adx, &temp16b[0])

		bytaalen = ScaleExpansionZeroElim(4, &aa[0], bdytail, &bytaa[0])
		temp16clen = ScaleExpansionZeroElim(bytaalen, &bytaa[0], -cdx, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if cdxtail != 0.0 {
		cxtablen = ScaleExpansionZeroElim(4, &ab[0], cdxtail, &cxtab[0])
		temp16alen = ScaleExpansionZeroElim(cxtablen, &cxtab[0], 2.0*cdx, &temp16a[0])

		cxtbblen = ScaleExpansionZeroElim(4, &bb[0], cdxtail, &cxtbb[0])
		temp16blen = ScaleExpansionZeroElim(cxtbblen, &cxtbb[0], ady, &temp16b[0])

		cxtaalen = ScaleExpansionZeroElim(4, &aa[0], cdxtail, &cxtaa[0])
		temp16clen = ScaleExpansionZeroElim(cxtaalen, &cxtaa[0], -bdy, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}
	if cdytail != 0.0 {
		cytablen = ScaleExpansionZeroElim(4, &ab[0], cdytail, &cytab[0])
		temp16alen = ScaleExpansionZeroElim(cytablen, &cytab[0], 2.0*cdy, &temp16a[0])

		cytaalen = ScaleExpansionZeroElim(4, &aa[0], cdytail, &cytaa[0])
		temp16blen = ScaleExpansionZeroElim(cytaalen, &cytaa[0], bdx, &temp16b[0])

		cytbblen = ScaleExpansionZeroElim(4, &bb[0], cdytail, &cytbb[0])
		temp16clen = ScaleExpansionZeroElim(cytbblen, &cytbb[0], -adx, &temp16c[0])

		temp32alen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32a[0])
		temp48len = FastExpansionSumZeroElim(temp16clen, &temp16c[0], temp32alen, &temp32a[0], &temp48[0])
		finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
		finswap = finnow
		finnow = finother
		finother = finswap
	}

	if (adxtail != 0.0) || (adytail != 0.0) {
		if (bdxtail != 0.0) || (bdytail != 0.0) ||
			(cdxtail != 0.0) || (cdytail != 0.0) {
			ti1 = (Float)(bdxtail * cdy)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * cdy)
			abig = (Float)(c - cdy)
			bhi = c - abig
			blo = cdy - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(bdx * cdytail)
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			ahi = c - abig
			alo = bdx - ahi
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			bhi = c - abig
			blo = cdytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			u[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _i)
			bvirt = (Float)(u3 - _j)
			avirt = u3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			u[2] = around + bround
			u[3] = u3
			negate = -bdy
			ti1 = (Float)(cdxtail * negate)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			negate = -bdytail
			tj1 = (Float)(cdx * negate)
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			ahi = c - abig
			alo = cdx - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			v[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			v[1] = around + bround
			v3 = (Float)(_j + _i)
			bvirt = (Float)(v3 - _j)
			avirt = v3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			v[2] = around + bround
			v[3] = v3
			bctlen = FastExpansionSumZeroElim(4, &u[0], 4, &v[0], &bct[0])

			ti1 = (Float)(bdxtail * cdytail)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			bhi = c - abig
			blo = cdytail - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(cdxtail * bdytail)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			bhi = c - abig
			blo = bdytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 - tj0)
			bvirt = (Float)(ti0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj0
			around = ti0 - avirt
			bctt[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - tj1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj1
			around = _0 - avirt
			bctt[1] = around + bround
			bctt3 = (Float)(_j + _i)
			bvirt = (Float)(bctt3 - _j)
			avirt = bctt3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			bctt[2] = around + bround
			bctt[3] = bctt3
			bcttlen = 4
		} else {
			bct[0] = 0.0
			bctlen = 1
			bctt[0] = 0.0
			bcttlen = 1
		}

		if adxtail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(axtbclen, &axtbc[0], adxtail, &temp16a[0])
			axtbctlen = ScaleExpansionZeroElim(bctlen, &bct[0], adxtail, &axtbct[0])
			temp32alen = ScaleExpansionZeroElim(axtbctlen, &axtbct[0], 2.0*adx, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if bdytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &cc[0], adxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], bdytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
			if cdytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &bb[0], -adxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], cdytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}

			temp32alen = ScaleExpansionZeroElim(axtbctlen, &axtbct[0], adxtail, &temp32a[0])
			axtbcttlen = ScaleExpansionZeroElim(bcttlen, &bctt[0], adxtail, &axtbctt[0])
			temp16alen = ScaleExpansionZeroElim(axtbcttlen, &axtbctt[0], 2.0*adx, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(axtbcttlen, &axtbctt[0], adxtail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
		if adytail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(aytbclen, &aytbc[0], adytail, &temp16a[0])
			aytbctlen = ScaleExpansionZeroElim(bctlen, &bct[0], adytail, &aytbct[0])
			temp32alen = ScaleExpansionZeroElim(aytbctlen, &aytbct[0], 2.0*ady, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap

			temp32alen = ScaleExpansionZeroElim(aytbctlen, &aytbct[0], adytail, &temp32a[0])
			aytbcttlen = ScaleExpansionZeroElim(bcttlen, &bctt[0], adytail, &aytbctt[0])
			temp16alen = ScaleExpansionZeroElim(aytbcttlen, &aytbctt[0], 2.0*ady, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(aytbcttlen, &aytbctt[0], adytail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
	}
	if (bdxtail != 0.0) || (bdytail != 0.0) {
		if (cdxtail != 0.0) || (cdytail != 0.0) ||
			(adxtail != 0.0) || (adytail != 0.0) {
			ti1 = (Float)(cdxtail * ady)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * ady)
			abig = (Float)(c - ady)
			bhi = c - abig
			blo = ady - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(cdx * adytail)
			c = (Float)(splitter * cdx)
			abig = (Float)(c - cdx)
			ahi = c - abig
			alo = cdx - ahi
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			bhi = c - abig
			blo = adytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			u[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _i)
			bvirt = (Float)(u3 - _j)
			avirt = u3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			u[2] = around + bround
			u[3] = u3
			negate = -cdy
			ti1 = (Float)(adxtail * negate)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			negate = -cdytail
			tj1 = (Float)(adx * negate)
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			ahi = c - abig
			alo = adx - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			v[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			v[1] = around + bround
			v3 = (Float)(_j + _i)
			bvirt = (Float)(v3 - _j)
			avirt = v3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			v[2] = around + bround
			v[3] = v3
			catlen = FastExpansionSumZeroElim(4, &u[0], 4, &v[0], &cat[0])

			ti1 = (Float)(cdxtail * adytail)
			c = (Float)(splitter * cdxtail)
			abig = (Float)(c - cdxtail)
			ahi = c - abig
			alo = cdxtail - ahi
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			bhi = c - abig
			blo = adytail - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(adxtail * cdytail)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * cdytail)
			abig = (Float)(c - cdytail)
			bhi = c - abig
			blo = cdytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 - tj0)
			bvirt = (Float)(ti0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj0
			around = ti0 - avirt
			catt[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - tj1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj1
			around = _0 - avirt
			catt[1] = around + bround
			catt3 = (Float)(_j + _i)
			bvirt = (Float)(catt3 - _j)
			avirt = catt3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			catt[2] = around + bround
			catt[3] = catt3
			cattlen = 4
		} else {
			cat[0] = 0.0
			catlen = 1
			catt[0] = 0.0
			cattlen = 1
		}

		if bdxtail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(bxtcalen, &bxtca[0], bdxtail, &temp16a[0])
			bxtcatlen = ScaleExpansionZeroElim(catlen, &cat[0], bdxtail, &bxtcat[0])
			temp32alen = ScaleExpansionZeroElim(bxtcatlen, &bxtcat[0], 2.0*bdx, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if cdytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &aa[0], bdxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], cdytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
			if adytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &cc[0], -bdxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], adytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}

			temp32alen = ScaleExpansionZeroElim(bxtcatlen, &bxtcat[0], bdxtail, &temp32a[0])
			bxtcattlen = ScaleExpansionZeroElim(cattlen, &catt[0], bdxtail, &bxtcatt[0])
			temp16alen = ScaleExpansionZeroElim(bxtcattlen, &bxtcatt[0], 2.0*bdx, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(bxtcattlen, &bxtcatt[0], bdxtail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
		if bdytail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(bytcalen, &bytca[0], bdytail, &temp16a[0])
			bytcatlen = ScaleExpansionZeroElim(catlen, &cat[0], bdytail, &bytcat[0])
			temp32alen = ScaleExpansionZeroElim(bytcatlen, &bytcat[0], 2.0*bdy, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap

			temp32alen = ScaleExpansionZeroElim(bytcatlen, &bytcat[0], bdytail, &temp32a[0])
			bytcattlen = ScaleExpansionZeroElim(cattlen, &catt[0], bdytail, &bytcatt[0])
			temp16alen = ScaleExpansionZeroElim(bytcattlen, &bytcatt[0], 2.0*bdy, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(bytcattlen, &bytcatt[0], bdytail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
	}
	if (cdxtail != 0.0) || (cdytail != 0.0) {
		if (adxtail != 0.0) || (adytail != 0.0) ||
			(bdxtail != 0.0) || (bdytail != 0.0) {
			ti1 = (Float)(adxtail * bdy)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * bdy)
			abig = (Float)(c - bdy)
			bhi = c - abig
			blo = bdy - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(adx * bdytail)
			c = (Float)(splitter * adx)
			abig = (Float)(c - adx)
			ahi = c - abig
			alo = adx - ahi
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			bhi = c - abig
			blo = bdytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			u[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			u[1] = around + bround
			u3 = (Float)(_j + _i)
			bvirt = (Float)(u3 - _j)
			avirt = u3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			u[2] = around + bround
			u[3] = u3
			negate = -ady
			ti1 = (Float)(bdxtail * negate)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			negate = -adytail
			tj1 = (Float)(bdx * negate)
			c = (Float)(splitter * bdx)
			abig = (Float)(c - bdx)
			ahi = c - abig
			alo = bdx - ahi
			c = (Float)(splitter * negate)
			abig = (Float)(c - negate)
			bhi = c - abig
			blo = negate - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 + tj0)
			bvirt = (Float)(_i - ti0)
			avirt = _i - bvirt
			bround = tj0 - bvirt
			around = ti0 - avirt
			v[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 + tj1)
			bvirt = (Float)(_i - _0)
			avirt = _i - bvirt
			bround = tj1 - bvirt
			around = _0 - avirt
			v[1] = around + bround
			v3 = (Float)(_j + _i)
			bvirt = (Float)(v3 - _j)
			avirt = v3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			v[2] = around + bround
			v[3] = v3
			abtlen = FastExpansionSumZeroElim(4, &u[0], 4, &v[0], &abt[0])

			ti1 = (Float)(adxtail * bdytail)
			c = (Float)(splitter * adxtail)
			abig = (Float)(c - adxtail)
			ahi = c - abig
			alo = adxtail - ahi
			c = (Float)(splitter * bdytail)
			abig = (Float)(c - bdytail)
			bhi = c - abig
			blo = bdytail - bhi
			err1 = ti1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			ti0 = (alo * blo) - err3
			tj1 = (Float)(bdxtail * adytail)
			c = (Float)(splitter * bdxtail)
			abig = (Float)(c - bdxtail)
			ahi = c - abig
			alo = bdxtail - ahi
			c = (Float)(splitter * adytail)
			abig = (Float)(c - adytail)
			bhi = c - abig
			blo = adytail - bhi
			err1 = tj1 - (ahi * bhi)
			err2 = err1 - (alo * bhi)
			err3 = err2 - (ahi * blo)
			tj0 = (alo * blo) - err3
			_i = (Float)(ti0 - tj0)
			bvirt = (Float)(ti0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj0
			around = ti0 - avirt
			abtt[0] = around + bround
			_j = (Float)(ti1 + _i)
			bvirt = (Float)(_j - ti1)
			avirt = _j - bvirt
			bround = _i - bvirt
			around = ti1 - avirt
			_0 = around + bround
			_i = (Float)(_0 - tj1)
			bvirt = (Float)(_0 - _i)
			avirt = _i + bvirt
			bround = bvirt - tj1
			around = _0 - avirt
			abtt[1] = around + bround
			abtt3 = (Float)(_j + _i)
			bvirt = (Float)(abtt3 - _j)
			avirt = abtt3 - bvirt
			bround = _i - bvirt
			around = _j - avirt
			abtt[2] = around + bround
			abtt[3] = abtt3
			abttlen = 4
		} else {
			abt[0] = 0.0
			abtlen = 1
			abtt[0] = 0.0
			abttlen = 1
		}

		if cdxtail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(cxtablen, &cxtab[0], cdxtail, &temp16a[0])
			cxtabtlen = ScaleExpansionZeroElim(abtlen, &abt[0], cdxtail, &cxtabt[0])
			temp32alen = ScaleExpansionZeroElim(cxtabtlen, &cxtabt[0], 2.0*cdx, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
			if adytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &bb[0], cdxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], adytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}
			if bdytail != 0.0 {
				temp8len = ScaleExpansionZeroElim(4, &aa[0], -cdxtail, &temp8[0])
				temp16alen = ScaleExpansionZeroElim(temp8len, &temp8[0], bdytail, &temp16a[0])
				finlength = FastExpansionSumZeroElim(finlength, finnow, temp16alen, &temp16a[0], finother)
				finswap = finnow
				finnow = finother
				finother = finswap
			}

			temp32alen = ScaleExpansionZeroElim(cxtabtlen, &cxtabt[0], cdxtail, &temp32a[0])
			cxtabttlen = ScaleExpansionZeroElim(abttlen, &abtt[0], cdxtail, &cxtabtt[0])
			temp16alen = ScaleExpansionZeroElim(cxtabttlen, &cxtabtt[0], 2.0*cdx, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(cxtabttlen, &cxtabtt[0], cdxtail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
		if cdytail != 0.0 {
			temp16alen = ScaleExpansionZeroElim(cytablen, &cytab[0], cdytail, &temp16a[0])
			cytabtlen = ScaleExpansionZeroElim(abtlen, &abt[0], cdytail, &cytabt[0])
			temp32alen = ScaleExpansionZeroElim(cytabtlen, &cytabt[0], 2.0*cdy, &temp32a[0])
			temp48len = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp32alen, &temp32a[0], &temp48[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp48len, &temp48[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap

			temp32alen = ScaleExpansionZeroElim(cytabtlen, &cytabt[0], cdytail, &temp32a[0])
			cytabttlen = ScaleExpansionZeroElim(abttlen, &abtt[0], cdytail, &cytabtt[0])
			temp16alen = ScaleExpansionZeroElim(cytabttlen, &cytabtt[0], 2.0*cdy, &temp16a[0])
			temp16blen = ScaleExpansionZeroElim(cytabttlen, &cytabtt[0], cdytail, &temp16b[0])
			temp32blen = FastExpansionSumZeroElim(temp16alen, &temp16a[0], temp16blen, &temp16b[0], &temp32b[0])
			temp64len = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64[0])
			finlength = FastExpansionSumZeroElim(finlength, finnow, temp64len, &temp64[0], finother)
			finswap = finnow
			finnow = finother
			finother = finswap
		}
	}
	return *(*Float)(unsafe.Pointer((uintptr(unsafe.Pointer(finnow)) + floatSize*uintptr(finlength-1)))) // finnow[finlength-1]
}

/*****************************************************************************/
/*                                                                           */
/*  incirclefast()   Approximate 2D incircle test.  Nonrobust.               */
/*  incircleexact()   Exact 2D incircle test.  Robust.                       */
/*  incircleslow()   Another exact 2D incircle test.  Robust.                */
/*  incircle()   Adaptive exact 2D incircle test.  Robust.                   */
/*                                                                           */
/*               Return a positive value if the point pd lies inside the     */
/*               circle passing through pa, pb, and pc; a negative value if  */
/*               it lies outside; and zero if the four points are cocircular.*/
/*               The points pa, pb, and pc must be in counterclockwise       */
/*               order, or the sign of the result will be reversed.          */
/*                                                                           */
/*  Only the first and last routine should be used; the middle two are for   */
/*  timings.                                                                 */
/*                                                                           */
/*  The last three use exact arithmetic to ensure a correct answer.  The     */
/*  result returned is the determinant of a matrix.  In incircle() only,     */
/*  this determinant is computed adaptively, in the sense that exact         */
/*  arithmetic is used only to the degree it is needed to ensure that the    */
/*  returned value has the correct sign.  Hence, incircle() is usually quite */
/*  fast, but will run more slowly when the input points are cocircular or   */
/*  nearly so.                                                               */
/*                                                                           */
/*****************************************************************************/
func Incircle(pa [2]Float, pb [2]Float, pc [2]Float, pd [2]Float) Float {
	// debug.Assert(Orient2d(pa, pb, pc) > 0, "pa, pb, and pc must be in counterclockwise order")
	var adx, bdx, cdx, ady, bdy, cdy Float
	var bdxcdy, cdxbdy, cdxady, adxcdy, adxbdy, bdxady Float
	var alift, blift, clift Float
	var det Float
	var permanent, errbound Float

	adx = pa[0] - pd[0]
	bdx = pb[0] - pd[0]
	cdx = pc[0] - pd[0]
	ady = pa[1] - pd[1]
	bdy = pb[1] - pd[1]
	cdy = pc[1] - pd[1]

	bdxcdy = bdx * cdy
	cdxbdy = cdx * bdy
	alift = adx*adx + ady*ady

	cdxady = cdx * ady
	adxcdy = adx * cdy
	blift = bdx*bdx + bdy*bdy

	adxbdy = adx * bdy
	bdxady = bdx * ady
	clift = cdx*cdx + cdy*cdy

	det = alift*(bdxcdy-cdxbdy) +
		blift*(cdxady-adxcdy) +
		clift*(adxbdy-bdxady)

	permanent = (abs(bdxcdy)+abs(cdxbdy))*alift +
		(abs(cdxady)+abs(adxcdy))*blift +
		(abs(adxbdy)+abs(bdxady))*clift
	errbound = iccerrboundA * permanent
	if (det > errbound) || (-det > errbound) {
		return det
	}

	return IncircleAdapt(pa, pb, pc, pd, permanent)
}

// # 3261 "./predicates.c.txt"
func InsphereFast(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float, pe [3]Float) Float {
	var aex, bex, cex, dex Float
	var aey, bey, cey, dey Float
	var aez, bez, cez, dez Float
	var alift, blift, clift, dlift Float
	var ab, bc, cd, da, ac, bd Float
	var abc, bcd, cda, dab Float

	aex = pa[0] - pe[0]
	bex = pb[0] - pe[0]
	cex = pc[0] - pe[0]
	dex = pd[0] - pe[0]
	aey = pa[1] - pe[1]
	bey = pb[1] - pe[1]
	cey = pc[1] - pe[1]
	dey = pd[1] - pe[1]
	aez = pa[2] - pe[2]
	bez = pb[2] - pe[2]
	cez = pc[2] - pe[2]
	dez = pd[2] - pe[2]

	ab = aex*bey - bex*aey
	bc = bex*cey - cex*bey
	cd = cex*dey - dex*cey
	da = dex*aey - aex*dey

	ac = aex*cey - cex*aey
	bd = bex*dey - dex*bey

	abc = aez*bc - bez*ac + cez*ab
	bcd = bez*cd - cez*bd + dez*bc
	cda = cez*da + dez*ac + aez*cd
	dab = dez*ab + aez*bd + bez*da

	alift = aex*aex + aey*aey + aez*aez
	blift = bex*bex + bey*bey + bez*bez
	clift = cex*cex + cey*cey + cez*cez
	dlift = dex*dex + dey*dey + dez*dez

	return (dlift*abc - clift*dab) + (blift*cda - alift*bcd)
}

func InsphereExact(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float, pe [3]Float) Float {
	var axby1, bxcy1, cxdy1, dxey1, exay1 Float
	var bxay1, cxby1, dxcy1, exdy1, axey1 Float
	var axcy1, bxdy1, cxey1, dxay1, exby1 Float
	var cxay1, dxby1, excy1, axdy1, bxey1 Float
	var axby0, bxcy0, cxdy0, dxey0, exay0 Float
	var bxay0, cxby0, dxcy0, exdy0, axey0 Float
	var axcy0, bxdy0, cxey0, dxay0, exby0 Float
	var cxay0, dxby0, excy0, axdy0, bxey0 Float
	var ab [4]Float
	var bc [4]Float
	var cd [4]Float
	var de [4]Float
	var ea [4]Float
	var ac [4]Float
	var bd [4]Float
	var ce [4]Float
	var da [4]Float
	var eb [4]Float
	var temp8a [8]Float
	var temp8b [8]Float
	var temp16 [16]Float
	var temp8alen, temp8blen, temp16len int
	var abc [24]Float
	var bcd [24]Float
	var cde [24]Float
	var dea [24]Float
	var eab [24]Float
	var abd [24]Float
	var bce [24]Float
	var cda [24]Float
	var deb [24]Float
	var eac [24]Float
	var abclen, bcdlen, cdelen, dealen, eablen int
	var abdlen, bcelen, cdalen, deblen, eaclen int
	var temp48a [48]Float
	var temp48b [48]Float
	var temp48alen, temp48blen int
	var abcd [96]Float
	var bcde [96]Float
	var cdea [96]Float
	var deab [96]Float
	var eabc [96]Float
	var abcdlen, bcdelen, cdealen, deablen, eabclen int
	var temp192 [192]Float
	var det384x [384]Float
	var det384y [384]Float
	var det384z [384]Float
	var xlen, ylen, zlen int
	var detxy [768]Float
	var xylen int
	var adet [1152]Float
	var bdet [1152]Float
	var cdet [1152]Float
	var ddet [1152]Float
	var edet [1152]Float
	var alen, blen, clen, dlen, elen int
	var abdet [2304]Float
	var cddet [2304]Float
	var cdedet [3456]Float
	var ablen, cdlen int
	var deter [5760]Float
	var deterlen int
	var i int

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	axby1 = (Float)(pa[0] * pb[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = axby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axby0 = (alo * blo) - err3
	bxay1 = (Float)(pb[0] * pa[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = bxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxay0 = (alo * blo) - err3
	_i = (Float)(axby0 - bxay0)
	bvirt = (Float)(axby0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay0
	around = axby0 - avirt
	ab[0] = around + bround
	_j = (Float)(axby1 + _i)
	bvirt = (Float)(_j - axby1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axby1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxay1
	around = _0 - avirt
	ab[1] = around + bround
	ab[3] = (Float)(_j + _i)
	bvirt = (Float)(ab[3] - _j)
	avirt = ab[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround

	bxcy1 = (Float)(pb[0] * pc[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = bxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxcy0 = (alo * blo) - err3
	cxby1 = (Float)(pc[0] * pb[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = cxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxby0 = (alo * blo) - err3
	_i = (Float)(bxcy0 - cxby0)
	bvirt = (Float)(bxcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby0
	around = bxcy0 - avirt
	bc[0] = around + bround
	_j = (Float)(bxcy1 + _i)
	bvirt = (Float)(_j - bxcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxby1
	around = _0 - avirt
	bc[1] = around + bround
	bc[3] = (Float)(_j + _i)
	bvirt = (Float)(bc[3] - _j)
	avirt = bc[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround

	cxdy1 = (Float)(pc[0] * pd[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = cxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxdy0 = (alo * blo) - err3
	dxcy1 = (Float)(pd[0] * pc[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = dxcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxcy0 = (alo * blo) - err3
	_i = (Float)(cxdy0 - dxcy0)
	bvirt = (Float)(cxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy0
	around = cxdy0 - avirt
	cd[0] = around + bround
	_j = (Float)(cxdy1 + _i)
	bvirt = (Float)(_j - cxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxcy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxcy1
	around = _0 - avirt
	cd[1] = around + bround
	cd[3] = (Float)(_j + _i)
	bvirt = (Float)(cd[3] - _j)
	avirt = cd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	cd[2] = around + bround

	dxey1 = (Float)(pd[0] * pe[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pe[1])
	abig = (Float)(c - pe[1])
	bhi = c - abig
	blo = pe[1] - bhi
	err1 = dxey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxey0 = (alo * blo) - err3
	exdy1 = (Float)(pe[0] * pd[1])
	c = (Float)(splitter * pe[0])
	abig = (Float)(c - pe[0])
	ahi = c - abig
	alo = pe[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = exdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	exdy0 = (alo * blo) - err3
	_i = (Float)(dxey0 - exdy0)
	bvirt = (Float)(dxey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - exdy0
	around = dxey0 - avirt
	de[0] = around + bround
	_j = (Float)(dxey1 + _i)
	bvirt = (Float)(_j - dxey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = dxey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - exdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - exdy1
	around = _0 - avirt
	de[1] = around + bround
	de[3] = (Float)(_j + _i)
	bvirt = (Float)(de[3] - _j)
	avirt = de[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	de[2] = around + bround

	exay1 = (Float)(pe[0] * pa[1])
	c = (Float)(splitter * pe[0])
	abig = (Float)(c - pe[0])
	ahi = c - abig
	alo = pe[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = exay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	exay0 = (alo * blo) - err3
	axey1 = (Float)(pa[0] * pe[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pe[1])
	abig = (Float)(c - pe[1])
	bhi = c - abig
	blo = pe[1] - bhi
	err1 = axey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axey0 = (alo * blo) - err3
	_i = (Float)(exay0 - axey0)
	bvirt = (Float)(exay0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axey0
	around = exay0 - avirt
	ea[0] = around + bround
	_j = (Float)(exay1 + _i)
	bvirt = (Float)(_j - exay1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = exay1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - axey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axey1
	around = _0 - avirt
	ea[1] = around + bround
	ea[3] = (Float)(_j + _i)
	bvirt = (Float)(ea[3] - _j)
	avirt = ea[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ea[2] = around + bround

	axcy1 = (Float)(pa[0] * pc[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = axcy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axcy0 = (alo * blo) - err3
	cxay1 = (Float)(pc[0] * pa[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = cxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxay0 = (alo * blo) - err3
	_i = (Float)(axcy0 - cxay0)
	bvirt = (Float)(axcy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay0
	around = axcy0 - avirt
	ac[0] = around + bround
	_j = (Float)(axcy1 + _i)
	bvirt = (Float)(_j - axcy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = axcy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cxay1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cxay1
	around = _0 - avirt
	ac[1] = around + bround
	ac[3] = (Float)(_j + _i)
	bvirt = (Float)(ac[3] - _j)
	avirt = ac[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ac[2] = around + bround

	bxdy1 = (Float)(pb[0] * pd[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = bxdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxdy0 = (alo * blo) - err3
	dxby1 = (Float)(pd[0] * pb[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = dxby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxby0 = (alo * blo) - err3
	_i = (Float)(bxdy0 - dxby0)
	bvirt = (Float)(bxdy0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby0
	around = bxdy0 - avirt
	bd[0] = around + bround
	_j = (Float)(bxdy1 + _i)
	bvirt = (Float)(_j - bxdy1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bxdy1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dxby1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dxby1
	around = _0 - avirt
	bd[1] = around + bround
	bd[3] = (Float)(_j + _i)
	bvirt = (Float)(bd[3] - _j)
	avirt = bd[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bd[2] = around + bround

	cxey1 = (Float)(pc[0] * pe[1])
	c = (Float)(splitter * pc[0])
	abig = (Float)(c - pc[0])
	ahi = c - abig
	alo = pc[0] - ahi
	c = (Float)(splitter * pe[1])
	abig = (Float)(c - pe[1])
	bhi = c - abig
	blo = pe[1] - bhi
	err1 = cxey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cxey0 = (alo * blo) - err3
	excy1 = (Float)(pe[0] * pc[1])
	c = (Float)(splitter * pe[0])
	abig = (Float)(c - pe[0])
	ahi = c - abig
	alo = pe[0] - ahi
	c = (Float)(splitter * pc[1])
	abig = (Float)(c - pc[1])
	bhi = c - abig
	blo = pc[1] - bhi
	err1 = excy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	excy0 = (alo * blo) - err3
	_i = (Float)(cxey0 - excy0)
	bvirt = (Float)(cxey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - excy0
	around = cxey0 - avirt
	ce[0] = around + bround
	_j = (Float)(cxey1 + _i)
	bvirt = (Float)(_j - cxey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cxey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - excy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - excy1
	around = _0 - avirt
	ce[1] = around + bround
	ce[3] = (Float)(_j + _i)
	bvirt = (Float)(ce[3] - _j)
	avirt = ce[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ce[2] = around + bround

	dxay1 = (Float)(pd[0] * pa[1])
	c = (Float)(splitter * pd[0])
	abig = (Float)(c - pd[0])
	ahi = c - abig
	alo = pd[0] - ahi
	c = (Float)(splitter * pa[1])
	abig = (Float)(c - pa[1])
	bhi = c - abig
	blo = pa[1] - bhi
	err1 = dxay1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dxay0 = (alo * blo) - err3
	axdy1 = (Float)(pa[0] * pd[1])
	c = (Float)(splitter * pa[0])
	abig = (Float)(c - pa[0])
	ahi = c - abig
	alo = pa[0] - ahi
	c = (Float)(splitter * pd[1])
	abig = (Float)(c - pd[1])
	bhi = c - abig
	blo = pd[1] - bhi
	err1 = axdy1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	axdy0 = (alo * blo) - err3
	_i = (Float)(dxay0 - axdy0)
	bvirt = (Float)(dxay0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy0
	around = dxay0 - avirt
	da[0] = around + bround
	_j = (Float)(dxay1 + _i)
	bvirt = (Float)(_j - dxay1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = dxay1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - axdy1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - axdy1
	around = _0 - avirt
	da[1] = around + bround
	da[3] = (Float)(_j + _i)
	bvirt = (Float)(da[3] - _j)
	avirt = da[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	da[2] = around + bround

	exby1 = (Float)(pe[0] * pb[1])
	c = (Float)(splitter * pe[0])
	abig = (Float)(c - pe[0])
	ahi = c - abig
	alo = pe[0] - ahi
	c = (Float)(splitter * pb[1])
	abig = (Float)(c - pb[1])
	bhi = c - abig
	blo = pb[1] - bhi
	err1 = exby1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	exby0 = (alo * blo) - err3
	bxey1 = (Float)(pb[0] * pe[1])
	c = (Float)(splitter * pb[0])
	abig = (Float)(c - pb[0])
	ahi = c - abig
	alo = pb[0] - ahi
	c = (Float)(splitter * pe[1])
	abig = (Float)(c - pe[1])
	bhi = c - abig
	blo = pe[1] - bhi
	err1 = bxey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bxey0 = (alo * blo) - err3
	_i = (Float)(exby0 - bxey0)
	bvirt = (Float)(exby0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxey0
	around = exby0 - avirt
	eb[0] = around + bround
	_j = (Float)(exby1 + _i)
	bvirt = (Float)(_j - exby1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = exby1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bxey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bxey1
	around = _0 - avirt
	eb[1] = around + bround
	eb[3] = (Float)(_j + _i)
	bvirt = (Float)(eb[3] - _j)
	avirt = eb[3] - bvirt
	bround = _i - bvirt
	around = _j - avirt
	eb[2] = around + bround

	temp8alen = ScaleExpansionZeroElim(4, &bc[0], pa[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ac[0], -pb[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &ab[0], pc[2], &temp8a[0])
	abclen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &abc[0])

	temp8alen = ScaleExpansionZeroElim(4, &cd[0], pb[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &bd[0], -pc[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &bc[0], pd[2], &temp8a[0])
	bcdlen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &bcd[0])

	temp8alen = ScaleExpansionZeroElim(4, &de[0], pc[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ce[0], -pd[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &cd[0], pe[2], &temp8a[0])
	cdelen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &cde[0])

	temp8alen = ScaleExpansionZeroElim(4, &ea[0], pd[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &da[0], -pe[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &de[0], pa[2], &temp8a[0])
	dealen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &dea[0])

	temp8alen = ScaleExpansionZeroElim(4, &ab[0], pe[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &eb[0], -pa[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &ea[0], pb[2], &temp8a[0])
	eablen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &eab[0])

	temp8alen = ScaleExpansionZeroElim(4, &bd[0], pa[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &da[0], pb[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &ab[0], pd[2], &temp8a[0])
	abdlen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &abd[0])

	temp8alen = ScaleExpansionZeroElim(4, &ce[0], pb[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &eb[0], pc[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &bc[0], pe[2], &temp8a[0])
	bcelen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &bce[0])

	temp8alen = ScaleExpansionZeroElim(4, &da[0], pc[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ac[0], pd[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &cd[0], pa[2], &temp8a[0])
	cdalen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &cda[0])

	temp8alen = ScaleExpansionZeroElim(4, &eb[0], pd[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &bd[0], pe[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &de[0], pb[2], &temp8a[0])
	deblen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &deb[0])

	temp8alen = ScaleExpansionZeroElim(4, &ac[0], pe[2], &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ce[0], pa[2], &temp8b[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp8alen = ScaleExpansionZeroElim(4, &ea[0], pc[2], &temp8a[0])
	eaclen = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp16len, &temp16[0], &eac[0])

	temp48alen = FastExpansionSumZeroElim(cdelen, &cde[0], bcelen, &bce[0], &temp48a[0])
	temp48blen = FastExpansionSumZeroElim(deblen, &deb[0], bcdlen, &bcd[0], &temp48b[0])
	for i = 0; i < temp48blen; i++ {
		temp48b[i] = -temp48b[i]
	}
	bcdelen = FastExpansionSumZeroElim(temp48alen, &temp48a[0], temp48blen, &temp48b[0], &bcde[0])
	xlen = ScaleExpansionZeroElim(bcdelen, &bcde[0], pa[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(xlen, &temp192[0], pa[0], &det384x[0])
	ylen = ScaleExpansionZeroElim(bcdelen, &bcde[0], pa[1], &temp192[0])
	ylen = ScaleExpansionZeroElim(ylen, &temp192[0], pa[1], &det384y[0])
	zlen = ScaleExpansionZeroElim(bcdelen, &bcde[0], pa[2], &temp192[0])
	zlen = ScaleExpansionZeroElim(zlen, &temp192[0], pa[2], &det384z[0])
	xylen = FastExpansionSumZeroElim(xlen, &det384x[0], ylen, &det384y[0], &detxy[0])
	alen = FastExpansionSumZeroElim(xylen, &detxy[0], zlen, &det384z[0], &adet[0])

	temp48alen = FastExpansionSumZeroElim(dealen, &dea[0], cdalen, &cda[0], &temp48a[0])
	temp48blen = FastExpansionSumZeroElim(eaclen, &eac[0], cdelen, &cde[0], &temp48b[0])
	for i = 0; i < temp48blen; i++ {
		temp48b[i] = -temp48b[i]
	}
	cdealen = FastExpansionSumZeroElim(temp48alen, &temp48a[0], temp48blen, &temp48b[0], &cdea[0])
	xlen = ScaleExpansionZeroElim(cdealen, &cdea[0], pb[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(xlen, &temp192[0], pb[0], &det384x[0])
	ylen = ScaleExpansionZeroElim(cdealen, &cdea[0], pb[1], &temp192[0])
	ylen = ScaleExpansionZeroElim(ylen, &temp192[0], pb[1], &det384y[0])
	zlen = ScaleExpansionZeroElim(cdealen, &cdea[0], pb[2], &temp192[0])
	zlen = ScaleExpansionZeroElim(zlen, &temp192[0], pb[2], &det384z[0])
	xylen = FastExpansionSumZeroElim(xlen, &det384x[0], ylen, &det384y[0], &detxy[0])
	blen = FastExpansionSumZeroElim(xylen, &detxy[0], zlen, &det384z[0], &bdet[0])

	temp48alen = FastExpansionSumZeroElim(eablen, &eab[0], deblen, &deb[0], &temp48a[0])
	temp48blen = FastExpansionSumZeroElim(abdlen, &abd[0], dealen, &dea[0], &temp48b[0])
	for i = 0; i < temp48blen; i++ {
		temp48b[i] = -temp48b[i]
	}
	deablen = FastExpansionSumZeroElim(temp48alen, &temp48a[0], temp48blen, &temp48b[0], &deab[0])
	xlen = ScaleExpansionZeroElim(deablen, &deab[0], pc[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(xlen, &temp192[0], pc[0], &det384x[0])
	ylen = ScaleExpansionZeroElim(deablen, &deab[0], pc[1], &temp192[0])
	ylen = ScaleExpansionZeroElim(ylen, &temp192[0], pc[1], &det384y[0])
	zlen = ScaleExpansionZeroElim(deablen, &deab[0], pc[2], &temp192[0])
	zlen = ScaleExpansionZeroElim(zlen, &temp192[0], pc[2], &det384z[0])
	xylen = FastExpansionSumZeroElim(xlen, &det384x[0], ylen, &det384y[0], &detxy[0])
	clen = FastExpansionSumZeroElim(xylen, &detxy[0], zlen, &det384z[0], &cdet[0])

	temp48alen = FastExpansionSumZeroElim(abclen, &abc[0], eaclen, &eac[0], &temp48a[0])
	temp48blen = FastExpansionSumZeroElim(bcelen, &bce[0], eablen, &eab[0], &temp48b[0])
	for i = 0; i < temp48blen; i++ {
		temp48b[i] = -temp48b[i]
	}
	eabclen = FastExpansionSumZeroElim(temp48alen, &temp48a[0], temp48blen, &temp48b[0], &eabc[0])
	xlen = ScaleExpansionZeroElim(eabclen, &eabc[0], pd[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(xlen, &temp192[0], pd[0], &det384x[0])
	ylen = ScaleExpansionZeroElim(eabclen, &eabc[0], pd[1], &temp192[0])
	ylen = ScaleExpansionZeroElim(ylen, &temp192[0], pd[1], &det384y[0])
	zlen = ScaleExpansionZeroElim(eabclen, &eabc[0], pd[2], &temp192[0])
	zlen = ScaleExpansionZeroElim(zlen, &temp192[0], pd[2], &det384z[0])
	xylen = FastExpansionSumZeroElim(xlen, &det384x[0], ylen, &det384y[0], &detxy[0])
	dlen = FastExpansionSumZeroElim(xylen, &detxy[0], zlen, &det384z[0], &ddet[0])

	temp48alen = FastExpansionSumZeroElim(bcdlen, &bcd[0], abdlen, &abd[0], &temp48a[0])
	temp48blen = FastExpansionSumZeroElim(cdalen, &cda[0], abclen, &abc[0], &temp48b[0])
	for i = 0; i < temp48blen; i++ {
		temp48b[i] = -temp48b[i]
	}
	abcdlen = FastExpansionSumZeroElim(temp48alen, &temp48a[0], temp48blen, &temp48b[0], &abcd[0])
	xlen = ScaleExpansionZeroElim(abcdlen, &abcd[0], pe[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(xlen, &temp192[0], pe[0], &det384x[0])
	ylen = ScaleExpansionZeroElim(abcdlen, &abcd[0], pe[1], &temp192[0])
	ylen = ScaleExpansionZeroElim(ylen, &temp192[0], pe[1], &det384y[0])
	zlen = ScaleExpansionZeroElim(abcdlen, &abcd[0], pe[2], &temp192[0])
	zlen = ScaleExpansionZeroElim(zlen, &temp192[0], pe[2], &det384z[0])
	xylen = FastExpansionSumZeroElim(xlen, &det384x[0], ylen, &det384y[0], &detxy[0])
	elen = FastExpansionSumZeroElim(xylen, &detxy[0], zlen, &det384z[0], &edet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	cdlen = FastExpansionSumZeroElim(clen, &cdet[0], dlen, &ddet[0], &cddet[0])
	cdelen = FastExpansionSumZeroElim(cdlen, &cddet[0], elen, &edet[0], &cdedet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], cdelen, &cdedet[0], &deter[0])

	return deter[deterlen-1]
}

func InsphereSlow(pa, pb, pc, pd, pe [3]Float) Float {
	var aex, bex, cex, dex, aey, bey, cey, dey, aez, bez, cez, dez Float
	var aextail, bextail, cextail, dextail Float
	var aeytail, beytail, ceytail, deytail Float
	var aeztail, beztail, ceztail, deztail Float
	var negate, negatetail Float
	var axby7, bxcy7, cxdy7, dxay7, axcy7, bxdy7 Float
	var bxay7, cxby7, dxcy7, axdy7, cxay7, dxby7 Float
	var axby, bxcy, cxdy, dxay, axcy, bxdy [8]Float
	var bxay, cxby, dxcy, axdy, cxay, dxby [8]Float
	var ab, bc, cd, da, ac, bd [16]Float
	var ablen, bclen, cdlen, dalen, aclen, bdlen int
	var temp32a, temp32b [32]Float
	var temp64a, temp64b, temp64c [64]Float
	var temp32alen, temp32blen, temp64alen, temp64blen, temp64clen int
	var temp128 [128]Float
	var temp192 [192]Float
	var temp128len, temp192len int
	var detx [384]Float
	var detxx [768]Float
	var detxt [384]Float
	var detxxt [768]Float
	var detxtxt [768]Float
	var xlen, xxlen, xtlen, xxtlen, xtxtlen int
	var x1 [1536]Float
	var x2 [2304]Float
	var x1len, x2len int
	var dety [384]Float
	var detyy [768]Float
	var detyt [384]Float
	var detyyt [768]Float
	var detytyt [768]Float
	var ylen, yylen, ytlen, yytlen, ytytlen int
	var y1 [1536]Float
	var y2 [2304]Float
	var y1len, y2len int
	var detz [384]Float
	var detzz [768]Float
	var detzt [384]Float
	var detzzt [768]Float
	var detztzt [768]Float
	var zlen, zzlen, ztlen, zztlen, ztztlen int
	var z1 [1536]Float
	var z2 [2304]Float
	var z1len, z2len int
	var detxy [4608]Float
	var xylen int
	var adet [6912]Float
	var bdet [6912]Float
	var cdet [6912]Float
	var ddet [6912]Float
	var alen, blen, clen, dlen int
	var abdet [13824]Float
	var cddet [13824]Float
	var deter [27648]Float
	var deterlen int
	var i int
	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var a0hi, a0lo, a1hi, a1lo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j, _k, _l, _m, _n Float
	var _0, _1, _2 Float

	aex = (Float)(pa[0] - pe[0])
	bvirt = (Float)(pa[0] - aex)
	avirt = aex + bvirt
	bround = bvirt - pe[0]
	around = pa[0] - avirt
	aextail = around + bround
	aey = (Float)(pa[1] - pe[1])
	bvirt = (Float)(pa[1] - aey)
	avirt = aey + bvirt
	bround = bvirt - pe[1]
	around = pa[1] - avirt
	aeytail = around + bround
	aez = (Float)(pa[2] - pe[2])
	bvirt = (Float)(pa[2] - aez)
	avirt = aez + bvirt
	bround = bvirt - pe[2]
	around = pa[2] - avirt
	aeztail = around + bround
	bex = (Float)(pb[0] - pe[0])
	bvirt = (Float)(pb[0] - bex)
	avirt = bex + bvirt
	bround = bvirt - pe[0]
	around = pb[0] - avirt
	bextail = around + bround
	bey = (Float)(pb[1] - pe[1])
	bvirt = (Float)(pb[1] - bey)
	avirt = bey + bvirt
	bround = bvirt - pe[1]
	around = pb[1] - avirt
	beytail = around + bround
	bez = (Float)(pb[2] - pe[2])
	bvirt = (Float)(pb[2] - bez)
	avirt = bez + bvirt
	bround = bvirt - pe[2]
	around = pb[2] - avirt
	beztail = around + bround
	cex = (Float)(pc[0] - pe[0])
	bvirt = (Float)(pc[0] - cex)
	avirt = cex + bvirt
	bround = bvirt - pe[0]
	around = pc[0] - avirt
	cextail = around + bround
	cey = (Float)(pc[1] - pe[1])
	bvirt = (Float)(pc[1] - cey)
	avirt = cey + bvirt
	bround = bvirt - pe[1]
	around = pc[1] - avirt
	ceytail = around + bround
	cez = (Float)(pc[2] - pe[2])
	bvirt = (Float)(pc[2] - cez)
	avirt = cez + bvirt
	bround = bvirt - pe[2]
	around = pc[2] - avirt
	ceztail = around + bround
	dex = (Float)(pd[0] - pe[0])
	bvirt = (Float)(pd[0] - dex)
	avirt = dex + bvirt
	bround = bvirt - pe[0]
	around = pd[0] - avirt
	dextail = around + bround
	dey = (Float)(pd[1] - pe[1])
	bvirt = (Float)(pd[1] - dey)
	avirt = dey + bvirt
	bround = bvirt - pe[1]
	around = pd[1] - avirt
	deytail = around + bround
	dez = (Float)(pd[2] - pe[2])
	bvirt = (Float)(pd[2] - dez)
	avirt = dez + bvirt
	bround = bvirt - pe[2]
	around = pd[2] - avirt
	deztail = around + bround

	c = (Float)(splitter * aextail)
	abig = (Float)(c - aextail)
	a0hi = c - abig
	a0lo = aextail - a0hi
	c = (Float)(splitter * beytail)
	abig = (Float)(c - beytail)
	bhi = c - abig
	blo = beytail - bhi
	_i = (Float)(aextail * beytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	a1hi = c - abig
	a1lo = aex - a1hi
	_j = (Float)(aex * beytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * bey)
	abig = (Float)(c - bey)
	bhi = c - abig
	blo = bey - bhi
	_i = (Float)(aextail * bey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(aex * bey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axby[5] = around + bround
	axby7 = (Float)(_m + _k)
	bvirt = (Float)(axby7 - _m)
	avirt = axby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axby[6] = around + bround

	axby[7] = axby7
	negate = -aey
	negatetail = -aeytail
	c = (Float)(splitter * bextail)
	abig = (Float)(c - bextail)
	a0hi = c - abig
	a0lo = bextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(bextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	a1hi = c - abig
	a1lo = bex - a1hi
	_j = (Float)(bex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(bextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxay[5] = around + bround
	bxay7 = (Float)(_m + _k)
	bvirt = (Float)(bxay7 - _m)
	avirt = bxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxay[6] = around + bround

	bxay[7] = bxay7
	ablen = FastExpansionSumZeroElim(8, &axby[0], 8, &bxay[0], &ab[0])
	c = (Float)(splitter * bextail)
	abig = (Float)(c - bextail)
	a0hi = c - abig
	a0lo = bextail - a0hi
	c = (Float)(splitter * ceytail)
	abig = (Float)(c - ceytail)
	bhi = c - abig
	blo = ceytail - bhi
	_i = (Float)(bextail * ceytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	a1hi = c - abig
	a1lo = bex - a1hi
	_j = (Float)(bex * ceytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * cey)
	abig = (Float)(c - cey)
	bhi = c - abig
	blo = cey - bhi
	_i = (Float)(bextail * cey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bex * cey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxcy[5] = around + bround
	bxcy7 = (Float)(_m + _k)
	bvirt = (Float)(bxcy7 - _m)
	avirt = bxcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxcy[6] = around + bround

	bxcy[7] = bxcy7
	negate = -bey
	negatetail = -beytail
	c = (Float)(splitter * cextail)
	abig = (Float)(c - cextail)
	a0hi = c - abig
	a0lo = cextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(cextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	a1hi = c - abig
	a1lo = cex - a1hi
	_j = (Float)(cex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(cextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxby[5] = around + bround
	cxby7 = (Float)(_m + _k)
	bvirt = (Float)(cxby7 - _m)
	avirt = cxby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxby[6] = around + bround

	cxby[7] = cxby7
	bclen = FastExpansionSumZeroElim(8, &bxcy[0], 8, &cxby[0], &bc[0])
	c = (Float)(splitter * cextail)
	abig = (Float)(c - cextail)
	a0hi = c - abig
	a0lo = cextail - a0hi
	c = (Float)(splitter * deytail)
	abig = (Float)(c - deytail)
	bhi = c - abig
	blo = deytail - bhi
	_i = (Float)(cextail * deytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxdy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	a1hi = c - abig
	a1lo = cex - a1hi
	_j = (Float)(cex * deytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * dey)
	abig = (Float)(c - dey)
	bhi = c - abig
	blo = dey - bhi
	_i = (Float)(cextail * dey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxdy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cex * dey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxdy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxdy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxdy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxdy[5] = around + bround
	cxdy7 = (Float)(_m + _k)
	bvirt = (Float)(cxdy7 - _m)
	avirt = cxdy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxdy[6] = around + bround

	cxdy[7] = cxdy7
	negate = -cey
	negatetail = -ceytail
	c = (Float)(splitter * dextail)
	abig = (Float)(c - dextail)
	a0hi = c - abig
	a0lo = dextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(dextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	dxcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	a1hi = c - abig
	a1lo = dex - a1hi
	_j = (Float)(dex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(dextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(dex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	dxcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	dxcy[5] = around + bround
	dxcy7 = (Float)(_m + _k)
	bvirt = (Float)(dxcy7 - _m)
	avirt = dxcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	dxcy[6] = around + bround

	dxcy[7] = dxcy7
	cdlen = FastExpansionSumZeroElim(8, &cxdy[0], 8, &dxcy[0], &cd[0])
	c = (Float)(splitter * dextail)
	abig = (Float)(c - dextail)
	a0hi = c - abig
	a0lo = dextail - a0hi
	c = (Float)(splitter * aeytail)
	abig = (Float)(c - aeytail)
	bhi = c - abig
	blo = aeytail - bhi
	_i = (Float)(dextail * aeytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	dxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	a1hi = c - abig
	a1lo = dex - a1hi
	_j = (Float)(dex * aeytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * aey)
	abig = (Float)(c - aey)
	bhi = c - abig
	blo = aey - bhi
	_i = (Float)(dextail * aey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(dex * aey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	dxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	dxay[5] = around + bround
	dxay7 = (Float)(_m + _k)
	bvirt = (Float)(dxay7 - _m)
	avirt = dxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	dxay[6] = around + bround

	dxay[7] = dxay7
	negate = -dey
	negatetail = -deytail
	c = (Float)(splitter * aextail)
	abig = (Float)(c - aextail)
	a0hi = c - abig
	a0lo = aextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(aextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axdy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	a1hi = c - abig
	a1lo = aex - a1hi
	_j = (Float)(aex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(aextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axdy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(aex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axdy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axdy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axdy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axdy[5] = around + bround
	axdy7 = (Float)(_m + _k)
	bvirt = (Float)(axdy7 - _m)
	avirt = axdy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axdy[6] = around + bround

	axdy[7] = axdy7
	dalen = FastExpansionSumZeroElim(8, &dxay[0], 8, &axdy[0], &da[0])
	c = (Float)(splitter * aextail)
	abig = (Float)(c - aextail)
	a0hi = c - abig
	a0lo = aextail - a0hi
	c = (Float)(splitter * ceytail)
	abig = (Float)(c - ceytail)
	bhi = c - abig
	blo = ceytail - bhi
	_i = (Float)(aextail * ceytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	axcy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	a1hi = c - abig
	a1lo = aex - a1hi
	_j = (Float)(aex * ceytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * cey)
	abig = (Float)(c - cey)
	bhi = c - abig
	blo = cey - bhi
	_i = (Float)(aextail * cey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(aex * cey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	axcy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	axcy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	axcy[5] = around + bround
	axcy7 = (Float)(_m + _k)
	bvirt = (Float)(axcy7 - _m)
	avirt = axcy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	axcy[6] = around + bround

	axcy[7] = axcy7
	negate = -aey
	negatetail = -aeytail
	c = (Float)(splitter * cextail)
	abig = (Float)(c - cextail)
	a0hi = c - abig
	a0lo = cextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(cextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	cxay[0] = (a0lo * blo) - err3
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	a1hi = c - abig
	a1lo = cex - a1hi
	_j = (Float)(cex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(cextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(cex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	cxay[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	cxay[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	cxay[5] = around + bround
	cxay7 = (Float)(_m + _k)
	bvirt = (Float)(cxay7 - _m)
	avirt = cxay7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	cxay[6] = around + bround

	cxay[7] = cxay7
	aclen = FastExpansionSumZeroElim(8, &axcy[0], 8, &cxay[0], &ac[0])
	c = (Float)(splitter * bextail)
	abig = (Float)(c - bextail)
	a0hi = c - abig
	a0lo = bextail - a0hi
	c = (Float)(splitter * deytail)
	abig = (Float)(c - deytail)
	bhi = c - abig
	blo = deytail - bhi
	_i = (Float)(bextail * deytail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	bxdy[0] = (a0lo * blo) - err3
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	a1hi = c - abig
	a1lo = bex - a1hi
	_j = (Float)(bex * deytail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * dey)
	abig = (Float)(c - dey)
	bhi = c - abig
	blo = dey - bhi
	_i = (Float)(bextail * dey)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxdy[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(bex * dey)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxdy[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	bxdy[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	bxdy[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	bxdy[5] = around + bround
	bxdy7 = (Float)(_m + _k)
	bvirt = (Float)(bxdy7 - _m)
	avirt = bxdy7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	bxdy[6] = around + bround

	bxdy[7] = bxdy7
	negate = -bey
	negatetail = -beytail
	c = (Float)(splitter * dextail)
	abig = (Float)(c - dextail)
	a0hi = c - abig
	a0lo = dextail - a0hi
	c = (Float)(splitter * negatetail)
	abig = (Float)(c - negatetail)
	bhi = c - abig
	blo = negatetail - bhi
	_i = (Float)(dextail * negatetail)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	dxby[0] = (a0lo * blo) - err3
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	a1hi = c - abig
	a1lo = dex - a1hi
	_j = (Float)(dex * negatetail)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_k = (Float)(_i + _0)
	bvirt = (Float)(_k - _i)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_1 = around + bround
	_l = (Float)(_j + _k)
	bvirt = _l - _j
	_2 = _k - bvirt
	c = (Float)(splitter * negate)
	abig = (Float)(c - negate)
	bhi = c - abig
	blo = negate - bhi
	_i = (Float)(dextail * negate)
	err1 = _i - (a0hi * bhi)
	err2 = err1 - (a0lo * bhi)
	err3 = err2 - (a0hi * blo)
	_0 = (a0lo * blo) - err3
	_k = (Float)(_1 + _0)
	bvirt = (Float)(_k - _1)
	avirt = _k - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxby[1] = around + bround
	_j = (Float)(_2 + _k)
	bvirt = (Float)(_j - _2)
	avirt = _j - bvirt
	bround = _k - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _j)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _j - bvirt
	around = _l - avirt
	_2 = around + bround
	_j = (Float)(dex * negate)
	err1 = _j - (a1hi * bhi)
	err2 = err1 - (a1lo * bhi)
	err3 = err2 - (a1hi * blo)
	_0 = (a1lo * blo) - err3
	_n = (Float)(_i + _0)
	bvirt = (Float)(_n - _i)
	avirt = _n - bvirt
	bround = _0 - bvirt
	around = _i - avirt
	_0 = around + bround
	_i = (Float)(_1 + _0)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxby[2] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	_1 = around + bround
	_l = (Float)(_m + _k)
	bvirt = (Float)(_l - _m)
	avirt = _l - bvirt
	bround = _k - bvirt
	around = _m - avirt
	_2 = around + bround
	_k = (Float)(_j + _n)
	bvirt = (Float)(_k - _j)
	avirt = _k - bvirt
	bround = _n - bvirt
	around = _j - avirt
	_0 = around + bround
	_j = (Float)(_1 + _0)
	bvirt = (Float)(_j - _1)
	avirt = _j - bvirt
	bround = _0 - bvirt
	around = _1 - avirt
	dxby[3] = around + bround
	_i = (Float)(_2 + _j)
	bvirt = (Float)(_i - _2)
	avirt = _i - bvirt
	bround = _j - bvirt
	around = _2 - avirt
	_1 = around + bround
	_m = (Float)(_l + _i)
	bvirt = (Float)(_m - _l)
	avirt = _m - bvirt
	bround = _i - bvirt
	around = _l - avirt
	_2 = around + bround
	_i = (Float)(_1 + _k)
	bvirt = (Float)(_i - _1)
	avirt = _i - bvirt
	bround = _k - bvirt
	around = _1 - avirt
	dxby[4] = around + bround
	_k = (Float)(_2 + _i)
	bvirt = (Float)(_k - _2)
	avirt = _k - bvirt
	bround = _i - bvirt
	around = _2 - avirt
	dxby[5] = around + bround
	dxby7 = (Float)(_m + _k)
	bvirt = (Float)(dxby7 - _m)
	avirt = dxby7 - bvirt
	bround = _k - bvirt
	around = _m - avirt
	dxby[6] = around + bround

	dxby[7] = dxby7
	bdlen = FastExpansionSumZeroElim(8, &bxdy[0], 8, &dxby[0], &bd[0])

	temp32alen = ScaleExpansionZeroElim(cdlen, &cd[0], -bez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(cdlen, &cd[0], -beztail, &temp32b[0])
	temp64alen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64a[0])
	temp32alen = ScaleExpansionZeroElim(bdlen, &bd[0], cez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(bdlen, &bd[0], ceztail, &temp32b[0])
	temp64blen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64b[0])
	temp32alen = ScaleExpansionZeroElim(bclen, &bc[0], -dez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(bclen, &bc[0], -deztail, &temp32b[0])
	temp64clen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64c[0])
	temp128len = FastExpansionSumZeroElim(temp64alen, &temp64a[0], temp64blen, &temp64b[0], &temp128[0])
	temp192len = FastExpansionSumZeroElim(temp64clen, &temp64c[0], temp128len, &temp128[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(temp192len, &temp192[0], aex, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], aex, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp192len, &temp192[0], aextail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], aex, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], aextail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])
	ylen = ScaleExpansionZeroElim(temp192len, &temp192[0], aey, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], aey, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp192len, &temp192[0], aeytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], aey, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], aeytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])
	zlen = ScaleExpansionZeroElim(temp192len, &temp192[0], aez, &detz[0])
	zzlen = ScaleExpansionZeroElim(zlen, &detz[0], aez, &detzz[0])
	ztlen = ScaleExpansionZeroElim(temp192len, &temp192[0], aeztail, &detzt[0])
	zztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], aez, &detzzt[0])
	for i = 0; i < zztlen; i++ {
		detzzt[i] *= 2.0
	}
	ztztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], aeztail, &detztzt[0])
	z1len = FastExpansionSumZeroElim(zzlen, &detzz[0], zztlen, &detzzt[0], &z1[0])
	z2len = FastExpansionSumZeroElim(z1len, &z1[0], ztztlen, &detztzt[0], &z2[0])
	xylen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &detxy[0])
	alen = FastExpansionSumZeroElim(z2len, &z2[0], xylen, &detxy[0], &adet[0])

	temp32alen = ScaleExpansionZeroElim(dalen, &da[0], cez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(dalen, &da[0], ceztail, &temp32b[0])
	temp64alen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64a[0])
	temp32alen = ScaleExpansionZeroElim(aclen, &ac[0], dez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(aclen, &ac[0], deztail, &temp32b[0])
	temp64blen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64b[0])
	temp32alen = ScaleExpansionZeroElim(cdlen, &cd[0], aez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(cdlen, &cd[0], aeztail, &temp32b[0])
	temp64clen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64c[0])
	temp128len = FastExpansionSumZeroElim(temp64alen, &temp64a[0], temp64blen, &temp64b[0], &temp128[0])
	temp192len = FastExpansionSumZeroElim(temp64clen, &temp64c[0], temp128len, &temp128[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(temp192len, &temp192[0], bex, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], bex, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp192len, &temp192[0], bextail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], bex, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], bextail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])
	ylen = ScaleExpansionZeroElim(temp192len, &temp192[0], bey, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], bey, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp192len, &temp192[0], beytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], bey, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], beytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])
	zlen = ScaleExpansionZeroElim(temp192len, &temp192[0], bez, &detz[0])
	zzlen = ScaleExpansionZeroElim(zlen, &detz[0], bez, &detzz[0])
	ztlen = ScaleExpansionZeroElim(temp192len, &temp192[0], beztail, &detzt[0])
	zztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], bez, &detzzt[0])
	for i = 0; i < zztlen; i++ {
		detzzt[i] *= 2.0
	}
	ztztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], beztail, &detztzt[0])
	z1len = FastExpansionSumZeroElim(zzlen, &detzz[0], zztlen, &detzzt[0], &z1[0])
	z2len = FastExpansionSumZeroElim(z1len, &z1[0], ztztlen, &detztzt[0], &z2[0])
	xylen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &detxy[0])
	blen = FastExpansionSumZeroElim(z2len, &z2[0], xylen, &detxy[0], &bdet[0])

	temp32alen = ScaleExpansionZeroElim(ablen, &ab[0], -dez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(ablen, &ab[0], -deztail, &temp32b[0])
	temp64alen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64a[0])
	temp32alen = ScaleExpansionZeroElim(bdlen, &bd[0], -aez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(bdlen, &bd[0], -aeztail, &temp32b[0])
	temp64blen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64b[0])
	temp32alen = ScaleExpansionZeroElim(dalen, &da[0], -bez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(dalen, &da[0], -beztail, &temp32b[0])
	temp64clen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64c[0])
	temp128len = FastExpansionSumZeroElim(temp64alen, &temp64a[0], temp64blen, &temp64b[0], &temp128[0])
	temp192len = FastExpansionSumZeroElim(temp64clen, &temp64c[0], temp128len, &temp128[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(temp192len, &temp192[0], cex, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], cex, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp192len, &temp192[0], cextail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], cex, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], cextail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])
	ylen = ScaleExpansionZeroElim(temp192len, &temp192[0], cey, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], cey, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp192len, &temp192[0], ceytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], cey, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], ceytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])
	zlen = ScaleExpansionZeroElim(temp192len, &temp192[0], cez, &detz[0])
	zzlen = ScaleExpansionZeroElim(zlen, &detz[0], cez, &detzz[0])
	ztlen = ScaleExpansionZeroElim(temp192len, &temp192[0], ceztail, &detzt[0])
	zztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], cez, &detzzt[0])
	for i = 0; i < zztlen; i++ {
		detzzt[i] *= 2.0
	}
	ztztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], ceztail, &detztzt[0])
	z1len = FastExpansionSumZeroElim(zzlen, &detzz[0], zztlen, &detzzt[0], &z1[0])
	z2len = FastExpansionSumZeroElim(z1len, &z1[0], ztztlen, &detztzt[0], &z2[0])
	xylen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &detxy[0])
	clen = FastExpansionSumZeroElim(z2len, &z2[0], xylen, &detxy[0], &cdet[0])

	temp32alen = ScaleExpansionZeroElim(bclen, &bc[0], aez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(bclen, &bc[0], aeztail, &temp32b[0])
	temp64alen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64a[0])
	temp32alen = ScaleExpansionZeroElim(aclen, &ac[0], -bez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(aclen, &ac[0], -beztail, &temp32b[0])
	temp64blen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64b[0])
	temp32alen = ScaleExpansionZeroElim(ablen, &ab[0], cez, &temp32a[0])
	temp32blen = ScaleExpansionZeroElim(ablen, &ab[0], ceztail, &temp32b[0])
	temp64clen = FastExpansionSumZeroElim(temp32alen, &temp32a[0], temp32blen, &temp32b[0], &temp64c[0])
	temp128len = FastExpansionSumZeroElim(temp64alen, &temp64a[0], temp64blen, &temp64b[0], &temp128[0])
	temp192len = FastExpansionSumZeroElim(temp64clen, &temp64c[0], temp128len, &temp128[0], &temp192[0])
	xlen = ScaleExpansionZeroElim(temp192len, &temp192[0], dex, &detx[0])
	xxlen = ScaleExpansionZeroElim(xlen, &detx[0], dex, &detxx[0])
	xtlen = ScaleExpansionZeroElim(temp192len, &temp192[0], dextail, &detxt[0])
	xxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], dex, &detxxt[0])
	for i = 0; i < xxtlen; i++ {
		detxxt[i] *= 2.0
	}
	xtxtlen = ScaleExpansionZeroElim(xtlen, &detxt[0], dextail, &detxtxt[0])
	x1len = FastExpansionSumZeroElim(xxlen, &detxx[0], xxtlen, &detxxt[0], &x1[0])
	x2len = FastExpansionSumZeroElim(x1len, &x1[0], xtxtlen, &detxtxt[0], &x2[0])
	ylen = ScaleExpansionZeroElim(temp192len, &temp192[0], dey, &dety[0])
	yylen = ScaleExpansionZeroElim(ylen, &dety[0], dey, &detyy[0])
	ytlen = ScaleExpansionZeroElim(temp192len, &temp192[0], deytail, &detyt[0])
	yytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], dey, &detyyt[0])
	for i = 0; i < yytlen; i++ {
		detyyt[i] *= 2.0
	}
	ytytlen = ScaleExpansionZeroElim(ytlen, &detyt[0], deytail, &detytyt[0])
	y1len = FastExpansionSumZeroElim(yylen, &detyy[0], yytlen, &detyyt[0], &y1[0])
	y2len = FastExpansionSumZeroElim(y1len, &y1[0], ytytlen, &detytyt[0], &y2[0])
	zlen = ScaleExpansionZeroElim(temp192len, &temp192[0], dez, &detz[0])
	zzlen = ScaleExpansionZeroElim(zlen, &detz[0], dez, &detzz[0])
	ztlen = ScaleExpansionZeroElim(temp192len, &temp192[0], deztail, &detzt[0])
	zztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], dez, &detzzt[0])
	for i = 0; i < zztlen; i++ {
		detzzt[i] *= 2.0
	}
	ztztlen = ScaleExpansionZeroElim(ztlen, &detzt[0], deztail, &detztzt[0])
	z1len = FastExpansionSumZeroElim(zzlen, &detzz[0], zztlen, &detzzt[0], &z1[0])
	z2len = FastExpansionSumZeroElim(z1len, &z1[0], ztztlen, &detztzt[0], &z2[0])
	xylen = FastExpansionSumZeroElim(x2len, &x2[0], y2len, &y2[0], &detxy[0])
	dlen = FastExpansionSumZeroElim(z2len, &z2[0], xylen, &detxy[0], &ddet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	cdlen = FastExpansionSumZeroElim(clen, &cdet[0], dlen, &ddet[0], &cddet[0])
	deterlen = FastExpansionSumZeroElim(ablen, &abdet[0], cdlen, &cddet[0], &deter[0])

	return deter[deterlen-1]
}

// # 3887 "./predicates.c.txt"
func InsphereAdapt(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float, pe [3]Float, permanent Float) Float {
	var aex, bex, cex, dex, aey, bey, cey, dey, aez, bez, cez, dez Float
	var det, errbound Float

	var aexbey1, bexaey1, bexcey1, cexbey1 Float
	var cexdey1, dexcey1, dexaey1, aexdey1 Float
	var aexcey1, cexaey1, bexdey1, dexbey1 Float
	var aexbey0, bexaey0, bexcey0, cexbey0 Float
	var cexdey0, dexcey0, dexaey0, aexdey0 Float
	var aexcey0, cexaey0, bexdey0, dexbey0 Float
	var ab [4]Float
	var bc [4]Float
	var cd [4]Float
	var da [4]Float
	var ac [4]Float
	var bd [4]Float
	var ab3, bc3, cd3, da3, ac3, bd3 Float
	var abeps, bceps, cdeps, daeps, aceps, bdeps Float
	var temp8a [8]Float
	var temp8b [8]Float
	var temp8c [8]Float
	var temp16 [16]Float
	var temp24 [24]Float
	var temp48 [48]Float
	var temp8alen, temp8blen, temp8clen, temp16len, temp24len, temp48len int
	var xdet [96]Float
	var ydet [96]Float
	var zdet [96]Float
	var xydet [192]Float
	var xlen, ylen, zlen, xylen int
	var adet [288]Float
	var bdet [288]Float
	var cdet [288]Float
	var ddet [288]Float
	var alen, blen, clen, dlen int
	var abdet [576]Float
	var cddet [576]Float
	var ablen, cdlen int
	var fin1 [1152]Float
	var finlength int

	var aextail, bextail, cextail, dextail Float
	var aeytail, beytail, ceytail, deytail Float
	var aeztail, beztail, ceztail, deztail Float

	var bvirt Float
	var avirt, bround, around Float
	var c Float
	var abig Float
	var ahi, alo, bhi, blo Float
	var err1, err2, err3 Float
	var _i, _j Float
	var _0 Float

	aex = (Float)(pa[0] - pe[0])
	bex = (Float)(pb[0] - pe[0])
	cex = (Float)(pc[0] - pe[0])
	dex = (Float)(pd[0] - pe[0])
	aey = (Float)(pa[1] - pe[1])
	bey = (Float)(pb[1] - pe[1])
	cey = (Float)(pc[1] - pe[1])
	dey = (Float)(pd[1] - pe[1])
	aez = (Float)(pa[2] - pe[2])
	bez = (Float)(pb[2] - pe[2])
	cez = (Float)(pc[2] - pe[2])
	dez = (Float)(pd[2] - pe[2])

	aexbey1 = (Float)(aex * bey)
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	ahi = c - abig
	alo = aex - ahi
	c = (Float)(splitter * bey)
	abig = (Float)(c - bey)
	bhi = c - abig
	blo = bey - bhi
	err1 = aexbey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	aexbey0 = (alo * blo) - err3
	bexaey1 = (Float)(bex * aey)
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	ahi = c - abig
	alo = bex - ahi
	c = (Float)(splitter * aey)
	abig = (Float)(c - aey)
	bhi = c - abig
	blo = aey - bhi
	err1 = bexaey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bexaey0 = (alo * blo) - err3
	_i = (Float)(aexbey0 - bexaey0)
	bvirt = (Float)(aexbey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bexaey0
	around = aexbey0 - avirt
	ab[0] = around + bround
	_j = (Float)(aexbey1 + _i)
	bvirt = (Float)(_j - aexbey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = aexbey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - bexaey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - bexaey1
	around = _0 - avirt
	ab[1] = around + bround
	ab3 = (Float)(_j + _i)
	bvirt = (Float)(ab3 - _j)
	avirt = ab3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ab[2] = around + bround
	ab[3] = ab3

	bexcey1 = (Float)(bex * cey)
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	ahi = c - abig
	alo = bex - ahi
	c = (Float)(splitter * cey)
	abig = (Float)(c - cey)
	bhi = c - abig
	blo = cey - bhi
	err1 = bexcey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bexcey0 = (alo * blo) - err3
	cexbey1 = (Float)(cex * bey)
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	ahi = c - abig
	alo = cex - ahi
	c = (Float)(splitter * bey)
	abig = (Float)(c - bey)
	bhi = c - abig
	blo = bey - bhi
	err1 = cexbey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cexbey0 = (alo * blo) - err3
	_i = (Float)(bexcey0 - cexbey0)
	bvirt = (Float)(bexcey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cexbey0
	around = bexcey0 - avirt
	bc[0] = around + bround
	_j = (Float)(bexcey1 + _i)
	bvirt = (Float)(_j - bexcey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bexcey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cexbey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cexbey1
	around = _0 - avirt
	bc[1] = around + bround
	bc3 = (Float)(_j + _i)
	bvirt = (Float)(bc3 - _j)
	avirt = bc3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bc[2] = around + bround
	bc[3] = bc3

	cexdey1 = (Float)(cex * dey)
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	ahi = c - abig
	alo = cex - ahi
	c = (Float)(splitter * dey)
	abig = (Float)(c - dey)
	bhi = c - abig
	blo = dey - bhi
	err1 = cexdey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cexdey0 = (alo * blo) - err3
	dexcey1 = (Float)(dex * cey)
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	ahi = c - abig
	alo = dex - ahi
	c = (Float)(splitter * cey)
	abig = (Float)(c - cey)
	bhi = c - abig
	blo = cey - bhi
	err1 = dexcey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dexcey0 = (alo * blo) - err3
	_i = (Float)(cexdey0 - dexcey0)
	bvirt = (Float)(cexdey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dexcey0
	around = cexdey0 - avirt
	cd[0] = around + bround
	_j = (Float)(cexdey1 + _i)
	bvirt = (Float)(_j - cexdey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = cexdey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dexcey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dexcey1
	around = _0 - avirt
	cd[1] = around + bround
	cd3 = (Float)(_j + _i)
	bvirt = (Float)(cd3 - _j)
	avirt = cd3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	cd[2] = around + bround
	cd[3] = cd3

	dexaey1 = (Float)(dex * aey)
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	ahi = c - abig
	alo = dex - ahi
	c = (Float)(splitter * aey)
	abig = (Float)(c - aey)
	bhi = c - abig
	blo = aey - bhi
	err1 = dexaey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dexaey0 = (alo * blo) - err3
	aexdey1 = (Float)(aex * dey)
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	ahi = c - abig
	alo = aex - ahi
	c = (Float)(splitter * dey)
	abig = (Float)(c - dey)
	bhi = c - abig
	blo = dey - bhi
	err1 = aexdey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	aexdey0 = (alo * blo) - err3
	_i = (Float)(dexaey0 - aexdey0)
	bvirt = (Float)(dexaey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - aexdey0
	around = dexaey0 - avirt
	da[0] = around + bround
	_j = (Float)(dexaey1 + _i)
	bvirt = (Float)(_j - dexaey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = dexaey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - aexdey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - aexdey1
	around = _0 - avirt
	da[1] = around + bround
	da3 = (Float)(_j + _i)
	bvirt = (Float)(da3 - _j)
	avirt = da3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	da[2] = around + bround
	da[3] = da3

	aexcey1 = (Float)(aex * cey)
	c = (Float)(splitter * aex)
	abig = (Float)(c - aex)
	ahi = c - abig
	alo = aex - ahi
	c = (Float)(splitter * cey)
	abig = (Float)(c - cey)
	bhi = c - abig
	blo = cey - bhi
	err1 = aexcey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	aexcey0 = (alo * blo) - err3
	cexaey1 = (Float)(cex * aey)
	c = (Float)(splitter * cex)
	abig = (Float)(c - cex)
	ahi = c - abig
	alo = cex - ahi
	c = (Float)(splitter * aey)
	abig = (Float)(c - aey)
	bhi = c - abig
	blo = aey - bhi
	err1 = cexaey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	cexaey0 = (alo * blo) - err3
	_i = (Float)(aexcey0 - cexaey0)
	bvirt = (Float)(aexcey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cexaey0
	around = aexcey0 - avirt
	ac[0] = around + bround
	_j = (Float)(aexcey1 + _i)
	bvirt = (Float)(_j - aexcey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = aexcey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - cexaey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - cexaey1
	around = _0 - avirt
	ac[1] = around + bround
	ac3 = (Float)(_j + _i)
	bvirt = (Float)(ac3 - _j)
	avirt = ac3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	ac[2] = around + bround
	ac[3] = ac3

	bexdey1 = (Float)(bex * dey)
	c = (Float)(splitter * bex)
	abig = (Float)(c - bex)
	ahi = c - abig
	alo = bex - ahi
	c = (Float)(splitter * dey)
	abig = (Float)(c - dey)
	bhi = c - abig
	blo = dey - bhi
	err1 = bexdey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	bexdey0 = (alo * blo) - err3
	dexbey1 = (Float)(dex * bey)
	c = (Float)(splitter * dex)
	abig = (Float)(c - dex)
	ahi = c - abig
	alo = dex - ahi
	c = (Float)(splitter * bey)
	abig = (Float)(c - bey)
	bhi = c - abig
	blo = bey - bhi
	err1 = dexbey1 - (ahi * bhi)
	err2 = err1 - (alo * bhi)
	err3 = err2 - (ahi * blo)
	dexbey0 = (alo * blo) - err3
	_i = (Float)(bexdey0 - dexbey0)
	bvirt = (Float)(bexdey0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dexbey0
	around = bexdey0 - avirt
	bd[0] = around + bround
	_j = (Float)(bexdey1 + _i)
	bvirt = (Float)(_j - bexdey1)
	avirt = _j - bvirt
	bround = _i - bvirt
	around = bexdey1 - avirt
	_0 = around + bround
	_i = (Float)(_0 - dexbey1)
	bvirt = (Float)(_0 - _i)
	avirt = _i + bvirt
	bround = bvirt - dexbey1
	around = _0 - avirt
	bd[1] = around + bround
	bd3 = (Float)(_j + _i)
	bvirt = (Float)(bd3 - _j)
	avirt = bd3 - bvirt
	bround = _i - bvirt
	around = _j - avirt
	bd[2] = around + bround
	bd[3] = bd3

	temp8alen = ScaleExpansionZeroElim(4, &cd[0], bez, &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &bd[0], -cez, &temp8b[0])
	temp8clen = ScaleExpansionZeroElim(4, &bc[0], dez, &temp8c[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp24len = FastExpansionSumZeroElim(temp8clen, &temp8c[0], temp16len, &temp16[0], &temp24[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], aex, &temp48[0])
	xlen = ScaleExpansionZeroElim(temp48len, &temp48[0], -aex, &xdet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], aey, &temp48[0])
	ylen = ScaleExpansionZeroElim(temp48len, &temp48[0], -aey, &ydet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], aez, &temp48[0])
	zlen = ScaleExpansionZeroElim(temp48len, &temp48[0], -aez, &zdet[0])
	xylen = FastExpansionSumZeroElim(xlen, &xdet[0], ylen, &ydet[0], &xydet[0])
	alen = FastExpansionSumZeroElim(xylen, &xydet[0], zlen, &zdet[0], &adet[0])

	temp8alen = ScaleExpansionZeroElim(4, &da[0], cez, &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ac[0], dez, &temp8b[0])
	temp8clen = ScaleExpansionZeroElim(4, &cd[0], aez, &temp8c[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp24len = FastExpansionSumZeroElim(temp8clen, &temp8c[0], temp16len, &temp16[0], &temp24[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], bex, &temp48[0])
	xlen = ScaleExpansionZeroElim(temp48len, &temp48[0], bex, &xdet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], bey, &temp48[0])
	ylen = ScaleExpansionZeroElim(temp48len, &temp48[0], bey, &ydet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], bez, &temp48[0])
	zlen = ScaleExpansionZeroElim(temp48len, &temp48[0], bez, &zdet[0])
	xylen = FastExpansionSumZeroElim(xlen, &xdet[0], ylen, &ydet[0], &xydet[0])
	blen = FastExpansionSumZeroElim(xylen, &xydet[0], zlen, &zdet[0], &bdet[0])

	temp8alen = ScaleExpansionZeroElim(4, &ab[0], dez, &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &bd[0], aez, &temp8b[0])
	temp8clen = ScaleExpansionZeroElim(4, &da[0], bez, &temp8c[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp24len = FastExpansionSumZeroElim(temp8clen, &temp8c[0], temp16len, &temp16[0], &temp24[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], cex, &temp48[0])
	xlen = ScaleExpansionZeroElim(temp48len, &temp48[0], -cex, &xdet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], cey, &temp48[0])
	ylen = ScaleExpansionZeroElim(temp48len, &temp48[0], -cey, &ydet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], cez, &temp48[0])
	zlen = ScaleExpansionZeroElim(temp48len, &temp48[0], -cez, &zdet[0])
	xylen = FastExpansionSumZeroElim(xlen, &xdet[0], ylen, &ydet[0], &xydet[0])
	clen = FastExpansionSumZeroElim(xylen, &xydet[0], zlen, &zdet[0], &cdet[0])

	temp8alen = ScaleExpansionZeroElim(4, &bc[0], aez, &temp8a[0])
	temp8blen = ScaleExpansionZeroElim(4, &ac[0], -bez, &temp8b[0])
	temp8clen = ScaleExpansionZeroElim(4, &ab[0], cez, &temp8c[0])
	temp16len = FastExpansionSumZeroElim(temp8alen, &temp8a[0], temp8blen, &temp8b[0], &temp16[0])
	temp24len = FastExpansionSumZeroElim(temp8clen, &temp8c[0], temp16len, &temp16[0], &temp24[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], dex, &temp48[0])
	xlen = ScaleExpansionZeroElim(temp48len, &temp48[0], dex, &xdet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], dey, &temp48[0])
	ylen = ScaleExpansionZeroElim(temp48len, &temp48[0], dey, &ydet[0])
	temp48len = ScaleExpansionZeroElim(temp24len, &temp24[0], dez, &temp48[0])
	zlen = ScaleExpansionZeroElim(temp48len, &temp48[0], dez, &zdet[0])
	xylen = FastExpansionSumZeroElim(xlen, &xdet[0], ylen, &ydet[0], &xydet[0])
	dlen = FastExpansionSumZeroElim(xylen, &xydet[0], zlen, &zdet[0], &ddet[0])

	ablen = FastExpansionSumZeroElim(alen, &adet[0], blen, &bdet[0], &abdet[0])
	cdlen = FastExpansionSumZeroElim(clen, &cdet[0], dlen, &ddet[0], &cddet[0])
	finlength = FastExpansionSumZeroElim(ablen, &abdet[0], cdlen, &cddet[0], &fin1[0])

	det = Estimate(finlength, &fin1[0])
	errbound = isperrboundB * permanent
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	bvirt = (Float)(pa[0] - aex)
	avirt = aex + bvirt
	bround = bvirt - pe[0]
	around = pa[0] - avirt
	aextail = around + bround
	bvirt = (Float)(pa[1] - aey)
	avirt = aey + bvirt
	bround = bvirt - pe[1]
	around = pa[1] - avirt
	aeytail = around + bround
	bvirt = (Float)(pa[2] - aez)
	avirt = aez + bvirt
	bround = bvirt - pe[2]
	around = pa[2] - avirt
	aeztail = around + bround
	bvirt = (Float)(pb[0] - bex)
	avirt = bex + bvirt
	bround = bvirt - pe[0]
	around = pb[0] - avirt
	bextail = around + bround
	bvirt = (Float)(pb[1] - bey)
	avirt = bey + bvirt
	bround = bvirt - pe[1]
	around = pb[1] - avirt
	beytail = around + bround
	bvirt = (Float)(pb[2] - bez)
	avirt = bez + bvirt
	bround = bvirt - pe[2]
	around = pb[2] - avirt
	beztail = around + bround
	bvirt = (Float)(pc[0] - cex)
	avirt = cex + bvirt
	bround = bvirt - pe[0]
	around = pc[0] - avirt
	cextail = around + bround
	bvirt = (Float)(pc[1] - cey)
	avirt = cey + bvirt
	bround = bvirt - pe[1]
	around = pc[1] - avirt
	ceytail = around + bround
	bvirt = (Float)(pc[2] - cez)
	avirt = cez + bvirt
	bround = bvirt - pe[2]
	around = pc[2] - avirt
	ceztail = around + bround
	bvirt = (Float)(pd[0] - dex)
	avirt = dex + bvirt
	bround = bvirt - pe[0]
	around = pd[0] - avirt
	dextail = around + bround
	bvirt = (Float)(pd[1] - dey)
	avirt = dey + bvirt
	bround = bvirt - pe[1]
	around = pd[1] - avirt
	deytail = around + bround
	bvirt = (Float)(pd[2] - dez)
	avirt = dez + bvirt
	bround = bvirt - pe[2]
	around = pd[2] - avirt
	deztail = around + bround
	if (aextail == 0.0) && (aeytail == 0.0) && (aeztail == 0.0) &&
		(bextail == 0.0) && (beytail == 0.0) && (beztail == 0.0) &&
		(cextail == 0.0) && (ceytail == 0.0) && (ceztail == 0.0) &&
		(dextail == 0.0) && (deytail == 0.0) && (deztail == 0.0) {
		return det
	}

	errbound = isperrboundC*permanent + resulterrbound*abs(det)
	abeps = (aex*beytail + bey*aextail) -
		(aey*bextail + bex*aeytail)
	bceps = (bex*ceytail + cey*bextail) -
		(bey*cextail + cex*beytail)
	cdeps = (cex*deytail + dey*cextail) -
		(cey*dextail + dex*ceytail)
	daeps = (dex*aeytail + aey*dextail) -
		(dey*aextail + aex*deytail)
	aceps = (aex*ceytail + cey*aextail) -
		(aey*cextail + cex*aeytail)
	bdeps = (bex*deytail + dey*bextail) -
		(bey*dextail + dex*beytail)
	det += (((bex*bex+bey*bey+bez*bez)*
		((cez*daeps+dez*aceps+aez*cdeps)+
			(ceztail*da3+deztail*ac3+aeztail*cd3)) +
		(dex*dex+dey*dey+dez*dez)*
			((aez*bceps-bez*aceps+cez*abeps)+
				(aeztail*bc3-beztail*ac3+ceztail*ab3))) -
		((aex*aex+aey*aey+aez*aez)*
			((bez*cdeps-cez*bdeps+dez*bceps)+
				(beztail*cd3-ceztail*bd3+deztail*bc3)) +
			(cex*cex+cey*cey+cez*cez)*
				((dez*abeps+aez*bdeps+bez*daeps)+
					(deztail*ab3+aeztail*bd3+beztail*da3)))) +
		2.0*(((bex*bextail+bey*beytail+bez*beztail)*
			(cez*da3+dez*ac3+aez*cd3)+
			(dex*dextail+dey*deytail+dez*deztail)*
				(aez*bc3-bez*ac3+cez*ab3))-
			((aex*aextail+aey*aeytail+aez*aeztail)*
				(bez*cd3-cez*bd3+dez*bc3)+
				(cex*cextail+cey*ceytail+cez*ceztail)*
					(dez*ab3+aez*bd3+bez*da3)))
	if (det >= errbound) || (-det >= errbound) {
		return det
	}

	return InsphereExact(pa, pb, pc, pd, pe)
}

/*****************************************************************************/
/*                                                                           */
/*  inspherefast()   Approximate 3D insphere test.  Nonrobust.               */
/*  insphereexact()   Exact 3D insphere test.  Robust.                       */
/*  insphereslow()   Another exact 3D insphere test.  Robust.                */
/*  insphere()   Adaptive exact 3D insphere test.  Robust.                   */
/*                                                                           */
/*               Return a positive value if the point pe lies inside the     */
/*               sphere passing through pa, pb, pc, and pd; a negative value */
/*               if it lies outside; and zero if the five points are         */
/*               cospherical.  The points pa, pb, pc, and pd must be ordered */
/*               so that they have a positive orientation (as defined by     */
/*               orient3d()), or the sign of the result will be reversed.    */
/*                                                                           */
/*  Only the first and last routine should be used; the middle two are for   */
/*  timings.                                                                 */
/*                                                                           */
/*  The last three use exact arithmetic to ensure a correct answer.  The     */
/*  result returned is the determinant of a matrix.  In insphere() only,     */
/*  this determinant is computed adaptively, in the sense that exact         */
/*  arithmetic is used only to the degree it is needed to ensure that the    */
/*  returned value has the correct sign.  Hence, insphere() is usually quite */
/*  fast, but will run more slowly when the input points are cospherical or  */
/*  nearly so.                                                               */
/*                                                                           */
/*****************************************************************************/
func Insphere(pa [3]Float, pb [3]Float, pc [3]Float, pd [3]Float, pe [3]Float) Float {
	var aex, bex, cex, dex Float
	var aey, bey, cey, dey Float
	var aez, bez, cez, dez Float
	var aexbey, bexaey, bexcey, cexbey, cexdey, dexcey, dexaey, aexdey Float
	var aexcey, cexaey, bexdey, dexbey Float
	var alift, blift, clift, dlift Float
	var ab, bc, cd, da, ac, bd Float
	var abc, bcd, cda, dab Float
	var aezplus, bezplus, cezplus, dezplus Float
	var aexbeyplus, bexaeyplus, bexceyplus, cexbeyplus Float
	var cexdeyplus, dexceyplus, dexaeyplus, aexdeyplus Float
	var aexceyplus, cexaeyplus, bexdeyplus, dexbeyplus Float
	var det Float
	var permanent, errbound Float

	aex = pa[0] - pe[0]
	bex = pb[0] - pe[0]
	cex = pc[0] - pe[0]
	dex = pd[0] - pe[0]
	aey = pa[1] - pe[1]
	bey = pb[1] - pe[1]
	cey = pc[1] - pe[1]
	dey = pd[1] - pe[1]
	aez = pa[2] - pe[2]
	bez = pb[2] - pe[2]
	cez = pc[2] - pe[2]
	dez = pd[2] - pe[2]

	aexbey = aex * bey
	bexaey = bex * aey
	ab = aexbey - bexaey
	bexcey = bex * cey
	cexbey = cex * bey
	bc = bexcey - cexbey
	cexdey = cex * dey
	dexcey = dex * cey
	cd = cexdey - dexcey
	dexaey = dex * aey
	aexdey = aex * dey
	da = dexaey - aexdey

	aexcey = aex * cey
	cexaey = cex * aey
	ac = aexcey - cexaey
	bexdey = bex * dey
	dexbey = dex * bey
	bd = bexdey - dexbey

	abc = aez*bc - bez*ac + cez*ab
	bcd = bez*cd - cez*bd + dez*bc
	cda = cez*da + dez*ac + aez*cd
	dab = dez*ab + aez*bd + bez*da

	alift = aex*aex + aey*aey + aez*aez
	blift = bex*bex + bey*bey + bez*bez
	clift = cex*cex + cey*cey + cez*cez
	dlift = dex*dex + dey*dey + dez*dez

	det = (dlift*abc - clift*dab) + (blift*cda - alift*bcd)

	aezplus = abs(aez)
	bezplus = abs(bez)
	cezplus = abs(cez)
	dezplus = abs(dez)
	aexbeyplus = abs(aexbey)
	bexaeyplus = abs(bexaey)
	bexceyplus = abs(bexcey)
	cexbeyplus = abs(cexbey)
	cexdeyplus = abs(cexdey)
	dexceyplus = abs(dexcey)
	dexaeyplus = abs(dexaey)
	aexdeyplus = abs(aexdey)
	aexceyplus = abs(aexcey)
	cexaeyplus = abs(cexaey)
	bexdeyplus = abs(bexdey)
	dexbeyplus = abs(dexbey)
	permanent = ((cexdeyplus+dexceyplus)*bezplus+
		(dexbeyplus+bexdeyplus)*cezplus+
		(bexceyplus+cexbeyplus)*dezplus)*
		alift +
		((dexaeyplus+aexdeyplus)*cezplus+
			(aexceyplus+cexaeyplus)*dezplus+
			(cexdeyplus+dexceyplus)*aezplus)*
			blift +
		((aexbeyplus+bexaeyplus)*dezplus+
			(bexdeyplus+dexbeyplus)*aezplus+
			(dexaeyplus+aexdeyplus)*bezplus)*
			clift +
		((bexceyplus+cexbeyplus)*aezplus+
			(cexaeyplus+aexceyplus)*bezplus+
			(aexbeyplus+bexaeyplus)*cezplus)*
			dlift
	errbound = isperrboundA * permanent
	if (det > errbound) || (-det > errbound) {
		return det
	}

	return InsphereAdapt(pa, pb, pc, pd, pe, permanent)
}

func Incircle2pFast(pa, pb, pc [2]Float) Float {
	// 算法原理: 判断∠acb是否超过90度
	acx := pa[0] - pc[0]
	bcx := pb[0] - pc[0]
	acy := pa[1] - pc[1]
	bcy := pb[1] - pc[1]

	return -acx*bcx - acy*bcy
}

func Incircle2p(pa, pb, pc [2]Float) Float {
	// return Incircle(pa, pa, pb, pc) // it doesn't works!
	det := Incircle2pFast(pa, pb, pc)

	// 这个算法也是浮点运算, 那么它是否和Orient2d一样, 有精度问题呢? 我不知道.
	// 这里暂且测试一下它的结果和双精度运算是否一致, 等遇到不一致的情况, 再研究如何改进.
	// acx := float64(pa[0]) - float64(pc[0])
	// bcx := float64(pb[0]) - float64(pc[0])
	// acy := float64(pa[1]) - float64(pc[1])
	// bcy := float64(pb[1]) - float64(pc[1])
	// det1 := -acx*bcx - acy*bcy
	// debug.Assert(det1 == 0 && det == 0 || det1 < 0 && det < 0 || det1 > 0 && det > 0)

	return det
}
