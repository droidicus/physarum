// NOTE: This requires FFMPEG to be avaliable on the PATH or in the local directory

package physarum

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type Video struct {
	FrameCount int
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	settings   *Settings
}

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}

func NewVideo(settings *Settings) *Video {
	// Make a new video, and start the process to receive data later
	v := &Video{settings: settings}
	v.StartVideo()
	return v
}

func (v *Video) StartVideo() {
	v.cmd = exec.Command(
		"ffmpeg.exe",
		"-y",             // Overwrite output file if it exists
		"-f", "rawvideo", // Raw framebuffer
		"-pix_fmt", "rgb24",
		"-s", fmt.Sprintf("%vx%v", v.settings.Width, v.settings.Height), // Resolution
		"-r", fmt.Sprint(v.settings.Fps),
		"-i", "pipe:0", // Take input from stdin
		"-c:v", "libx264",
		"-profile:v", "high", // bells
		"-preset", "veryslow", // and
		"-tune", "film", // whistles
		"-bf", "2", // 2 b-frames
		"-rc-lookahead", "2",
		"-g", fmt.Sprint(v.settings.Fps/2), // Closed GOP at half frame rate
		"-crf", fmt.Sprint(v.settings.Crf),
		"-pix_fmt", "yuv420p",
		"-movflags", "frag_keyframe", // Fragmented output file for crash recoverability
		v.settings.GetFilePathWOExtension()+".mp4",
	)

	// Set up pipe to send data to FFMPEG
	stdin, err := v.cmd.StdinPipe()
	check(err)
	v.stdin = stdin

	// Redirect both stdout and stderr from the process to the console
	v.cmd.Stderr = os.Stderr
	v.cmd.Stdout = os.Stdout

	// Start the process
	check(v.cmd.Start())
}

func (v *Video) SaveVideoFfmpeg(videoFameChan <-chan []uint8, videoDoneChan chan<- bool) {
	for frame := range videoFameChan {
		// Write framebuffer to the pipe and check for errors
		_, err := v.stdin.Write(frame)
		check(err)
		v.FrameCount++
	}

	fmt.Println("Video Frame Channel Closed, shutting down video encode and cleaning up!")

	// Close the pipe and wait for the process to complete
	check(v.stdin.Close())
	v.cmd.Wait()

	// Run second pass to defragment the file
	v.FaststartVideoFfmpeg()

	// Send sync signal, Done!
	videoDoneChan <- true
}

func (v *Video) FaststartVideoFfmpeg() {
	// run a second pass to allow defragmentation and "faststart" optimization
	faststart_cmd := exec.Command(
		"ffmpeg.exe",
		"-y",
		"-i", v.settings.GetFilePathWOExtension()+".mp4",
		"-c", "copy",
		"-movflags", "faststart",
		v.settings.GetFilePathWOExtension()+"_faststart.mp4",
	)
	stdoutStderr, err := faststart_cmd.CombinedOutput()
	check(err)
	fmt.Print(string(stdoutStderr))

	// Get rid of the fragmented video file
	check(os.Remove(v.settings.GetFilePathWOExtension() + ".mp4"))
}
