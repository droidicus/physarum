package physarum

import "sync"

func slowBoxBlurH(src, dst []float32, w, h, r int, scale float32) {
	m := scale / float32(r+r+1)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var val float32
			for k := -r; k <= r; k++ {
				i := y*w + (x+w+k)%w
				val += src[i]
			}
			dst[y*w+x] = val * m
		}
	}
}

func slowBoxBlurV(src, dst []float32, w, h, r int, scale float32) {
	m := scale / float32(r+r+1)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			var val float32
			for k := -r; k <= r; k++ {
				i := x + ((y+h+k)%h)*w
				val += src[i]
			}
			dst[y*w+x] = val * m
		}
	}
}

func slowThreadedBoxBlurH(src, dst []float32, w, h, r int, scale float32) {
	// waitgroup for threads
	var wg sync.WaitGroup

	m := scale / float32(r+r+1)
	for y := 0; y < h; y++ {
		// New thread to wait on
		wg.Add(1)

		go func(y int) {
			// Defer
			defer wg.Done()

			// Do parallel loops
			for x := 0; x < w; x++ {
				var val float32
				for k := -r; k <= r; k++ {
					i := y*w + (x+w+k)%w
					val += src[i]
				}
				dst[y*w+x] = val * m
			}
		}(y)
	}

	// Wait for the threads to finish
	wg.Wait()
}

func slowThreadedBoxBlurV(src, dst []float32, w, h, r int, scale float32) {
	// waitgroup for threads
	var wg sync.WaitGroup

	m := scale / float32(r+r+1)
	for x := 0; x < w; x++ {
		// New thread to wait on
		wg.Add(1)

		go func(x int) {
			// Defer
			defer wg.Done()

			// Do parallel loops
			for y := 0; y < h; y++ {
				var val float32
				for k := -r; k <= r; k++ {
					i := x + ((y+h+k)%h)*w
					val += src[i]
				}
				dst[y*w+x] = val * m
			}
		}(x)
	}

	// Wait for the threads to finish
	wg.Wait()
}

// droid: I don't understand these blur functions...
func boxBlurH(src, dst []float32, w, h, r int, scale float32) {
	m := scale / float32(r+r+1)
	ww := w - (r*2 + 1)
	for i := 0; i < h; i++ {
		ti := i * w
		li := ti + w - 1 - r
		ri := ti + r
		val := src[li]
		for j := 0; j < r; j++ {
			val += src[li+j+1]
			val += src[ti+j]
		}
		for j := 0; j <= r; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li++
			ri++
			ti++
		}
		li = i * w
		for j := 0; j < ww; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li++
			ri++
			ti++
		}
		ri = i * w
		for j := 0; j < r; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li++
			ri++
			ti++
		}
	}
}

// droid: I don't understand these blur functions...
func boxBlurV(src, dst []float32, w, h, r int, scale float32) {
	m := scale / float32(r+r+1)
	hh := h - (r*2 + 1)
	for i := 0; i < w; i++ {
		ti := i
		li := ti + (h-1-r)*w
		ri := ti + r*w
		val := src[li]
		for j := 0; j < r; j++ {
			val += src[li+(j+1)*w]
			val += src[ti+j*w]
		}
		for j := 0; j <= r; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li += w
			ri += w
			ti += w
		}
		li = i
		for j := 0; j < hh; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li += w
			ri += w
			ti += w
		}
		ri = i
		for j := 0; j < r; j++ {
			val += src[ri] - src[li]
			dst[ti] = val * m
			li += w
			ri += w
			ti += w
		}
	}
}

func threadedBoxBlurH(src, dst []float32, w, h, r int, scale float32) {
	// waitgroup for threads
	var wg sync.WaitGroup

	m := scale / float32(r+r+1)
	ww := w - (r*2 + 1)
	for i := 0; i < h; i++ {
		// New thread to wait on
		wg.Add(1)

		go func(i int) {
			// Defer
			defer wg.Done()

			// Do parallel loops
			ti := i * w
			li := ti + w - 1 - r
			ri := ti + r
			val := src[li]
			for j := 0; j < r; j++ {
				val += src[li+j+1]
				val += src[ti+j]
			}
			for j := 0; j <= r; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li++
				ri++
				ti++
			}
			li = i * w
			for j := 0; j < ww; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li++
				ri++
				ti++
			}
			ri = i * w
			for j := 0; j < r; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li++
				ri++
				ti++
			}
		}(i)
	}

	// Wait for the threads to finish
	wg.Wait()
}

func threadedBoxBlurV(src, dst []float32, w, h, r int, scale float32) {
	// waitgroup for threads
	var wg sync.WaitGroup

	m := scale / float32(r+r+1)
	hh := h - (r*2 + 1)
	for i := 0; i < w; i++ {
		// New thread to wait on
		wg.Add(1)

		go func(i int) {
			// Defer
			defer wg.Done()

			// Do parallel loops
			ti := i
			li := ti + (h-1-r)*w
			ri := ti + r*w
			val := src[li]
			for j := 0; j < r; j++ {
				val += src[li+(j+1)*w]
				val += src[ti+j*w]
			}
			for j := 0; j <= r; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li += w
				ri += w
				ti += w
			}
			li = i
			for j := 0; j < hh; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li += w
				ri += w
				ti += w
			}
			ri = i
			for j := 0; j < r; j++ {
				val += src[ri] - src[li]
				dst[ti] = val * m
				li += w
				ri += w
				ti += w
			}
		}(i)
	}

	// Wait for the threads to finish
	wg.Wait()
}

func boxBlur(src, tmp []float32, w, h, r int, scale float32) {
	// TODO: Are these the same or different? If different, add to settings
	// boxBlurH(src, tmp, w, h, r, 1)
	// boxBlurV(tmp, src, w, h, r, scale)

	threadedBoxBlurH(src, tmp, w, h, r, 1)
	threadedBoxBlurV(tmp, src, w, h, r, scale)

	// slowBoxBlurH(src, tmp, w, h, r, 1)
	// slowBoxBlurV(tmp, src, w, h, r, scale)

	// slowThreadedBoxBlurH(src, tmp, w, h, r, 1)
	// slowThreadedBoxBlurV(tmp, src, w, h, r, scale)
}
