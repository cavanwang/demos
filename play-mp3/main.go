// This package shows how to get mp3 playing duration and how to play a mp3 on our computer using default audio device.
// I referred the codes from:
// https://blog.csdn.net/weixin_43823363/article/details/118101380
// https://blog.csdn.net/andrewgithub/article/details/119920839
package main

/*
#define MINIMP3_IMPLEMENTATION

#include "minimp3.h"
#include <stdlib.h>
#include <stdio.h>

int decode(mp3dec_t *dec, mp3dec_frame_info_t *info, unsigned char *data, int *length, unsigned char *decoded, int *decoded_length) {
    int samples;
    short pcm[MINIMP3_MAX_SAMPLES_PER_FRAME];
    samples = mp3dec_decode_frame(dec, data, *length, pcm, info);
    *decoded_length = samples * info->channels * 2;
    *length -= info->frame_bytes;
    unsigned char buffer[samples * info->channels * 2];
    memcpy(buffer, (unsigned char*)&(pcm), sizeof(short) * samples * info->channels);
    memcpy(decoded, buffer, sizeof(short) * samples * info->channels);
    return info->frame_bytes;
}
*/
import "C"
import (
	"context"
	"log"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	// Get the milliseconds of mp3 file.
	by, err := os.ReadFile("./test.mp3")
	if err != nil {
		log.Fatal(err)
	}
	milliseconds, err := GetMP3PlayMilliseconds(by)
	if err != nil {
		log.Fatal(err)
	}
	println("mp3 file milliseconds is: ", milliseconds)

	// Prepare mp3 input stream for playing.
	audioFile, err := os.Open("./test.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer audioFile.Close()
	audioStreamer, format, err := mp3.Decode(audioFile)
	if err != nil {
		log.Fatal(err)
	}
	defer audioStreamer.Close()

	// SampleRate is the number of samples per second.
	_ = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Start to play.
	done := make(chan bool)
	speaker.Play(beep.Seq(audioStreamer, beep.Callback(func() {
		done <- true
	})))

	// Wait for playing done.
	<-done
}

const maxSamplesPerFrame = 1152 * 2

// Decoder decode the mp3 stream by minimp3
type Decoder struct {
	readerLocker  *sync.Mutex
	data          []byte
	decoderLocker *sync.Mutex
	decodedData   []byte
	decode        C.mp3dec_t
	info          C.mp3dec_frame_info_t
	context       context.Context
	contextCancel context.CancelFunc
	SampleRate    int
	Channels      int
	Kbps          int
	Layer         int

	originalEof bool // if the original reader is EOF, set this to true
}

func GetMP3PlayMilliseconds(mp3Data []byte) (milliseconds int, err error) {
	dec, _, err := DecodeFull(mp3Data)
	if err != nil {
		return 0, err
	}
	// Play time = (FileSize(byte) - 128(ID3 information)) * 8(to bit) / kbps(kilo bit to bit)
	milliseconds = (len(mp3Data) - 128) * 8 / dec.Kbps
	return milliseconds, nil
}

// DecodeFull put all of the mp3 data to decode.
func DecodeFull(mp3 []byte) (dec *Decoder, decodedData []byte, err error) {
	dec = new(Decoder)
	dec.decode = C.mp3dec_t{}
	C.mp3dec_init(&dec.decode)
	info := C.mp3dec_frame_info_t{}
	var length = C.int(len(mp3))
	for {
		var decoded = [maxSamplesPerFrame * 2]byte{}
		var decodedLength = C.int(0)
		frameSize := C.decode(&dec.decode,
			&info, (*C.uchar)(unsafe.Pointer(&mp3[0])),
			&length, (*C.uchar)(unsafe.Pointer(&decoded[0])),
			&decodedLength)
		if int(frameSize) == 0 {
			break
		}
		decodedData = append(decodedData, decoded[:decodedLength]...)
		if int(frameSize) < len(mp3) {
			mp3 = mp3[int(frameSize):]
		}
		dec.SampleRate = int(info.hz)
		dec.Channels = int(info.channels)
		dec.Kbps = int(info.bitrate_kbps)
		dec.Layer = int(info.layer)
	}
	return
}
