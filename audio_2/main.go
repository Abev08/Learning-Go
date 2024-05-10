package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

func main() {
	var err error
	var format = beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 2}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Request TTS data from StreamElements api
	var text = "This is long test message"
	var voice = "Brian"
	var response *http.Response
	response, err = http.Get(fmt.Sprintf("https://api.streamelements.com/kappa/v2/speech?voice=%s&text=%s", voice, url.QueryEscape(text)))
	if err != nil {
		slog.Error("Error when requesting tts data", "Err", err)
		return
	}
	if response.Header.Get("Content-Type") != "audio/mp3" {
		slog.Error("Response didn't contain mp3 data")
		return
	}
	var ttsData []byte
	ttsData, err = io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Error when reading response data", "Err", err)
		return
	}
	// TTS data can be saved to a file
	// if file, err := os.Create("test.mp3"); err == nil {
	// 	if n, err := file.Write(ttsData); err != nil || n != len(ttsData) {
	// 		slog.Error("Error when writing to mp3 file", "Err", err)
	// 		return
	// 	}
	// } else {
	// 	slog.Error("Error when creating mp3 file", "Err", err)
	// 	return
	// }
	var ttsStream beep.StreamSeekCloser
	var ttsFormat beep.Format
	ttsStream, ttsFormat, err = mp3.Decode(io.NopCloser(bytes.NewReader(ttsData)))
	if err != nil {
		slog.Error("Error when decoding tts data", "Err", err)
		return
	}
	defer ttsStream.Close()

	// Load tone1.wav
	var wavStream beep.StreamSeekCloser
	var wavFormat beep.Format
	if audioFile, err := os.Open("sounds/tone1.wav"); err != nil {
		slog.Error("Error when opening audio file", "Err", err)
		return
	} else {
		wavStream, wavFormat, err = wav.Decode(audioFile)
		if err != nil {
			slog.Error("Error when decoding audio file", "Err", err)
			return
		}
	}
	defer wavStream.Close()

	// Create buffer
	var buf = beep.NewBuffer(format)
	buf.Append(beep.Resample(4, wavFormat.SampleRate, format.SampleRate, wavStream))
	buf.Append(beep.Callback(func() { wavStream.Seek(0) }))
	buf.Append(beep.Resample(4, ttsFormat.SampleRate, format.SampleRate, ttsStream))
	// buf.Append(beep.Callback(func() { ttsStream.Seek(0) })) // Seek doesn't work, NopCloser doesn't implement Seek?
	buf.Append(beep.Resample(4, wavFormat.SampleRate, format.SampleRate, wavStream))
	buf.Append(beep.Callback(func() { wavStream.Seek(0) }))
	buf.Append(beep.Resample(4, ttsFormat.SampleRate, format.SampleRate, ttsStream))

	// Play
	var done = false
	var audio = buf.Streamer(0, buf.Len())
	speaker.Play(beep.Seq(audio, beep.Callback(func() {
		done = true
	})))
	for !done {
		time.Sleep(time.Millisecond)
	}
}
