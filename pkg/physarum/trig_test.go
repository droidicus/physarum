package physarum

import (
	"fmt"
	"math"
	"testing"
)

func TestTrigLookupTables(t *testing.T) {
	const N = 10000000
	const tolerance = 1e-4
	var maxCosError, maxSinError float64
	for i := 0; i <= N; i++ {
		p := float32(i) / N
		a := (p*20 - 10) * math.Pi
		cosError := math.Abs(float64(cos(a)) - math.Cos(float64(a)))
		sinError := math.Abs(float64(sin(a)) - math.Sin(float64(a)))
		maxCosError = math.Max(maxCosError, cosError)
		maxSinError = math.Max(maxSinError, sinError)
		if cosError > tolerance {
			t.Fatalf("cos(%v) = %v, math.Cos(%v) = %v (%g)", a, cos(a), a, math.Cos(float64(a)), cosError)
		}
		if sinError > tolerance {
			t.Fatalf("sin(%v) = %v, math.Sin(%v) = %v (%g)", a, sin(a), a, math.Sin(float64(a)), sinError)
		}
	}
	fmt.Println("Max cos error: ", maxCosError)
	fmt.Println("Max sin error: ", maxSinError)
}

var result_float float32

func BenchmarkCos64(b *testing.B) {
	var r float32

	// Run the cos benchmark
	for i := 0; i < b.N; i++ {
		r = float32(math.Cos(float64(float32(i))))
	}

	result_float = r
}

func BenchmarkCustonCos32(b *testing.B) {
	var r float32

	// Run the cos benchmark
	for i := 0; i < b.N; i++ {
		r = cos(float32(i))
	}

	result_float = r
}

func BenchmarkSin64(b *testing.B) {
	var r float32

	// Run the cos benchmark
	for i := 0; i < b.N; i++ {
		r = float32(math.Cos(float64(float32(i))))
	}

	result_float = r
}

func BenchmarkCustonSin32(b *testing.B) {
	var r float32

	// Run the cos benchmark
	for i := 0; i < b.N; i++ {
		r = cos(float32(i))
	}

	result_float = r
}
