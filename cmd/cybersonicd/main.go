package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	e "github.com/gzj/cybersonic/assets"

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

func main() {
	var (
		address string
	)

	flag.StringVar(&address, "address", "127.0.0.1:49161", "server address")

	flag.Parse()

	http.HandleFunc("/sfx", HandlerSfx)
	http.HandleFunc("/all", HandlerAll)

	log.Println("Server listening on " + address + " ...")
	http.ListenAndServe(address, nil)
}
