package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	e "github.com/gzj/cybersonic/assets"

	"github.com/getlantern/systray"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

// -------------------------------- init --------------------------------
var P map[string]*SoundPlayer
var Sfx map[string]io.ReadCloser
var SfxFiles []string

func init() {
	P = make(map[string]*SoundPlayer)
	Sfx = make(map[string]io.ReadCloser)
	files, err := e.Sfx.ReadDir("sfx")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		SfxFiles = append(SfxFiles, name)
		f, err := e.Sfx.Open("sfx" + "/" + name)
		if err != nil {
			log.Println(err)
		}
		Sfx[name] = f
		P[name], err = NewSoundPlayer(f)
		if err != nil {
			log.Println(err)
		}
	}
}

// -------------------------------- sound player --------------------------------
type SoundPlayer struct {
	reader   io.ReadCloser
	streamer beep.StreamSeekCloser
	buffer   *beep.Buffer
}

func NewSoundPlayer(r io.ReadCloser) (*SoundPlayer, error) {
	streamer, format, err := wav.Decode(r)
	if err != nil {
		log.Println(err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	return &SoundPlayer{
		reader:   r,
		streamer: streamer,
		buffer:   buffer,
	}, nil
}

func (sp *SoundPlayer) Play() {
	s := sp.buffer.Streamer(0, sp.buffer.Len())
	speaker.Play(s)
}

func (sp *SoundPlayer) Close() {
	sp.streamer.Close()
	sp.reader.Close()
}

// -------------------------------- server --------------------------------
func HandlerAll(w http.ResponseWriter, r *http.Request) {
	for _, file := range SfxFiles {
		fmt.Fprintf(w, "%s\n", file)
	}
}
func HandlerSfx(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	name := values.Get("name")
	n := name + ".wav"

	if sound, ok := P[n]; ok {
		sound.Play()
	} else {
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}
}

func onReady() {
	iconBytes, err := e.Icon.ReadFile("icons/tray.ico")
	if err != nil {
		log.Fatalf("Failed to load icon: %v", err)
	}

	systray.SetIcon(iconBytes)
	systray.SetTitle("cybersonicd")
	systray.SetTooltip("cybersonicd")

	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	log.Println("Stopping Cybersonicd...")
	for _, sp := range P {
		sp.Close()
	}
	os.Exit(0)
}

func main() {
	var (
		address string
	)
	flag.StringVar(&address, "address", "127.0.0.1:49161", "server address")
	flag.Parse()

	go func() {
		http.HandleFunc("/sfx", HandlerSfx)
		http.HandleFunc("/all", HandlerAll)
		log.Printf("Cybersonicd is listening at http://%s ...\n", address)
		http.ListenAndServe(address, nil)
	}()

	systray.Run(onReady, onExit)
}
