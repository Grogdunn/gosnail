package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/asticode/go-astiav"
	"image"
	"log"
	"math"
)

type stream struct {
	decCodec        *astiav.Codec
	decCodecContext *astiav.CodecContext
	inputStream     *astiav.Stream
}

func main() {

	videoPosition := binding.NewFloat()
	_ = videoPosition.Set(0)

	//input := "20230427_AvatarG0003.mp4"
	//input := "AvatarG0000.mp4"
	input := "/home/lcaenazzo/Downloads/20230427_AvatarG0002.mp4"

	pkt := astiav.AllocPacket()
	defer pkt.Free()
	f := astiav.AllocFrame()
	defer f.Free()
	inputFormatContext := astiav.AllocFormatContext()
	if inputFormatContext == nil {
		log.Fatal(errors.New("main: input format context is nil"))
	}
	defer inputFormatContext.Free()
	if err := inputFormatContext.OpenInput(input, nil, nil); err != nil {
		log.Fatal(fmt.Errorf("main: opening input failed: %w", err))
	}
	defer inputFormatContext.CloseInput()
	if err := inputFormatContext.FindStreamInfo(nil); err != nil {
		log.Fatal(fmt.Errorf("main: finding stream info failed: %w", err))
	}

	// Loop through streams
	streams := getVideoStreams(inputFormatContext)
	for _, s := range streams {
		defer s.decCodecContext.Free()
	}

	/////////////////////////////////////////////////

	a := app.New()
	w := a.NewWindow("GoSnail")

	videoCanvas := canvas.NewRaster(func(w, h int) image.Image {
		var img image.Image
		count := 0
		videoPos, err := videoPosition.Get()
		if err != nil {
			return image.Black
		}

		err = inputFormatContext.SeekFrame(pkt.StreamIndex(), int64(math.Round(videoPos)), astiav.NewSeekFlags(astiav.SeekFlagFrame))
		if err != nil {
			fmt.Printf("errore! %v\n", err)
			return image.Black
		}
		if err := inputFormatContext.ReadFrame(pkt); err != nil {
			if errors.Is(err, astiav.ErrEof) {
				return image.Black
			}
			log.Fatal(fmt.Errorf("main: reading frame failed: %w", err))
		}

		s, ok := streams[pkt.StreamIndex()]
		if !ok {
			return image.Black
		}

		if err := s.decCodecContext.SendPacket(pkt); err != nil {
			log.Fatal(fmt.Errorf("main: sending packet failed: %w", err))
		}

		if err := s.decCodecContext.ReceiveFrame(f); err != nil {
			if errors.Is(err, astiav.ErrEof) || errors.Is(err, astiav.ErrEagain) {
				return image.Black
			}
			log.Fatal(fmt.Errorf("main: receiving frame failed: %w", err))
		}

		count = count + 1
		img, err = convertToImg(f)
		if err != nil {
			panic(err)
		}
		return img
	})
	videoPosition.AddListener(binding.NewDataListener(func() {
		videoCanvas.Refresh()
	}))
	videoCanvas.SetMinSize(fyne.Size{
		Width:  1000,
		Height: 1000,
	})
	sliderMaxDuration := 0.0
	for _, s := range streams {
		sliderMaxDuration = math.Max(float64(s.inputStream.Duration()), sliderMaxDuration)
	}
	slider := widget.NewSliderWithData(0, sliderMaxDuration, videoPosition)
	vbox := container.NewVBox(
		videoCanvas,
		slider,
	)

	vbox.Refresh()
	w.SetContent(
		vbox,
	)
	w.Resize(fyne.Size{
		Width:  2000,
		Height: 2000,
	})
	w.ShowAndRun()
}

func getVideoStreams(inputFormatContext *astiav.FormatContext) map[int]*stream {
	streams := make(map[int]*stream) // Indexed by input stream index
	for _, is := range inputFormatContext.Streams() {
		// Only process audio or video
		if is.CodecParameters().MediaType() != astiav.MediaTypeAudio &&
			is.CodecParameters().MediaType() != astiav.MediaTypeVideo {
			continue
		}
		// Create stream
		s := &stream{inputStream: is}
		// Find decoder
		if s.decCodec = astiav.FindDecoder(is.CodecParameters().CodecID()); s.decCodec == nil {
			log.Fatal(errors.New("main: codec is nil"))
		}
		// Alloc codec context
		if s.decCodecContext = astiav.AllocCodecContext(s.decCodec); s.decCodecContext == nil {
			log.Fatal(errors.New("main: codec context is nil"))
		}

		// Update codec context
		if err := is.CodecParameters().ToCodecContext(s.decCodecContext); err != nil {
			log.Fatal(fmt.Errorf("main: updating codec context failed: %w", err))
		}
		// Open codec context
		if err := s.decCodecContext.Open(s.decCodec, nil); err != nil {
			log.Fatal(fmt.Errorf("main: opening codec context failed: %w", err))
		}
		// Add stream
		streams[is.Index()] = s
	}
	//defer s.decCodecContext.Free()
	return streams
}

func convertToImg(frame *astiav.Frame) (image.Image, error) {
	w := frame.Width()
	h := frame.Height()
	var subSampleRatio image.YCbCrSubsampleRatio
	switch frame.PixelFormat() {
	case astiav.PixelFormatYuv444P:
		subSampleRatio = image.YCbCrSubsampleRatio444
	case astiav.PixelFormatYuv422P:
		subSampleRatio = image.YCbCrSubsampleRatio422
	case astiav.PixelFormatYuv420P:
		subSampleRatio = image.YCbCrSubsampleRatio420
	case astiav.PixelFormatYuv440P:
		subSampleRatio = image.YCbCrSubsampleRatio440
	case astiav.PixelFormatYuv411P:
		subSampleRatio = image.YCbCrSubsampleRatio411
	case astiav.PixelFormatYuv410P:
		subSampleRatio = image.YCbCrSubsampleRatio410
	default:
		return nil, errors.New("unknown format")
	}

	img := &image.YCbCr{
		Y:              frame.Data()[0],
		Cb:             frame.Data()[1],
		Cr:             frame.Data()[2],
		YStride:        frame.Linesize()[0],
		CStride:        frame.Linesize()[1],
		SubsampleRatio: subSampleRatio,
		Rect:           image.Rect(0, 0, w, h),
	}
	return img, nil
}
