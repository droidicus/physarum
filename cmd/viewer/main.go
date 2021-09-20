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

func init() {
	// droid: I don't know what this does, but the OpenGL window doesn't work if this is removed...
	runtime.LockOSThread()
}

func makeModel(settings physarum.Settings) *physarum.Model {
	model := physarum.NewModel(
		settings["width"].(int),
		settings["height"].(int),
		settings["particles"].(int),
		settings["blurRadius"].(int),
		settings["blurPasses"].(int),
		float32(settings["zoomFactor"].(int)),
		settings["configs"].([]physarum.Config),
		settings["attract_table"].([][]float32),
		settings["initType"].(string),
	)
	log.Println("********************")
	physarum.PrintConfigs(model.Configs, model.AttractionTable)
	physarum.SummarizeConfigs(model.Configs)
	log.Println("********************")
	return model
}

func main() {
	settings := physarum.NewSettings()
	log.Println(settings)
	log.Println(string(settings.GetSettingsJson()))
	settings.WriteSettingsToFile()

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
	gamma := float32(settings["scale"].(float64))
	displayWidth := int(float32(settings["width"].(int)) * gamma)
	displayHeight := int(float32(settings["height"].(int)) * gamma)
	window, err := glfw.CreateWindow(displayWidth, displayHeight, "physarum", nil, nil)
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
		texture.SetPalette(settings["palette"].(physarum.Palette), float32(settings["gamma"].(float64)))
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
				settings["palette"] = physarum.RandomPalette()
				texture.SetPalette(settings["palette"].(physarum.Palette), float32(settings["gamma"].(float64)))
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
