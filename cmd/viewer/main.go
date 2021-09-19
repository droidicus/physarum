package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"

	// _ "net/http/pprof"

	"github.com/droidicus/physarum/pkg/physarum"
	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// const (
// 	width      = 512
// 	height     = 512
// 	particles  = 1 << 20
// 	blurRadius = 1
// 	blurPasses = 2+
// 	zoomFactor = 1
// 	scale      = 1
// 	gamma      = 1 / 2.2
// 	title      = "physarum"
// )

const (
	// width      = 2048
	// height     = 1024
	// particles  = 1 << 23
	blurRadius = 1
	blurPasses = 2
	zoomFactor = 1
	scale      = 0.5
	gamma      = 1 / 2.2
	title      = "physarum"
)

var Configs = []physarum.Config{
	// NOTE: These no longer seem to work...

	// cyclones
	// {4, 0.87946403, 42.838207, 0.97047323, 2.8447638, 5, 0.29681, 1.4512},
	// {4, 1.7357124, 17.430664, 0.30490428, 2.1706762, 5, 0.27878627, 0.46232897},

	// dunes
	// {2, 0.99931663, 44.21652, 1.9704952, 1.4215798, 5, 0.1580779, 0.7574965},
	// {2, 1.9694986, 1.294038, 0.5384646, 1.1613986, 5, 0.21102181, 1.5123861},

	// dot grid
	// {1.3333334, 1.3433642, 49.39263, 0.91616887, 0.69644034, 5, 0.17888786, 0.2036435},
	// {1.3333334, 0.0856143, 1.6695175, 1.8827246, 2.3155663, 5, 0.14249614, 0.0026361942},
	// {1.3333334, 0.7959472, 33.977413, 0.5246451, 2.2891424, 5, 0.22549233, 1.4248871},

	// untitled
	// {1.7433162, 56.586433, 0.45428953, 0.78228176, 5, 0.19172272, 1.6682954},
	// {1.8340914, 1.6538872, 1.4098115, 1.6714363, 5, 0.17746642, 1.491355},
	// {0.0049473564, 13.269191, 0.033447478, 1.0102618, 5, 0.2197167, 1.6166985},
	// {0.37645763, 31.045816, 0.81319964, 2.5322618, 5, 0.10834738, 1.5170672},
	// {0.7355474, 14.832715, 0.2641479, 0.8953786, 5, 0.14977153, 0.14632958},

	// cool
	// {1.4107815, 61.27741, 0.49201587, 1.3007548, 5, 0.49895996, 1.0823951},
	// {1.1534524, 13.299458, 0.48315683, 1.8219115, 5, 0.41845483, 0.4055887},
	// {0.31089303, 60.62575, 1.0241486, 0.39942655, 5, 0.4576149, 0.24079543},
	// {0.40245488, 27.844227, 1.9592205, 0.5504824, 5, 0.19568197, 1.1694417},
	// {1.227412, 1.7987814, 0.39546785, 1.2640203, 5, 0.14201605, 0.77068233},
}

func init() {
	runtime.LockOSThread()
}

func makeModel(settings physarum.Settings) *physarum.Model {
	configs := physarum.RandomConfigs(2 + rand.Intn(4))
	if len(Configs) > 0 {
		configs = Configs
	}
	table := physarum.RandomAttractionTable(len(configs))
	model := physarum.NewModel(
		settings["width"].(int), settings["height"].(int), settings["particles"].(int), blurRadius, blurPasses,
		zoomFactor, configs, table, "random_circle_random")
	physarum.PrintConfigs(model.Configs, model.AttractionTable)
	physarum.SummarizeConfigs(model.Configs)
	log.Println()
	return model
}

func main() {
	settings := physarum.NewSettings()
	log.Println(settings)

	rand.Seed(settings["seed"].(int64))

	num_steps := 1

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
	displayWidth := int(float32(settings["width"].(int)) * scale)
	displayHeight := int(float32(settings["height"].(int)) * scale)
	window, err := glfw.CreateWindow(displayWidth, displayHeight, title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()

	// initialize gl
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	gl.Enable(gl.TEXTURE_2D)

	var model *physarum.Model
	texture := physarum.NewTexture(settings)

	reset := func() {
		model = makeModel(settings)
		texture.Init(len(model.Configs))
		texture.SetPalette(physarum.RandomPalette(), gamma)
	}

	reset()

	window.SetKeyCallback(func(window *glfw.Window, key glfw.Key, code int, action glfw.Action, mods glfw.ModifierKey) {
		// Manage key presses
		if action == glfw.Press {
			switch key {
			case glfw.KeySpace:
				reset()
			case glfw.KeyR:
				model.StartOver()
			case glfw.KeyP:
				texture.SetPalette(physarum.RandomPalette(), gamma)
			case glfw.KeyO:
				texture.ShufflePalette()
			case glfw.KeyA:
				texture.AutoLevel(model.Data(), 0.001, 0.999)
			case glfw.KeyKPAdd:
				if num_steps < math.MaxInt {
					num_steps++
				}
			case glfw.KeyKPSubtract:
				if num_steps > 1 {
					num_steps--
				}
			case glfw.Key1:
				model.InitType = "random"
			case glfw.Key2:
				model.InitType = "point"
			case glfw.Key3:
				model.InitType = "random_circle_random"
			case glfw.Key4:
				model.InitType = "random_circle_out"
			case glfw.Key5:
				model.InitType = "random_circle_in"
			case glfw.Key6:
				model.InitType = "random_circle_quads"
			case glfw.Key7:
				model.InitType = "random_circle_cw"
			}
		}
	})

	// Set up goroutine to save video with FFMPEG if required
	saveVideo := true
	var video *physarum.Video
	videoFameChann := make(chan []uint8, 1024)
	videoDoneChann := make(chan bool)
	if saveVideo {
		video = physarum.NewVideo(settings)
		go video.SaveVideoFfmpeg(videoFameChann, videoDoneChann)
	}

	// Until the window needs closing
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		for i := 0; i < num_steps; i++ {
			// Step model at desired rate
			model.Step()
		}
		if saveVideo {
			// Send framebuffer for rendering into video if required
			videoFameChann <- texture.GetFramebuffer()
		}

		// Display image and manage interface
		texture.Draw(window, model.Data())
		window.SwapBuffers()
		glfw.PollEvents()
	}

	// Close the channel and let the video finish
	close(videoFameChann)
	log.Println("sent all frames, waiting for encoding to complete")
	<-videoDoneChann // wait for the goroutine to be finished
}
