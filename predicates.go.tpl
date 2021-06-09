package calc

#define USE_DOUBLE

#ifdef USE_DOUBLE 
# define REAL float64 
# define SUFFIX F64 
#else
# define REAL float32 
# define SUFFIX F64 
#endif

#define FUNC_NAME_1(Name, Suffix) Name ## Suffix
#define FUNC_NAME_2(Name, Suffix) FUNC_NAME_1(Name, Suffix)
#define FUNC_NAME(Name) FUNC_NAME_2(Name, SUFFIX)



import (
	"fmt"
	"math"
	"math/rand"
	"unsafe"

	"github.com/toy80/toy/coord"
	"github.com/toy80/toy/utils/debug"
)

const floatSize = coord.Size

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

func DoubleToString(number float64) (s string) {
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

func FloatToString(number float32) (s string) {
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

func RealToString(x Float) string {
	if unsafe.Sizeof(x) == 4 {
		return FloatToString(float32(x))
	} else {
		return DoubleToString(float64(x))
	}
}

func ExpansionToString(elen int, e *Float) (s string) {
	for i := elen - 1; i >= 0; i-- {
		s += RealToString(*(*Float)(unsafe.Pointer(uintptr(unsafe.Pointer(e)) + floatSize*uintptr(i))))
		if i > 0 {
			s += " +\n"
		} else {
			s += "\n"
		}
	}
	return
}

func NarrowRealRand() (x Float) {
	if unsafe.Sizeof(x) == 8 {
		return Float(NarrowDoubleRand())
	}
	return Float(NarrowFloatRand())
}

func RealRand() (x Float) {
	if unsafe.Sizeof(x) == 8 {
		return Float(DoubleRand())
	}
	return Float(FloatRand())
}

// # 567 "./predicates.c.txt"
func DoubleRand() float64 {
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
func NarrowDoubleRand() float64 {
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

func UniformDoubleRand() float64 {
	var result float64
	var a, b int32

	a = random()
	b = random()
	result = (float64)(a-1073741824)*8388608.0 + (float64)(b>>8)
	return result
}

// # 636 "./predicates.c.txt"
func FloatRand() float32 {
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
func NarrowFloatRand() float32 {
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

func UniformFloatRand() float32 {
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
	te = orient2dExact(pa, pb, pc)
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
func fast_expansion_sum_zeroelim(elen int, e *Float, flen int, f *Float, h *Float) int {
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
func scale_expansion_zeroelim(elen int, e *Float, b Float, h *Float) int {
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
func estimate(elen int, e *Float) Float {
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
