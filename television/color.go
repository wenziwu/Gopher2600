package television

// ColorSignal represents the signal that is sent from the VCS to the
// television
type ColorSignal uint16

// VideoBlack is the PixelSignal value that indicates no VCS pixel is to be
// shown - the need for the video black signal explains why ColorSignal is
// defined to be uint16 and not uint8, as you might expect
const VideoBlack ColorSignal = 0xffff

// components of the translated color signal
const (
	red = iota
	green
	blue
	colorDepth
)

// a color is made of a number of color components
type color [colorDepth]byte

// colors is the entire palette
type colors []color

// the entire palette is made up of many colors
var colorsNTSC = colors{}
var colorsPAL = colors{}
var colorsAlt = colors{}

// the VideoBlack signal results in the following color
var videoBlack = color{0, 0, 0}

// the raw color values are the component values expressed as a single 32 bit
// number. we'll use these raw values in the init() function below to create
// the real palette
var colorsNTSCRaw = []uint32{
	0x000000, 0x000000, 0x404040, 0x404040, 0x6c6c6c, 0x6c6c6c, 0x909090, 0x909090, 0xb0b0b0, 0xb0b0b0, 0xc8c8c8, 0xc8c8c8, 0xdcdcdc, 0xdcdcdc, 0xececec, 0xececec,
	0x444400, 0x444400, 0x646410, 0x646410, 0x848424, 0x848424, 0xa0a034, 0xa0a034, 0xb8b840, 0xb8b840, 0xd0d050, 0xd0d050, 0xe8e85c, 0xe8e85c, 0xfcfc68, 0xfcfc68,
	0x702800, 0x702800, 0x844414, 0x844414, 0x985c28, 0x985c28, 0xac783c, 0xac783c, 0xbc8c4c, 0xbc8c4c, 0xcca05c, 0xcca05c, 0xdcb468, 0xdcb468, 0xecc878, 0xecc878,
	0x841800, 0x841800, 0x983418, 0x983418, 0xac5030, 0xac5030, 0xc06848, 0xc06848, 0xd0805c, 0xd0805c, 0xe09470, 0xe09470, 0xeca880, 0xeca880, 0xfcbc94, 0xfcbc94,
	0x880000, 0x880000, 0x9c2020, 0x9c2020, 0xb03c3c, 0xb03c3c, 0xc05858, 0xc05858, 0xd07070, 0xd07070, 0xe08888, 0xe08888, 0xeca0a0, 0xeca0a0, 0xfcb4b4, 0xfcb4b4,
	0x78005c, 0x78005c, 0x8c2074, 0x8c2074, 0xa03c88, 0xa03c88, 0xb0589c, 0xb0589c, 0xc070b0, 0xc070b0, 0xd084c0, 0xd084c0, 0xdc9cd0, 0xdc9cd0, 0xecb0e0, 0xecb0e0,
	0x480078, 0x480078, 0x602090, 0x602090, 0x783ca4, 0x783ca4, 0x8c58b8, 0x8c58b8, 0xa070cc, 0xa070cc, 0xb484dc, 0xb484dc, 0xc49cec, 0xc49cec, 0xd4b0fc, 0xd4b0fc,
	0x140084, 0x140084, 0x302098, 0x302098, 0x4c3cac, 0x4c3cac, 0x6858c0, 0x6858c0, 0x7c70d0, 0x7c70d0, 0x9488e0, 0x9488e0, 0xa8a0ec, 0xa8a0ec, 0xbcb4fc, 0xbcb4fc,
	0x000088, 0x000088, 0x1c209c, 0x1c209c, 0x3840b0, 0x3840b0, 0x505cc0, 0x505cc0, 0x6874d0, 0x6874d0, 0x7c8ce0, 0x7c8ce0, 0x90a4ec, 0x90a4ec, 0xa4b8fc, 0xa4b8fc,
	0x00187c, 0x00187c, 0x1c3890, 0x1c3890, 0x3854a8, 0x3854a8, 0x5070bc, 0x5070bc, 0x6888cc, 0x6888cc, 0x7c9cdc, 0x7c9cdc, 0x90b4ec, 0x90b4ec, 0xa4c8fc, 0xa4c8fc,
	0x002c5c, 0x002c5c, 0x1c4c78, 0x1c4c78, 0x386890, 0x386890, 0x5084ac, 0x5084ac, 0x689cc0, 0x689cc0, 0x7cb4d4, 0x7cb4d4, 0x90cce8, 0x90cce8, 0xa4e0fc, 0xa4e0fc,
	0x003c2c, 0x003c2c, 0x1c5c48, 0x1c5c48, 0x387c64, 0x387c64, 0x509c80, 0x509c80, 0x68b494, 0x68b494, 0x7cd0ac, 0x7cd0ac, 0x90e4c0, 0x90e4c0, 0xa4fcd4, 0xa4fcd4,
	0x003c00, 0x003c00, 0x205c20, 0x205c20, 0x407c40, 0x407c40, 0x5c9c5c, 0x5c9c5c, 0x74b474, 0x74b474, 0x8cd08c, 0x8cd08c, 0xa4e4a4, 0xa4e4a4, 0xb8fcb8, 0xb8fcb8,
	0x143800, 0x143800, 0x345c1c, 0x345c1c, 0x507c38, 0x507c38, 0x6c9850, 0x6c9850, 0x84b468, 0x84b468, 0x9ccc7c, 0x9ccc7c, 0xb4e490, 0xb4e490, 0xc8fca4, 0xc8fca4,
	0x2c3000, 0x2c3000, 0x4c501c, 0x4c501c, 0x687034, 0x687034, 0x848c4c, 0x848c4c, 0x9ca864, 0x9ca864, 0xb4c078, 0xb4c078, 0xccd488, 0xccd488, 0xe0ec9c, 0xe0ec9c,
	0x442800, 0x442800, 0x644818, 0x644818, 0x846830, 0x846830, 0xa08444, 0xa08444, 0xb89c58, 0xb89c58, 0xd0b46c, 0xd0b46c, 0xe8cc7c, 0xe8cc7c, 0xfce08c, 0xfce08c,
}

var colorsPALRaw = []uint32{
	0x000000, 0x000000, 0x282828, 0x282828, 0x505050, 0x505050, 0x747474, 0x747474, 0x949494, 0x949494, 0xb4b4b4, 0xb4b4b4, 0xd0d0d0, 0xd0d0d0, 0xececed, 0xececec,
	0x000000, 0x000000, 0x282828, 0x282828, 0x505050, 0x505050, 0x747474, 0x747474, 0x949494, 0x949494, 0xb4b4b4, 0xb4b4b4, 0xd0d0d0, 0xd0d0d0, 0xececec, 0xececec,
	0x805800, 0x805800, 0x947020, 0x947020, 0xa8843c, 0xa8843c, 0xbc9c58, 0xbc9c58, 0xccac70, 0xccac70, 0xdcc084, 0xdcc084, 0xecd09c, 0xecd09c, 0xfce0b0, 0xfce0b0,
	0x445c00, 0x445c00, 0x5c7820, 0x5c7820, 0x74903c, 0x74903c, 0x8cac58, 0x8cac58, 0xa0c070, 0xa0c070, 0xb0d484, 0xb0d484, 0xc4e89c, 0xc4e89c, 0xd4fcb0, 0xd4fcb0,
	0x703400, 0x703400, 0x885020, 0x885020, 0xa0683c, 0xa0683c, 0xb48458, 0xb48458, 0xc89870, 0xc89870, 0xdcac84, 0xdcac84, 0xecc09c, 0xecc09c, 0xfcd4b0, 0xfcd4b0,
	0x006414, 0x006414, 0x208034, 0x208034, 0x3c9850, 0x3c9850, 0x58b06c, 0x58b06c, 0x70c484, 0x70c484, 0x84d89c, 0x84d89c, 0x9ce8b4, 0x9ce8b4, 0xb0fcc8, 0xb0fcc8,
	0x700014, 0x700014, 0x882034, 0x882034, 0xa03c50, 0xa03c50, 0xb4586c, 0xb4586c, 0xc87084, 0xc87084, 0xdc849c, 0xdc849c, 0xec9cb4, 0xec9cb4, 0xfcb0c8, 0xfcb0c8,
	0x005c5c, 0x005c5c, 0x207474, 0x207474, 0x3c8c8c, 0x3c8c8c, 0x58a4a4, 0x58a4a4, 0x70b8b8, 0x70b8b8, 0x84c8c8, 0x84c8c8, 0x9cdcdc, 0x9cdcdc, 0xb0ecec, 0xb0ecec,
	0x70005c, 0x70005c, 0x842074, 0x842074, 0x943c88, 0x943c88, 0xa8589c, 0xa8589c, 0xb470b0, 0xb470b0, 0xc484c0, 0xc484c0, 0xd09cd0, 0xd09cd0, 0xe0b0e0, 0xe0b0e0,
	0x003c70, 0x003c70, 0x1c5888, 0x1c5888, 0x3874a0, 0x3874a0, 0x508cb4, 0x508cb4, 0x68a4c8, 0x68a4c8, 0x7cb8dc, 0x7cb8dc, 0x90ccec, 0x90ccec, 0xa4e0fc, 0xa4e0fc,
	0x580070, 0x580070, 0x6c2088, 0x6c2088, 0x803ca0, 0x803ca0, 0x9458b4, 0x9458b4, 0xa470c8, 0xa470c8, 0xb484dc, 0xb484dc, 0xc49cec, 0xc49cec, 0xd4b0fc, 0xd4b0fc,
	0x002070, 0x002070, 0x1c3c88, 0x1c3c88, 0x3858a0, 0x3858a0, 0x5074b4, 0x5074b4, 0x6888c8, 0x6888c8, 0x7ca0dc, 0x7ca0dc, 0x90b4ec, 0x90b4ec, 0xa4c8fc, 0xa4c8fc,
	0x3c0080, 0x3c0080, 0x542094, 0x542094, 0x6c3ca8, 0x6c3ca8, 0x8058bd, 0x8058bd, 0x9470cc, 0x9470cc, 0xa884dc, 0xa884dc, 0xb89cec, 0xb89cec, 0xc8b0fc, 0xc8b0fc,
	0x000088, 0x000088, 0x20209c, 0x20209c, 0x3c3cb0, 0x3c3cb0, 0x5858c0, 0x5858c0, 0x7070d0, 0x7070d0, 0x8484e0, 0x8484e0, 0x9c9cec, 0x9c9cec, 0xb0b0fc, 0xb0b0fc,
	0x000000, 0x000000, 0x282828, 0x282828, 0x505050, 0x505050, 0x747474, 0x747474, 0x949494, 0x949494, 0xb4b4b4, 0xb4b4b4, 0xd0d0d0, 0xd0d0d0, 0xececec, 0xececec,
	0x000000, 0x000000, 0x282828, 0x282828, 0x505050, 0x505050, 0x747474, 0x747474, 0x949494, 0x949494, 0xb4b4b4, 0xb4b4b4, 0xd0d0d0, 0xd0d0d0, 0xececec, 0xececec,
}

// colors used when alternative colors are selected. these colors mirror the
// so called debug colors used by the Stella emulator
var colorsAltRaw = []uint32{
	0x333333, 0x84c8fc, 0x9246c0, 0x901c00, 0xe8e84a, 0xd5824a, 0x328432,
}

// as used in Stella, the colors above are used as below. note that using a
// alt pixel ColorSignal not in the following list may result in a panic
const (
	AltColBackground = iota
	AltColBall
	AltColPlayfield
	AltColPlayer0
	AltColPlayer1
	AltColMissile0
	AltColMissile1
)

func init() {
	for i := range colorsNTSCRaw {
		col := colorsNTSCRaw[i]
		red, green, blue := byte((col&0xff0000)>>16), byte((col&0xff00)>>8), byte(col&0xff)
		colorsNTSC = append(colorsNTSC, color{red, green, blue})
	}

	for i := range colorsPALRaw {
		col := colorsPALRaw[i]
		red, green, blue := byte((col&0xff0000)>>16), byte((col&0xff00)>>8), byte(col&0xff)
		colorsPAL = append(colorsPAL, color{red, green, blue})
	}

	for i := range colorsAltRaw {
		col := colorsAltRaw[i]
		red, green, blue := byte((col&0xff0000)>>16), byte((col&0xff00)>>8), byte(col&0xff)
		colorsAlt = append(colorsAlt, color{red, green, blue})
	}
}

// getColor translates a color signal to the individual color components
func getColor(spec *Specification, sig ColorSignal) (byte, byte, byte) {
	if sig == VideoBlack {
		return videoBlack[red], videoBlack[green], videoBlack[blue]
	}
	col := spec.Colors[sig]
	return col[red], col[green], col[blue]
}

// getColor translates a color signal to the individual color components
func getAltColor(sig ColorSignal) (byte, byte, byte) {
	if sig == VideoBlack {
		return videoBlack[red], videoBlack[green], videoBlack[blue]
	}

	if int(sig) >= len(colorsAlt) {
		panic("alt color signal too big")
	}

	col := colorsAlt[sig]
	return col[red], col[green], col[blue]
}
