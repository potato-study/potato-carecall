package recorder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"log"
	"math"
	"os"
	"time"
)

const (
	MaxInitialWaitTime = 10 * time.Second
	MaxSilenceDuration = 3 * time.Second
	MaxRecordingTime   = 10 * time.Second
)

func RecordAudio() ([]byte, error) {
	selectedDevice, err := getDevice()
	if err != nil {
		return nil, err
	}

	stream, buffer, err := initializeStream(selectedDevice)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	defer stream.Stop()

	var audioBuffer bytes.Buffer
	if err := recordStream(stream, buffer, &audioBuffer); err != nil {
		return nil, err
	}

	return createWAVData(&audioBuffer, 16000, 1, 16)
}

func initializeStream(device *portaudio.DeviceInfo) (*portaudio.Stream, []int16, error) {
	sampleRate := 16000.0
	channels := 1
	framesPerBuffer := 1024

	buffer := make([]int16, framesPerBuffer)
	stream, err := portaudio.OpenStream(portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: channels,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: framesPerBuffer,
	}, buffer)
	if err != nil {
		return nil, nil, fmt.Errorf("오디오 스트림 생성 중 오류 발생: %v", err)
	}

	if err := stream.Start(); err != nil {
		return nil, nil, fmt.Errorf("오디오 스트림 시작 중 오류 발생: %v", err)
	}

	return stream, buffer, nil
}

func recordStream(stream *portaudio.Stream, buffer []int16, audioBuffer *bytes.Buffer) error {
	fmt.Print("녹음 시작...")
	silenceDuration := 0 * time.Second
	lastSoundTime := time.Now()
	startTime := time.Now()

	for {
		if err := stream.Read(); err != nil {
			return fmt.Errorf("오디오 데이터 읽기 중 오류 발생: %v", err)
		}

		binary.Write(audioBuffer, binary.LittleEndian, buffer)
		rms := calculateRMS(buffer)

		if rms > 500 {
			lastSoundTime = time.Now()
		}

		silenceDuration = time.Since(lastSoundTime)
		if time.Since(startTime) > MaxInitialWaitTime && silenceDuration >= MaxSilenceDuration {
			fmt.Println("녹음을 종료합니다. 무음 시간:", silenceDuration)
			break
		}

		if time.Since(startTime) >= MaxRecordingTime {
			fmt.Println("녹음을 종료합니다. 녹음 시간 초과")
			break
		}
	}
	return nil
}

func calculateRMS(buffer []int16) float64 {
	var sum int64
	for _, sample := range buffer {
		sum += int64(sample) * int64(sample)
	}
	mean := float64(sum) / float64(len(buffer))
	return math.Sqrt(mean)
}

func createWAVData(audioBuffer *bytes.Buffer, sampleRate, channels, bitDepth int) ([]byte, error) {
	dataLen := audioBuffer.Len()
	blockAlign := channels * (bitDepth / 8)

	wavHeader := &bytes.Buffer{}
	wavHeader.WriteString("RIFF")
	binary.Write(wavHeader, binary.LittleEndian, uint32(36+dataLen))
	wavHeader.WriteString("WAVE")
	wavHeader.WriteString("fmt ")
	binary.Write(wavHeader, binary.LittleEndian, uint32(16))
	binary.Write(wavHeader, binary.LittleEndian, uint16(1))
	binary.Write(wavHeader, binary.LittleEndian, uint16(channels))
	binary.Write(wavHeader, binary.LittleEndian, uint32(sampleRate))
	binary.Write(wavHeader, binary.LittleEndian, uint32(sampleRate*blockAlign))
	binary.Write(wavHeader, binary.LittleEndian, uint16(blockAlign))
	binary.Write(wavHeader, binary.LittleEndian, uint16(bitDepth))
	wavHeader.WriteString("data")
	binary.Write(wavHeader, binary.LittleEndian, uint32(dataLen))

	wavData := append(wavHeader.Bytes(), audioBuffer.Bytes()...)
	if err := os.WriteFile("output.wav", wavData, 0644); err != nil {
		log.Fatalf("Failed to write WAV data to file: %v", err)
	}

	return wavData, nil
}
