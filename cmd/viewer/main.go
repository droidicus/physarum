package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

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
		texture.Update(model.Data())
		texture.SetPalette(settings.Palette, settings.Gamma)
	}
	reset()

	// Manage key presses
	window.SetKeyCallback(func(window *glfw.Window, key glfw.Key, code int, action glfw.Action, mods glfw.ModifierKey) {
		// Helper to set init type and reset the simulation
		setInitType := func(initType string) {
			// TODO: refactor the model to take a settings object, this is a hack for now
			settings.InitType = initType
			model.InitType = initType
			reset()
		}

		if action == glfw.Press {
			switch key {
			case glfw.KeySpace:
				reset()
			case glfw.KeyA:
				texture.AutoLevel(model.Data(), 0.001, 0.999)
			case glfw.KeyO:
				// TODO: this is not currently saved in settings
				texture.ShufflePalette()
			case glfw.KeyP:
				settings.Palette = physarum.RandomPalette()
				texture.SetPalette(settings.Palette, settings.Gamma)
			case glfw.KeyR:
				model.StartOver()
			case glfw.KeyW:
				err := settings.WriteSettingsToFileForce(physarum.GetSettingFileRandString())
				if err != nil {
					log.Println("Error writing settings to file!", err)
				}
			case glfw.Key1:
				setInitType("random")
			case glfw.Key2:
				setInitType("point")
			case glfw.Key3:
				setInitType("random_circle_random")
			case glfw.Key4:
				setInitType("random_circle_out")
			case glfw.Key5:
				setInitType("random_circle_in")
			case glfw.Key6:
				setInitType("random_circle_quads")
			case glfw.Key7:
				setInitType("random_circle_cw")
			case glfw.KeyKPAdd:
				if settings.StepsPerFrame < math.MaxInt {
					settings.StepsPerFrame++
				}
			case glfw.KeyKPSubtract:
				if settings.StepsPerFrame > 1 {
					settings.StepsPerFrame--
				}
			}
		}
	})

	// Set up goroutine to save video with FFMPEG if required
	saveVideo := true
	var video *physarum.Video
	videoFameChan := make(chan []uint8, 1024)
	videoDoneChan := make(chan bool)
	if saveVideo {
		video = physarum.NewVideo(settings)
		go video.SaveVideoFfmpeg(videoFameChan, videoDoneChan)
	}

	// Record start time
	start := time.Now()

	// Until the window needs closing
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		for i := 0; i < settings.StepsPerFrame; i++ {
			// Step model at desired rate
			model.Step()
		}
		if saveVideo {
			// Send a copy of the framebuffer for rendering into video if required
			videoFameChan <- texture.GetFramebufferCopy()

			// End if we have the desired number of frames
			if (settings.MaxSteps > 0) && (video.FrameCount >= settings.MaxSteps-1) {
				break
			}
		}

		// Display image and manage interface
		texture.Draw(window, model.Data())
		window.SwapBuffers()
		glfw.PollEvents()
	}

	// Get elapsed time for the simulation
	elapsed := time.Since(start)

	// Close the channel and let the video finish
	close(videoFameChan)
	log.Println("sent all frames, waiting for encoding to complete")
	<-videoDoneChan // wait for the goroutine to be finished

	// Print stats
	log.Println("Elapsed Time:\t", elapsed)
	if saveVideo {
		log.Println("Number of frames:\t", video.FrameCount)
		log.Println("Frames/sec:\t", float64(video.FrameCount)/elapsed.Seconds())
	}
}
