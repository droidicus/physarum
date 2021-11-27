package physarum

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

// All the supported init types
const (
	Random             = "random"
	Point              = "point"
	RandomCircleRandom = "random_circle_random"
	RandomCircleOut    = "random_circle_out"
	RandomCircleIn     = "random_circle_in"
	RandomCircleCW     = "random_circle_cw"
	RandomCircleQuads  = "random_circle_quads"
)

// All of the supported init types in a slice
var AllInitTypes = [...]string{
	Random,
	Point,
	RandomCircleRandom,
	RandomCircleOut,
	RandomCircleIn,
	RandomCircleCW,
	RandomCircleQuads,
}

// Pick a random init type from above
func RandomInitType() string {
	return AllInitTypes[rand.Intn(len(AllInitTypes))]
}

type Model struct {
	W int
	H int

	BlurRadius int
	BlurPasses int

	ZoomFactor float32

	Configs         []Config
	AttractionTable [][]float32

	Grids     []*Grid
	Particles []Particle

	Iteration int

	InitType string

	seed int64
}

func MakeModel(settings *Settings) *Model {
	model := NewModel(
		settings.Width,
		settings.Height,
		settings.Particles,
		settings.BlurRadius,
		settings.BlurPasses,
		settings.ZoomFactor,
		settings.Configs,
		settings.AttractionTable,
		settings.InitType,
		settings.Seed,
	)

	log.Println("********************")
	PrintConfigs(model.Configs, model.AttractionTable)
	SummarizeConfigs(model.Configs)
	log.Println("********************")

	return model
}

func NewModel(
	w, h, numParticles, blurRadius, blurPasses int, zoomFactor float32,
	configs []Config, attractionTable [][]float32, initType string, seed int64) *Model {

	grids := make([]*Grid, len(configs))
	numParticlesPerConfig := int(math.Ceil(
		float64(numParticles) / float64(len(configs))))
	actualNumParticles := numParticlesPerConfig * len(configs)
	particles := make([]Particle, actualNumParticles)
	m := &Model{
		w, h, blurRadius, blurPasses, zoomFactor,
		configs, attractionTable, grids, particles, 0, initType, seed}
	m.StartOver()
	return m
}

func (m *Model) StartOver() {
	numParticlesPerConfig := len(m.Particles) / len(m.Configs)
	m.Particles = m.Particles[:0]
	m.Iteration = 0
	for c := range m.Configs {
		m.Grids[c] = NewGrid(m.W, m.H)
		for i := 0; i < numParticlesPerConfig; i++ {
			var x, y, a float32
			switch m.InitType {
			case Random:
				x = rand.Float32() * float32(m.W)
				y = rand.Float32() * float32(m.H)
				a = rand.Float32() * 2 * math.Pi
			case Point:
				x = float32(m.W) / 2
				y = float32(m.H) / 2
				a = rand.Float32() * 2 * math.Pi
			case RandomCircleRandom:
				a = rand.Float32() * 2 * math.Pi
				circle_radius_fraction := 0.25
				r := circle_radius_fraction * math.Min(float64(m.H), float64(m.W)) * math.Sqrt(rand.Float64())
				x_tmp, y_tmp := math.Sincos(float64(a))
				x = float32(r*x_tmp) + float32(m.W)/2
				y = float32(r*y_tmp) + float32(m.H)/2
				a = rand.Float32() * 2 * math.Pi
			case RandomCircleOut:
				a = rand.Float32() * 2 * math.Pi
				circle_radius_fraction := 0.25
				r := circle_radius_fraction * math.Min(float64(m.H), float64(m.W)) * math.Sqrt(rand.Float64())
				y_tmp, x_tmp := math.Sincos(float64(a))
				x = float32(r*x_tmp) + float32(m.W)/2
				y = float32(r*y_tmp) + float32(m.H)/2
			case RandomCircleIn:
				a = rand.Float32() * 2 * math.Pi
				circle_radius_fraction := 0.25
				r := circle_radius_fraction * math.Min(float64(m.H), float64(m.W)) * math.Sqrt(rand.Float64())
				y_tmp, x_tmp := math.Sincos(float64(a))
				x = float32(r*x_tmp) + float32(m.W)/2
				y = float32(r*y_tmp) + float32(m.H)/2
				a_tmp := float64(a + math.Pi)
				a = float32(math.Atan2(math.Sin(a_tmp), math.Cos(a_tmp)))
			case RandomCircleCW:
				a = rand.Float32() * 2 * math.Pi
				circle_radius_fraction := 0.25
				r := circle_radius_fraction * math.Min(float64(m.H), float64(m.W)) * math.Sqrt(rand.Float64())
				y_tmp, x_tmp := math.Sincos(float64(a))
				x = float32(r*x_tmp) + float32(m.W)/2
				y = float32(r*y_tmp) + float32(m.H)/2
				a_tmp := float64(a + math.Pi/2.0)
				a = float32(math.Atan2(math.Sincos(a_tmp)))
			case RandomCircleQuads:
				a = rand.Float32() * 2 * math.Pi
				circle_radius_fraction := 0.25
				r := circle_radius_fraction * math.Min(float64(m.H), float64(m.W)) * math.Sqrt(rand.Float64())
				x_tmp, y_tmp := math.Sincos(float64(a))
				x = float32(r*x_tmp) + float32(m.W)/2
				y = float32(r*y_tmp) + float32(m.H)/2
			}
			if false { // for testing, it is a LOT...
				fmt.Println(x-float32(m.W)/2, y-float32(m.H)/2, a)
			}
			p := Particle{x, y, a, uint32(c)}
			m.Particles = append(m.Particles, p)
		}
	}
}

func (m *Model) Step() {
	updateParticle := func(rnd *rand.Rand, i int) {
		p := m.Particles[i]
		config := m.Configs[p.C]
		grid := m.Grids[p.C]

		// u := p.X / float32(m.W)
		// v := p.Y / float32(m.H)

		sensorDistance := config.SensorDistance * m.ZoomFactor
		sensorAngle := config.SensorAngle
		rotationAngle := config.RotationAngle
		stepDistance := config.StepDistance * m.ZoomFactor

		sinResult, cosResult := sincos(p.A)
		xc := p.X + cosResult*sensorDistance
		yc := p.Y + sinResult*sensorDistance

		sinResult, cosResult = sincos(p.A - sensorAngle)
		xl := p.X + cosResult*sensorDistance
		yl := p.Y + sinResult*sensorDistance

		sinResult, cosResult = sincos(p.A + sensorAngle)
		xr := p.X + cosResult*sensorDistance
		yr := p.Y + sinResult*sensorDistance

		C := grid.GetTemp(xc, yc)
		L := grid.GetTemp(xl, yl)
		R := grid.GetTemp(xr, yr)

		var da float32
		if true {
			da = rotationAngle * direction(rnd, C, L, R)
		} else {
			// TODO: what does this do???
			da = rotationAngle * weightedDirection(rnd, C, L, R)
		}
		p.A = Shift(p.A+da, 2*math.Pi)
		sinResult, cosResult = sincos(p.A)
		p.X = Shift(p.X+cosResult*stepDistance, float32(m.W))
		p.Y = Shift(p.Y+sinResult*stepDistance, float32(m.H))
		m.Particles[i] = p
	}

	updateParticles := func(wi, wn int, wg *sync.WaitGroup) {
		seed := (int64(m.Iteration)<<8 | int64(wi)) + int64(m.seed)
		rnd := rand.New(rand.NewSource(seed))
		n := len(m.Particles)
		batch := int(math.Ceil(float64(n) / float64(wn)))
		i0 := wi * batch
		i1 := i0 + batch
		if wi == wn-1 {
			i1 = n
		}
		for i := i0; i < i1; i++ {
			updateParticle(rnd, i)
		}
		wg.Done()
	}

	updateGrids := func(c int, wg *sync.WaitGroup) {
		config := m.Configs[c]
		grid := m.Grids[c]
		for _, p := range m.Particles {
			if uint32(c) == p.C {
				grid.Add(p.X, p.Y, config.DepositionAmount)
			}
		}
		grid.BoxBlur(m.BlurRadius, m.BlurPasses, config.DecayFactor)
		wg.Done()
	}

	combineGrids := func(c int, wg *sync.WaitGroup) {
		grid := m.Grids[c]
		for i := range grid.Temp {
			grid.Temp[i] = 0
		}
		for i, other := range m.Grids {
			factor := m.AttractionTable[c][i]
			for j, value := range other.Data {
				grid.Temp[j] += value * factor
			}
		}
		wg.Done()
	}

	var wg sync.WaitGroup

	// step 1: combine grids
	for i := range m.Configs {
		wg.Add(1)
		go combineGrids(i, &wg)
	}
	wg.Wait()

	// step 2: move particles
	wn := runtime.NumCPU()
	for wi := 0; wi < wn; wi++ {
		wg.Add(1)
		go updateParticles(wi, wn, &wg)
	}
	wg.Wait()

	// step 3: deposit, blur, and decay
	for i := range m.Configs {
		wg.Add(1)
		go updateGrids(i, &wg)
	}
	wg.Wait()

	m.Iteration++
}

func (m *Model) Data() [][]float32 {
	result := make([][]float32, len(m.Grids))
	for i, grid := range m.Grids {
		result[i] = make([]float32, len(grid.Data))
		copy(result[i], grid.Data)
	}
	return result
}

func direction(rnd *rand.Rand, C, L, R float32) float32 {
	if C > L && C > R {
		return 0
	} else if C < L && C < R {
		return float32((rnd.Int63()&1)<<1 - 1)
	} else if L < R {
		return 1
	} else if R < L {
		return -1
	}
	return 0
}

func weightedDirection(rnd *rand.Rand, C, L, R float32) float32 {
	W := [3]float32{C, L, R}
	D := [3]float32{0, -1, 1}

	if W[0] > W[1] {
		W[0], W[1] = W[1], W[0]
		D[0], D[1] = D[1], D[0]
	}
	if W[0] > W[2] {
		W[0], W[2] = W[2], W[0]
		D[0], D[2] = D[2], D[0]
	}
	if W[1] > W[2] {
		W[1], W[2] = W[2], W[1]
		D[1], D[2] = D[2], D[1]
	}

	a := W[1] - W[0]
	b := W[2] - W[1]
	if rnd.Float32()*(a+b) < a {
		return D[1]
	}
	return D[2]
}
