package assets

import (
    "embed"
)

//go:embed sfx/*.wav
var Sfx embed.FS

//go:embed icons/*
var Icon embed.FS
