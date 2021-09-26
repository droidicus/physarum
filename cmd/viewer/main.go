package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"runtime"

	// "net/http"
	// _ "net/http/pprof"

	"github.com/droidicus/physarum/pkg/physarum"
	"github.com/go-gl/gl/v4.6-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// droid: I don't know what this does, but the OpenGL window doesn't work if this is removed...
	runtime.LockOSThread()
}

func main() {
	// // Profiler
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// Command line options
	settingsFilePtr := flag.String("settings", "", "Location of a json file to use for settings to run the simulation")
	flag.Parse()

	// Read settings if they are given, and write them to record complete settings
	settings := physarum.NewSettings(*settingsFilePtr)
	settings.WriteSettingsToFile()

	// Reset the seed
	rand.Seed(settings.Seed)

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
	displayWidth := int(float32(settings.Width) * settings.Scale)
	displayHeight := int(float32(settings.Height) * settings.Scale)
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

	// Create the model to simulate
	var model *physarum.Model
	texture := physarum.NewTexture(settings)

	// Function that runs whenever we want to reset the simulation, and run it now
	reset := func() {
		model = physarum.MakeModel(settings)
		texture.Init(len(model.Configs), settings.Width, settings.Height, settings.Particles)
		texture.SetPalette(settings.Palette, settings.Gamma)
	}
	reset()

	// Manage key presses
	window.SetKeyCallback(func(window *glfw.Window, key glfw.Key, code int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			case glfw.KeySpace:
				reset()
			case glfw.KeyR:
				model.StartOver()
			case glfw.KeyP:
				settings.Palette = physarum.RandomPalette()
				texture.SetPalette(settings.Palette, settings.Gamma)
			case glfw.KeyO:
				texture.ShufflePalette()
			case glfw.KeyA:
				texture.AutoLevel(model.Data(), 0.001, 0.999)
			case glfw.KeyKPAdd:
				if settings.StepsPerFrame < math.MaxInt {
					settings.StepsPerFrame++
				}
			case glfw.KeyKPSubtract:
				if settings.StepsPerFrame > 1 {
					settings.StepsPerFrame--
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
		for i := 0; i < settings.StepsPerFrame; i++ {
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
