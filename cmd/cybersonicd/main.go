package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getlantern/systray"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	e "github.com/gzj/cybersonic/assets"
)

// -------------------------------- init --------------------------------
var (
	P        map[string]*SoundPlayer
	Sfx      map[string]io.ReadCloser
	SfxFiles []string
	logger   *slog.Logger
	logLevel *slog.LevelVar
)

func init() {
	logLevel = &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)

	P = make(map[string]*SoundPlayer)
	Sfx = make(map[string]io.ReadCloser)

	files, err := e.Sfx.ReadDir("sfx")
	if err != nil {
		slog.Error("Failed to read sfx directory", "error", err)
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		SfxFiles = append(SfxFiles, name)
		f, err := e.Sfx.Open("sfx" + "/" + name)
		if err != nil {
			slog.Error("Failed to open sfx file", "filename", name, "error", err)
			continue
		}
		Sfx[name] = f
		P[name], err = NewSoundPlayer(f)
		if err != nil {
			slog.Error("Failed to create sound player", "filename", name, "error", err)
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
		logger.Error("Failed to decode WAV file", "error", err)
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
type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeListAll
	EventTypePlaySound
)

func (et EventType) String() string {
	switch et {
	case EventTypeListAll:
		return "list_all"
	case EventTypePlaySound:
		return "play_sound"
	default:
		return "unknown"
	}
}

func HandlerAll(w http.ResponseWriter, r *http.Request) {
	logger.Info("Sound files list",
		"method", r.Method,
		"event_type", EventTypeListAll.String(),
		"remote_addr", r.RemoteAddr)

	for _, file := range SfxFiles {
		fmt.Fprintf(w, "%s\n", file)
	}
}

func HandlerSfx(w http.ResponseWriter, r *http.Request) {
	sfxName := r.URL.Query().Get("name")
	n := sfxName + ".wav"

	var payload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)

	name := ""
	var id int
	message := ""

	if err != nil {
		logger.Error("Request body decoding failed",
			"error", err,
			"sfx_name", sfxName,
			"event_type", EventTypePlaySound.String(),
			"remote_addr", r.RemoteAddr)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if nameVal, ok := payload["name"]; ok {
		name, _ = nameVal.(string)
	}

	if idVal, ok := payload["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			id = int(v)
		case int:
			id = v
		case string:
			idInt, err := strconv.Atoi(v)
			if err != nil {
				logger.Warn("Invalid ID format",
					"id", v,
					"event_type", EventTypePlaySound.String(),
					"error", err)
			} else {
				id = idInt
			}
		}
	}

	if messageVal, ok := payload["message"]; ok {
		message, _ = messageVal.(string)
	}

	logger.Debug("Sound play request details",
		"sfx_name", sfxName,
		"sound_file", n,
		"name", name,
		"id", id,
		"message", message,
		"method", r.Method,
		"event_type", EventTypePlaySound.String(),
		"remote_addr", r.RemoteAddr)

	logger.Info("Sound play request",
		"sfx_name", sfxName,
		"name", name,
		"id", id,
		"message", message,
		"event_type", EventTypePlaySound.String())

	if sound, ok := P[n]; ok {
		sound.Play()
		fmt.Fprintf(w, "Playing sound: %s\n", n)
	} else {
		logger.Warn("Sound not found",
			"sound_name", n,
			"event_type", EventTypePlaySound.String())
		http.Error(w, "Sound not found", http.StatusNotFound)
		return
	}
}

func onReady() {
	iconBytes, err := e.Icon.ReadFile("icons/tray.ico")
	if err != nil {
		logger.Error("Failed to load icon", "error", err)
		os.Exit(1)
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
	logger.Info("Stopping Cybersonicd...")
	for _, sp := range P {
		sp.Close()
	}
	os.Exit(0)
}

func main() {
	var (
		address      string
		logLevelFlag string
	)

	flag.StringVar(&address, "address", "127.0.0.1:49161", "server address")
	flag.StringVar(&logLevelFlag, "log-level", "info", "Logging level (debug, info, warn, error)")
	flag.Parse()

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	switch logLevelFlag {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		fmt.Fprintf(os.Stderr, "Invalid log level: %s. Using default (info)\n", logLevelFlag)
		logLevel.Set(slog.LevelInfo)
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(jsonHandler)

	slog.SetDefault(logger)

	go func() {
		http.HandleFunc("/sfx", HandlerSfx)
		http.HandleFunc("/all", HandlerAll)
		logger.Info("Cybersonicd is listening", "address", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	systray.Run(onReady, onExit)
}
