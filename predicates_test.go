package predicates

import (
	"testing"
	"unsafe"
)

func TestOrient2d(t *testing.T) {
	// 这个测试用例, 线段ab特别短, 点c在特别远之外, 这样
	fastFailed := 0
	numTest := 0
	dy := (Float)(0.0009999)
	var m Float
	if unsafe.Sizeof(m) == 8 {
		m = 10000
	} else {
		m = 1
	}
	pa := [2]Float{0, 0}
	pb := [2]Float{2, 3}
	pc := [2]Float{20000 * m, 30000 * m}
	for pc[1] = 30000*m - 1; pc[1] < 30000*m+1; pc[1] += dy {
		numTest++
		te := Orient2dExact(pa, pb, pc)
		ts := Orient2dSlow(pa, pb, pc)
		tn := Orient2d(pa, pb, pc)
		tf := Orient2dFast(pa, pb, pc)

		if !isSamePred(te, ts) || !isSamePred(te, tn) {
			t.Errorf("Orient2dExact()=%v, Orient2dSlow()=%v, Orient2d()=%v, Orient2dFast()=%v, pa=%v, pb=%v, pc=%v", te, ts, tn, tf, pa, pb, pc)
		}
		if !isSamePred(te, tf) {
			// t.Logf("%v", pc)
			fastFailed++
		}
	}
	// t.Logf("Orient2d: numTest=%d, fastFailed=%d, cover=%f", numTest, fastFailed, float32(fastFailed)/float32(numTest))
	if float32(fastFailed)/float32(numTest) < 0.1 {
		t.Errorf("this testcase should be improve")
	}
}

func TestOrient2dRand(t *testing.T) {
	for i := 0; i < 100000; i++ {
		pa := [2]Float{narrowRealRand(), narrowRealRand()}
		pb := [2]Float{narrowRealRand(), narrowRealRand()}
		pc := [2]Float{narrowRealRand(), narrowRealRand()}

		te := Orient2dExact(pa, pb, pc)
		ts := Orient2dSlow(pa, pb, pc)
		tn := Orient2d(pa, pb, pc)
		tf := Orient2dFast(pa, pb, pc)
		if !isSamePred(te, ts) || !isSamePred(te, tn) {
			t.Errorf("Orient2dExact()=%v, Orient2dSlow()=%v, Orient2d()=%v, Orient2dFast()=%v, pa=%v, pb=%v, pc=%v", te, ts, tn, tf, pa, pb, pc)
		}

	}
}

func TestOrientSign(t *testing.T) {
	// a,b,c 逆时针排列, predicates里的注释, 它应该返回正值
	pa, pb, pc := [2]Float{0, 0}, [2]Float{1, 0}, [2]Float{0, 1}
	if Orient2d(pa, pb, pc) <= 0 {
		t.Errorf("Orient2d() sign error")
	}
}

func TestIncircle(t *testing.T) {
	type args struct {
		pa [2]Float
		pb [2]Float
		pc [2]Float
		pd [2]Float
	}
	tests := []struct {
		name string
		args args
		want Float
	}{
		{
			name: "on circle 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0, 1}, pd: [2]Float{0, 0}},
			want: 0,
		},
		{
			name: "on circle 2",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0, 1}, pd: [2]Float{1, 1}},
			want: 0,
		},
		{
			name: "outer 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0, 1}, pd: [2]Float{1.1, 1.1}},
			want: -1,
		},
		{
			name: "inner 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0, 1}, pd: [2]Float{0.5, 0.5}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Incircle(tt.args.pa, tt.args.pb, tt.args.pc, tt.args.pd); !isSamePred(got, tt.want) {
				t.Errorf("Incircle() = %v, want sign=%v", got, tt.want)
			}
		})
	}
}

func TestIncircle2p(t *testing.T) {
	type args struct {
		pa [2]Float
		pb [2]Float
		pc [2]Float
	}
	tests := []struct {
		name string
		args args
		want Float
	}{
		{
			name: "on circle 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0, 0}},
			want: 0,
		},
		{
			name: "on circle 2",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{1, 0}},
			want: 0,
		},
		{
			name: "outer 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0.5, 1}},
			want: -1,
		},
		{
			name: "inner 1",
			args: args{pa: [2]Float{0, 0}, pb: [2]Float{1, 0}, pc: [2]Float{0.5, 0.1}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Incircle2p(tt.args.pa, tt.args.pb, tt.args.pc); !isSamePred(got, tt.want) {
				t.Errorf("Incircle2p() = %v, want sign=%v", got, tt.want)
			}
		})
	}
}

func TestIncircle2pRand(t *testing.T) {
	for i := 0; i < 10000; i++ {
		pa := [2]Float{narrowRealRand(), narrowRealRand()}
		pb := [2]Float{narrowRealRand(), narrowRealRand()}
		pc := [2]Float{narrowRealRand(), narrowRealRand()}

		tn := Incircle2p(pa, pb, pc)
		tf := Incircle2pFast(pa, pb, pc)
		if !isSamePred(tn, tf) {
			t.Errorf("Incircle2p()=%v, Incircle2pFast()=%v, pa=%v, pb=%v, pc=%v", tn, tf, pa, pb, pc)
		}

	}
}
