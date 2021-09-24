package physarum

import (
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/gonum/stat"
)

type Texture struct {
	w        int
	h        int
	id       uint32
	buf      []uint8
	acc      []float32
	r        [][]float32
	g        [][]float32
	b        [][]float32
	min      []float32
	max      []float32
	settings Settings
}

func NewTexture(settings Settings) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return &Texture{id: id, settings: settings}
}

func (t *Texture) Init(count int) {
	const N = 65536
	t.w = t.settings["width"].(int)
	t.h = t.settings["height"].(int)
	t.buf = make([]uint8, t.w*t.h*3)
	t.acc = make([]float32, t.w*t.h*3)
	t.r = make([][]float32, count)
	t.g = make([][]float32, count)
	t.b = make([][]float32, count)
	for i := 0; i < count; i++ {
		t.r[i] = make([]float32, N)
		t.g[i] = make([]float32, N)
		t.b[i] = make([]float32, N)
	}
	max := float32(t.settings["particles"].(int)) / float32(t.w*t.h) * 10
	t.min = make([]float32, count)
	t.max = make([]float32, count)
	for i := range t.min {
		t.min[i] = 0
		t.max[i] = max
	}
}

func (t *Texture) SetPalette(palette Palette, gamma float32) {
	count := len(t.r)
	N := len(t.r[0])
	for i := 0; i < count; i++ {
		c := palette[i]
		for j := 0; j < N; j++ {
			p := float32(j) / float32(N-1)
			p = float32(math.Pow(float64(p), float64(gamma)))
			t.r[i][j] = float32(c.R) * p
			t.g[i][j] = float32(c.G) * p
			t.b[i][j] = float32(c.B) * p
		}
	}
	palette.Print()
}

func (t *Texture) ShufflePalette() {
	rand.Shuffle(len(t.r), func(i, j int) {
		t.r[i], t.r[j] = t.r[j], t.r[i]
		t.g[i], t.g[j] = t.g[j], t.g[i]
		t.b[i], t.b[j] = t.b[j], t.b[i]
	})
}

func (t *Texture) AutoLevel(data [][]float32, minPercentile, maxPercentile float64) {
	for i, grid := range data {
		temp := make([]float64, len(grid))
		for j, v := range grid {
			temp[j] = float64(v)
		}
		sort.Float64s(temp)
		t.min[i] = float32(stat.Quantile(minPercentile, stat.Empirical, temp, nil))
		t.max[i] = float32(stat.Quantile(maxPercentile, stat.Empirical, temp, nil))
	}
}

func (t *Texture) update(data [][]float32) {
	// waitgroup for threads
	var wg sync.WaitGroup

	for i := range t.acc {
		t.acc[i] = 0
	}
	f := float32(len(t.r[0]) - 1)
	for i, grid := range data {
		// New thread to wait on
		wg.Add(1)

		go func(i int, grid []float32) {
			// Defer
			defer wg.Done()

			// Do parallel loops
			min, max := t.min[i], t.max[i]
			m := 1 / float32(max-min)
			for j, value := range grid {
				p := (value - min) * m
				if p < 0 {
					p = 0
				}
				if p > 1 {
					p = 1
				}
				index := int(p * f)
				t.acc[j*3+0] += t.r[i][index]
				t.acc[j*3+1] += t.g[i][index]
				t.acc[j*3+2] += t.b[i][index]
			}
		}(i, grid)
	}

	// Wait for the threads to finish
	wg.Wait()

	for i, value := range t.acc {
		if value > 255 {
			value = 255
		}
		t.buf[i] = uint8(value)
	}
}

func (t *Texture) draw(window *glfw.Window) {
	const padding = 0
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / float32(t.w)
	s2 := float32(h) / float32(t.h)
	f := float32(1 - padding)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func (t *Texture) Draw(window *glfw.Window, data [][]float32) {
	t.update(data)
	gl.BindTexture(gl.TEXTURE_2D, t.id)
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGB, int32(t.w), int32(t.h),
		0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(t.buf))
	t.draw(window)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (t *Texture) GetFramebuffer() []uint8 {
	return t.buf
}
