package assets

import (
    "embed"
)

//go:embed sfx/*.wav
var Sfx embed.FS
