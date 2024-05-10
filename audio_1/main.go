package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	"github.com/orcaman/writerseeker"
)

func main() {
	// Required output format, every sample would get resampled to that format
	var format = beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}

	var err error
	var stream1, stream2 beep.StreamSeekCloser
	var stream1Format, stream2Format beep.Format

	fmt.Printf("Output audio format:\n Sample rate: %d,\n Channels: %d,\n Precision: %d\n", format.SampleRate, format.NumChannels, format.Precision)
	time.Sleep(time.Second)

	// wav file
	audioFile1, err := os.Open("sounds/tone1.wav")
	if err != nil {
		slog.Error("Error when opening audio file", "Err", err)
		return
	} else {
		stream1, stream1Format, err = wav.Decode(audioFile1)
		if err != nil {
			slog.Error("Error when decoding audio file", "Err", err)
			return
		}
		defer stream1.Close()
	}

	// mp3 file
	audioFile2, err := os.Open("sounds/flute.mp3")
	if err != nil {
		slog.Error("Error when opening audio file", "Err", err)
		return
	} else {
		stream2, stream2Format, err = mp3.Decode(audioFile2)
		if err != nil {
			slog.Error("Error when decoding audio file", "Err", err)
			return
		}
		defer stream2.Close()
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Using beep.Streamer
	{
		fmt.Println("Playing using beep.Streamer")
		var audioStream beep.Streamer
		audioStream = beep.Seq(
			beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1),
			beep.Callback(func() { stream1.Seek(0) }))
		audioStream = beep.Seq(audioStream,
			beep.Resample(4, stream2Format.SampleRate, format.SampleRate, stream2),
			beep.Callback(func() { stream2.Seek(0) }))
		audioStream = beep.Seq(audioStream,
			beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1),
			beep.Callback(func() { stream1.Seek(0) }))
		var done = false
		audioStream = beep.Seq(audioStream, beep.Callback(func() {
			done = true
		}))
		speaker.Play(audioStream)

		for !done {
			time.Sleep(time.Millisecond)
		}
	}
	time.Sleep(time.Second * 2)

	// Using beep.Buffer
	{
		fmt.Println("Playing using beep.Buffer")
		var buf = beep.NewBuffer(format)
		buf.Append(beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1))
		buf.Append(beep.Callback(func() { stream1.Seek(0) }))
		buf.Append(beep.Resample(4, stream2Format.SampleRate, format.SampleRate, stream2))
		buf.Append(beep.Callback(func() { stream2.Seek(0) }))
		buf.Append(beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1))
		buf.Append(beep.Callback(func() { stream1.Seek(0) }))
		var done = false
		var audio = buf.Streamer(0, buf.Len())
		speaker.Play(beep.Seq(audio, beep.Callback(func() {
			done = true
		})))

		for !done {
			time.Sleep(time.Millisecond)
		}

		// Save streamer audio to .wav file
		// var file, _ = os.Create("test.wav")
		// wav.Encode(file, buf.Streamer(0, buf.Len()), format)

		// Get array of bytes of the audio streamer
		var ws = &writerseeker.WriterSeeker{}
		err = wav.Encode(ws, buf.Streamer(0, buf.Len()), format)
		if err != nil {
			slog.Error("Error when writing data to memory buffer", "Err", err)
		} else {
			fmt.Println(ws)
		}
	}
	time.Sleep(time.Second * 2)

	// Using queue of streamers
	{
		fmt.Println("Playing using queue of streamers")
		var streamers Queue
		streamers.q = make([]beep.Streamer, 0, 10) // Create streamer array with a capacity of 10 to not allocate new array every append call
		streamers.q = append(streamers.q, beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1))
		streamers.q = append(streamers.q, beep.Callback(func() { stream1.Seek(0) }))
		streamers.q = append(streamers.q, beep.Resample(4, stream2Format.SampleRate, format.SampleRate, stream2))
		streamers.q = append(streamers.q, beep.Callback(func() { stream2.Seek(0) }))
		streamers.q = append(streamers.q, beep.Resample(4, stream1Format.SampleRate, format.SampleRate, stream1))
		streamers.q = append(streamers.q, beep.Callback(func() { stream1.Seek(0) }))
		var done = false
		streamers.q = append(streamers.q, beep.Callback(func() {
			done = true
		}))
		speaker.Play(&streamers)

		for !done {
			time.Sleep(time.Millisecond)
		}
	}
}

// Queue stuct code yionked from examples
type Queue struct {
	q []beep.Streamer
}

func (q *Queue) Err() error {
	return nil
}

func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	// We use the filled variable to track how many samples we've
	// successfully filled already. We loop until all samples are filled.
	filled := 0
	for filled < len(samples) {
		// There are no streamers in the queue, so we stream silence.
		if len(q.q) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		// We stream from the first streamer in the queue.
		n, ok := q.q[0].Stream(samples[filled:])
		// If it's drained, we pop it from the queue, thus continuing with
		// the next streamer.
		if !ok {
			q.q = q.q[1:]
		}
		// We update the number of filled samples.
		filled += n
	}
	return len(samples), true
}
