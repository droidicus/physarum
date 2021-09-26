package physarum

import (
	"testing"
)

func TestBoxBlurH(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		boxBlurH(src, dst1, w, h, r, 1)
		slowBoxBlurH(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}

func TestBoxBlurV(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		boxBlurV(src, dst1, w, h, r, 1)
		slowBoxBlurV(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}

// Make sure the optimizer doesn't optimize away the functions!
var result []float32

func BenchmarkBoxBlurH(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		boxBlurH(src, dst, w, h, 1, 1)
	}
	result = dst
}

func BenchmarkBoxBlurV(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		boxBlurV(src, dst, w, h, 1, 1)
	}
	result = dst
}

func BenchmarkSlowBoxBlurH(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		slowBoxBlurH(src, dst, w, h, 1, 1)
	}
	result = dst
}

func BenchmarkSlowBoxBlurV(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		slowBoxBlurV(src, dst, w, h, 1, 1)
	}
	result = dst
}

func BenchmarkSlowThreadedBoxBlurH(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		slowThreadedBoxBlurH(src, dst, w, h, 1, 1)
	}
	result = dst
}

func BenchmarkSlowThreadedBoxBlurV(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		slowThreadedBoxBlurV(src, dst, w, h, 1, 1)
	}
	result = dst
}

func TestSlowThreadedBoxBlurH(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		slowBoxBlurH(src, dst1, w, h, r, 1)
		slowThreadedBoxBlurH(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}

func TestSlowThreadedBoxBlurV(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		slowThreadedBoxBlurV(src, dst1, w, h, r, 1)
		slowBoxBlurV(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}

func BenchmarkThreadedBoxBlurH(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		threadedBoxBlurH(src, dst, w, h, 1, 1)
	}
	result = dst
}

func TestThreadedBoxBlurH(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		threadedBoxBlurH(src, dst1, w, h, r, 1)
		slowBoxBlurH(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}

func BenchmarkThreadedBoxBlurV(b *testing.B) {
	// Setup
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst := make([]float32, w*h)

	// Init with some data
	for i := range src {
		src[i] = float32(i)
	}

	// Run the blur benchmark
	for i := 0; i < b.N; i++ {
		threadedBoxBlurV(src, dst, w, h, 1, 1)
	}
	result = dst
}

func TestThreadedBoxBlurV(t *testing.T) {
	w := 1024
	h := 1024
	src := make([]float32, w*h)
	dst1 := make([]float32, w*h)
	dst2 := make([]float32, w*h)
	for i := range src {
		src[i] = float32(i)
	}
	for r := 0; r < 5; r++ {
		threadedBoxBlurV(src, dst1, w, h, r, 1)
		slowBoxBlurV(src, dst2, w, h, r, 1)
		for i := range src {
			if dst1[i] != dst2[i] {
				t.Fatalf("got %v, want %v", dst1, dst2)
			}
		}
	}
}
