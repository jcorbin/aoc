package ansi

// Color definitions taken from https://en.wikipedia.org/wiki/ANSI_escape_code#Colors.

// Palette is a limited palette of color for legacy terminals.
type Palette []SGRColor

// Palette3 is the classic 3-bit palette of 8 colors.
var Palette3 = Palette{
	SGRBlack,
	SGRRed,
	SGRGreen,
	SGRYellow,
	SGRBlue,
	SGRMagenta,
	SGRCyan,
	SGRWhite,
}

// Palette4 is the extended 4-bit palette of the 8 classic colors and their
// bright counterparts.
var Palette4 = Palette3.concat(
	SGRBrightBlack,
	SGRBrightRed,
	SGRBrightGreen,
	SGRBrightYellow,
	SGRBrightBlue,
	SGRBrightMagenta,
	SGRBrightCyan,
	SGRBrightWhite,
)

// Palette8 is the extended 8-bit palette of the first 16 extended colors, a
// 6x6x6=216 color cube, and 24 shades of gray.
var Palette8 = Palette4.concat(
	// plane 1, row 1
	SGRCube16, SGRCube17, SGRCube18, SGRCube19, SGRCube20, SGRCube21,
	SGRCube22, SGRCube23, SGRCube24, SGRCube25, SGRCube26, SGRCube27,
	SGRCube28, SGRCube29, SGRCube30, SGRCube31, SGRCube32, SGRCube33,
	SGRCube34, SGRCube35, SGRCube36, SGRCube37, SGRCube38, SGRCube39,
	SGRCube40, SGRCube41, SGRCube42, SGRCube43, SGRCube44, SGRCube45,
	SGRCube46, SGRCube47, SGRCube48, SGRCube49, SGRCube50, SGRCube51,

	// plane 2, row 1
	SGRCube52, SGRCube53, SGRCube54, SGRCube55, SGRCube56, SGRCube57,
	SGRCube58, SGRCube59, SGRCube60, SGRCube61, SGRCube62, SGRCube63,
	SGRCube64, SGRCube65, SGRCube66, SGRCube67, SGRCube68, SGRCube69,
	SGRCube70, SGRCube71, SGRCube72, SGRCube73, SGRCube74, SGRCube75,
	SGRCube76, SGRCube77, SGRCube78, SGRCube79, SGRCube80, SGRCube81,
	SGRCube82, SGRCube83, SGRCube84, SGRCube85, SGRCube86, SGRCube87,

	// plane 3, row 1
	SGRCube88, SGRCube89, SGRCube90, SGRCube91, SGRCube92, SGRCube93,
	SGRCube94, SGRCube95, SGRCube96, SGRCube97, SGRCube98, SGRCube99,
	SGRCube100, SGRCube101, SGRCube102, SGRCube103, SGRCube104, SGRCube105,
	SGRCube106, SGRCube107, SGRCube108, SGRCube109, SGRCube110, SGRCube111,
	SGRCube112, SGRCube113, SGRCube114, SGRCube115, SGRCube116, SGRCube117,
	SGRCube118, SGRCube119, SGRCube120, SGRCube121, SGRCube122, SGRCube123,

	// plane 4, row 1
	SGRCube124, SGRCube125, SGRCube126, SGRCube127, SGRCube128, SGRCube129,
	SGRCube130, SGRCube131, SGRCube132, SGRCube133, SGRCube134, SGRCube135,
	SGRCube136, SGRCube137, SGRCube138, SGRCube139, SGRCube140, SGRCube141,
	SGRCube142, SGRCube143, SGRCube144, SGRCube145, SGRCube146, SGRCube147,
	SGRCube148, SGRCube149, SGRCube150, SGRCube151, SGRCube152, SGRCube153,
	SGRCube154, SGRCube155, SGRCube156, SGRCube157, SGRCube158, SGRCube159,

	// plane 5, row 1
	SGRCube160, SGRCube161, SGRCube162, SGRCube163, SGRCube164, SGRCube165,
	SGRCube166, SGRCube167, SGRCube168, SGRCube169, SGRCube170, SGRCube171,
	SGRCube172, SGRCube173, SGRCube174, SGRCube175, SGRCube176, SGRCube177,
	SGRCube178, SGRCube179, SGRCube180, SGRCube181, SGRCube182, SGRCube183,
	SGRCube184, SGRCube185, SGRCube186, SGRCube187, SGRCube188, SGRCube189,
	SGRCube190, SGRCube191, SGRCube192, SGRCube193, SGRCube194, SGRCube195,

	// plane 6, row 1
	SGRCube196, SGRCube197, SGRCube198, SGRCube199, SGRCube200, SGRCube201,
	SGRCube202, SGRCube203, SGRCube204, SGRCube205, SGRCube206, SGRCube207,
	SGRCube208, SGRCube209, SGRCube210, SGRCube211, SGRCube212, SGRCube213,
	SGRCube214, SGRCube215, SGRCube216, SGRCube217, SGRCube218, SGRCube219,
	SGRCube220, SGRCube221, SGRCube222, SGRCube223, SGRCube224, SGRCube225,
	SGRCube226, SGRCube227, SGRCube228, SGRCube229, SGRCube230, SGRCube231,

	// Grayscale colors
	SGRGray1, SGRGray2, SGRGray3, SGRGray4, SGRGray5, SGRGray6,
	SGRGray7, SGRGray8, SGRGray9, SGRGray10, SGRGray11, SGRGray12,
	SGRGray13, SGRGray14, SGRGray15, SGRGray16, SGRGray17, SGRGray18,
	SGRGray19, SGRGray20, SGRGray21, SGRGray22, SGRGray23, SGRGray24,
)

// Palette3Colors is the canonical 24-bit definitions for the classic 3-bit
// palette of 8 colors.
var Palette3Colors = Palette{
	RGB(0x00, 0x00, 0x00), // SGRBlack
	RGB(0x80, 0x00, 0x00), // SGRRed
	RGB(0x00, 0x80, 0x00), // SGRGreen
	RGB(0x80, 0x80, 0x00), // SGRYellow
	RGB(0x00, 0x00, 0x80), // SGRBlue
	RGB(0x80, 0x00, 0x80), // SGRMagenta
	RGB(0x00, 0x80, 0x80), // SGRCyan
	RGB(0xC0, 0xC0, 0xC0), // SGRWhite
}

// Palette4Colors is the canonical 24-bit definitions for the extended 4-bit
// palette of 16 colors.
var Palette4Colors = Palette3Colors.concat(
	RGB(0x80, 0x80, 0x80), // SGRBrightBlack
	RGB(0xFF, 0x00, 0x00), // SGRBrightRed
	RGB(0x00, 0xFF, 0x00), // SGRBrightGreen
	RGB(0xFF, 0xFF, 0x00), // SGRBrightYellow
	RGB(0x00, 0x00, 0xFF), // SGRBrightBlue
	RGB(0xFF, 0x00, 0xFF), // SGRBrightMagenta
	RGB(0x00, 0xFF, 0xFF), // SGRBrightCyan
	RGB(0xFF, 0xFF, 0xFF), // SGRBrightWhite
)

// Palette8Colors is the canonical 24-bit definitions for the extended 8-bit
// palette of 256 colors
var Palette8Colors = Palette4Colors.concat(
	// plane 1, row 1
	RGB(0x00, 0x00, 0x00), // SGRCube16
	RGB(0x00, 0x00, 0x5F), // SGRCube17
	RGB(0x00, 0x00, 0x87), // SGRCube18
	RGB(0x00, 0x00, 0xAF), // SGRCube19
	RGB(0x00, 0x00, 0xD7), // SGRCube20
	RGB(0x00, 0x00, 0xFF), // SGRCube21

	// plane 1, row 2
	RGB(0x00, 0x5F, 0x00), // SGRCube22
	RGB(0x00, 0x5F, 0x5F), // SGRCube23
	RGB(0x00, 0x5F, 0x87), // SGRCube24
	RGB(0x00, 0x5F, 0xAF), // SGRCube25
	RGB(0x00, 0x5F, 0xD7), // SGRCube26
	RGB(0x00, 0x5F, 0xFF), // SGRCube27

	// plane 1, row 3
	RGB(0x00, 0x87, 0x00), // SGRCube28
	RGB(0x00, 0x87, 0x5F), // SGRCube29
	RGB(0x00, 0x87, 0x87), // SGRCube30
	RGB(0x00, 0x87, 0xAF), // SGRCube31
	RGB(0x00, 0x87, 0xD7), // SGRCube32
	RGB(0x00, 0x87, 0xFF), // SGRCube33

	// plane 1, row 4
	RGB(0x00, 0xAF, 0x00), // SGRCube34
	RGB(0x00, 0xAF, 0x5F), // SGRCube35
	RGB(0x00, 0xAF, 0x87), // SGRCube36
	RGB(0x00, 0xAF, 0xAF), // SGRCube37
	RGB(0x00, 0xAF, 0xD7), // SGRCube38
	RGB(0x00, 0xAF, 0xFF), // SGRCube39

	// plane 1, row 5
	RGB(0x00, 0xD7, 0x00), // SGRCube40
	RGB(0x00, 0xD7, 0x5F), // SGRCube41
	RGB(0x00, 0xD7, 0x87), // SGRCube42
	RGB(0x00, 0xD7, 0xAF), // SGRCube43
	RGB(0x00, 0xD7, 0xD7), // SGRCube44
	RGB(0x00, 0xD7, 0xFF), // SGRCube45

	// plane 1, row 6
	RGB(0x00, 0xFF, 0x00), // SGRCube46
	RGB(0x00, 0xFF, 0x5F), // SGRCube47
	RGB(0x00, 0xFF, 0x87), // SGRCube48
	RGB(0x00, 0xFF, 0xAF), // SGRCube49
	RGB(0x00, 0xFF, 0xD7), // SGRCube50
	RGB(0x00, 0xFF, 0xFF), // SGRCube51

	// plane 2, row 1
	RGB(0x5F, 0x00, 0x00), // SGRCube52
	RGB(0x5F, 0x00, 0x5F), // SGRCube53
	RGB(0x5F, 0x00, 0x87), // SGRCube54
	RGB(0x5F, 0x00, 0xAF), // SGRCube55
	RGB(0x5F, 0x00, 0xD7), // SGRCube56
	RGB(0x5F, 0x00, 0xFF), // SGRCube57

	// plane 2, row 2
	RGB(0x5F, 0x5F, 0x00), // SGRCube58
	RGB(0x5F, 0x5F, 0x5F), // SGRCube59
	RGB(0x5F, 0x5F, 0x87), // SGRCube60
	RGB(0x5F, 0x5F, 0xAF), // SGRCube61
	RGB(0x5F, 0x5F, 0xD7), // SGRCube62
	RGB(0x5F, 0x5F, 0xFF), // SGRCube63

	// plane 2, row 3
	RGB(0x5F, 0x87, 0x00), // SGRCube64
	RGB(0x5F, 0x87, 0x5F), // SGRCube65
	RGB(0x5F, 0x87, 0x87), // SGRCube66
	RGB(0x5F, 0x87, 0xAF), // SGRCube67
	RGB(0x5F, 0x87, 0xD7), // SGRCube68
	RGB(0x5F, 0x87, 0xFF), // SGRCube69

	// plane 2, row 4
	RGB(0x5F, 0xAF, 0x00), // SGRCube70
	RGB(0x5F, 0xAF, 0x5F), // SGRCube71
	RGB(0x5F, 0xAF, 0x87), // SGRCube72
	RGB(0x5F, 0xAF, 0xAF), // SGRCube73
	RGB(0x5F, 0xAF, 0xD7), // SGRCube74
	RGB(0x5F, 0xAF, 0xFF), // SGRCube75

	// plane 2, row 5
	RGB(0x5F, 0xD7, 0x00), // SGRCube76
	RGB(0x5F, 0xD7, 0x5F), // SGRCube77
	RGB(0x5F, 0xD7, 0x87), // SGRCube78
	RGB(0x5F, 0xD7, 0xAF), // SGRCube79
	RGB(0x5F, 0xD7, 0xD7), // SGRCube80
	RGB(0x5F, 0xD7, 0xFF), // SGRCube81

	// plane 2, row 6
	RGB(0x5F, 0xFF, 0x00), // SGRCube82
	RGB(0x5F, 0xFF, 0x5F), // SGRCube83
	RGB(0x5F, 0xFF, 0x87), // SGRCube84
	RGB(0x5F, 0xFF, 0xAF), // SGRCube85
	RGB(0x5F, 0xFF, 0xD7), // SGRCube86
	RGB(0x5F, 0xFF, 0xFF), // SGRCube87

	// plane 3, row 1
	RGB(0x87, 0x00, 0x00), // SGRCube88
	RGB(0x87, 0x00, 0x5F), // SGRCube89
	RGB(0x87, 0x00, 0x87), // SGRCube90
	RGB(0x87, 0x00, 0xAF), // SGRCube91
	RGB(0x87, 0x00, 0xD7), // SGRCube92
	RGB(0x87, 0x00, 0xFF), // SGRCube93

	// plane 3, row 2
	RGB(0x87, 0x5F, 0x00), // SGRCube94
	RGB(0x87, 0x5F, 0x5F), // SGRCube95
	RGB(0x87, 0x5F, 0x87), // SGRCube96
	RGB(0x87, 0x5F, 0xAF), // SGRCube97
	RGB(0x87, 0x5F, 0xD7), // SGRCube98
	RGB(0x87, 0x5F, 0xFF), // SGRCube99

	// plane 3, row 3
	RGB(0x87, 0x87, 0x00), // SGRCube100
	RGB(0x87, 0x87, 0x5F), // SGRCube101
	RGB(0x87, 0x87, 0x87), // SGRCube102
	RGB(0x87, 0x87, 0xAF), // SGRCube103
	RGB(0x87, 0x87, 0xD7), // SGRCube104
	RGB(0x87, 0x87, 0xFF), // SGRCube105

	// plane 3, row 4
	RGB(0x87, 0xAF, 0x00), // SGRCube106
	RGB(0x87, 0xAF, 0x5F), // SGRCube107
	RGB(0x87, 0xAF, 0x87), // SGRCube108
	RGB(0x87, 0xAF, 0xAF), // SGRCube109
	RGB(0x87, 0xAF, 0xD7), // SGRCube110
	RGB(0x87, 0xAF, 0xFF), // SGRCube111

	// plane 3, row 5
	RGB(0x87, 0xD7, 0x00), // SGRCube112
	RGB(0x87, 0xD7, 0x5F), // SGRCube113
	RGB(0x87, 0xD7, 0x87), // SGRCube114
	RGB(0x87, 0xD7, 0xAF), // SGRCube115
	RGB(0x87, 0xD7, 0xD7), // SGRCube116
	RGB(0x87, 0xD7, 0xFF), // SGRCube117

	// plane 3, row 6
	RGB(0x87, 0xFF, 0x00), // SGRCube118
	RGB(0x87, 0xFF, 0x5F), // SGRCube119
	RGB(0x87, 0xFF, 0x87), // SGRCube120
	RGB(0x87, 0xFF, 0xAF), // SGRCube121
	RGB(0x87, 0xFF, 0xD7), // SGRCube122
	RGB(0x87, 0xFF, 0xFF), // SGRCube123

	// plane 4, row 1
	RGB(0xAF, 0x00, 0x00), // SGRCube124
	RGB(0xAF, 0x00, 0x5F), // SGRCube125
	RGB(0xAF, 0x00, 0x87), // SGRCube126
	RGB(0xAF, 0x00, 0xAF), // SGRCube127
	RGB(0xAF, 0x00, 0xD7), // SGRCube128
	RGB(0xAF, 0x00, 0xFF), // SGRCube129

	// plane 4, row 2
	RGB(0xAF, 0x5F, 0x00), // SGRCube130
	RGB(0xAF, 0x5F, 0x5F), // SGRCube131
	RGB(0xAF, 0x5F, 0x87), // SGRCube132
	RGB(0xAF, 0x5F, 0xAF), // SGRCube133
	RGB(0xAF, 0x5F, 0xD7), // SGRCube134
	RGB(0xAF, 0x5F, 0xFF), // SGRCube135

	// plane 4, row 3
	RGB(0xAF, 0x87, 0x00), // SGRCube136
	RGB(0xAF, 0x87, 0x5F), // SGRCube137
	RGB(0xAF, 0x87, 0x87), // SGRCube138
	RGB(0xAF, 0x87, 0xAF), // SGRCube139
	RGB(0xAF, 0x87, 0xD7), // SGRCube140
	RGB(0xAF, 0x87, 0xFF), // SGRCube141

	// plane 4, row 4
	RGB(0xAF, 0xAF, 0x00), // SGRCube142
	RGB(0xAF, 0xAF, 0x5F), // SGRCube143
	RGB(0xAF, 0xAF, 0x87), // SGRCube144
	RGB(0xAF, 0xAF, 0xAF), // SGRCube145
	RGB(0xAF, 0xAF, 0xD7), // SGRCube146
	RGB(0xAF, 0xAF, 0xFF), // SGRCube147

	// plane 4, row 5
	RGB(0xAF, 0xD7, 0x00), // SGRCube148
	RGB(0xAF, 0xD7, 0x5F), // SGRCube149
	RGB(0xAF, 0xD7, 0x87), // SGRCube150
	RGB(0xAF, 0xD7, 0xAF), // SGRCube151
	RGB(0xAF, 0xD7, 0xD7), // SGRCube152
	RGB(0xAF, 0xD7, 0xFF), // SGRCube153

	// plane 4, row 6
	RGB(0xAF, 0xFF, 0x00), // SGRCube154
	RGB(0xAF, 0xFF, 0x5F), // SGRCube155
	RGB(0xAF, 0xFF, 0x87), // SGRCube156
	RGB(0xAF, 0xFF, 0xAF), // SGRCube157
	RGB(0xAF, 0xFF, 0xD7), // SGRCube158
	RGB(0xAF, 0xFF, 0xFF), // SGRCube159

	// plane 5, row 1
	RGB(0xD7, 0x00, 0x00), // SGRCube160
	RGB(0xD7, 0x00, 0x5F), // SGRCube161
	RGB(0xD7, 0x00, 0x87), // SGRCube162
	RGB(0xD7, 0x00, 0xAF), // SGRCube163
	RGB(0xD7, 0x00, 0xD7), // SGRCube164
	RGB(0xD7, 0x00, 0xFF), // SGRCube165

	// plane 5, row 2
	RGB(0xD7, 0x5F, 0x00), // SGRCube166
	RGB(0xD7, 0x5F, 0x5F), // SGRCube167
	RGB(0xD7, 0x5F, 0x87), // SGRCube168
	RGB(0xD7, 0x5F, 0xAF), // SGRCube169
	RGB(0xD7, 0x5F, 0xD7), // SGRCube170
	RGB(0xD7, 0x5F, 0xFF), // SGRCube171

	// plane 5, row 3
	RGB(0xD7, 0x87, 0x00), // SGRCube172
	RGB(0xD7, 0x87, 0x5F), // SGRCube173
	RGB(0xD7, 0x87, 0x87), // SGRCube174
	RGB(0xD7, 0x87, 0xAF), // SGRCube175
	RGB(0xD7, 0x87, 0xD7), // SGRCube176
	RGB(0xD7, 0x87, 0xFF), // SGRCube177

	// plane 5, row 4
	RGB(0xD7, 0xAF, 0x00), // SGRCube178
	RGB(0xD7, 0xAF, 0x5F), // SGRCube179
	RGB(0xD7, 0xAF, 0x87), // SGRCube180
	RGB(0xD7, 0xAF, 0xAF), // SGRCube181
	RGB(0xD7, 0xAF, 0xD7), // SGRCube182
	RGB(0xD7, 0xAF, 0xFF), // SGRCube183

	// plane 5, row 5
	RGB(0xD7, 0xD7, 0x00), // SGRCube184
	RGB(0xD7, 0xD7, 0x5F), // SGRCube185
	RGB(0xD7, 0xD7, 0x87), // SGRCube186
	RGB(0xD7, 0xD7, 0xAF), // SGRCube187
	RGB(0xD7, 0xD7, 0xD7), // SGRCube188
	RGB(0xD7, 0xD7, 0xFF), // SGRCube189

	// plane 5, row 6
	RGB(0xD7, 0xFF, 0x00), // SGRCube190
	RGB(0xD7, 0xFF, 0x5F), // SGRCube191
	RGB(0xD7, 0xFF, 0x87), // SGRCube192
	RGB(0xD7, 0xFF, 0xAF), // SGRCube193
	RGB(0xD7, 0xFF, 0xD7), // SGRCube194
	RGB(0xD7, 0xFF, 0xFF), // SGRCube195

	// plane 6, row 1
	RGB(0xFF, 0x00, 0x00), // SGRCube196
	RGB(0xFF, 0x00, 0x5F), // SGRCube197
	RGB(0xFF, 0x00, 0x87), // SGRCube198
	RGB(0xFF, 0x00, 0xAF), // SGRCube199
	RGB(0xFF, 0x00, 0xD7), // SGRCube200
	RGB(0xFF, 0x00, 0xFF), // SGRCube201

	// plane 6, row 2
	RGB(0xFF, 0x5F, 0x00), // SGRCube202
	RGB(0xFF, 0x5F, 0x5F), // SGRCube203
	RGB(0xFF, 0x5F, 0x87), // SGRCube204
	RGB(0xFF, 0x5F, 0xAF), // SGRCube205
	RGB(0xFF, 0x5F, 0xD7), // SGRCube206
	RGB(0xFF, 0x5F, 0xFF), // SGRCube207

	// plane 6, row 3
	RGB(0xFF, 0x87, 0x00), // SGRCube208
	RGB(0xFF, 0x87, 0x5F), // SGRCube209
	RGB(0xFF, 0x87, 0x87), // SGRCube210
	RGB(0xFF, 0x87, 0xAF), // SGRCube211
	RGB(0xFF, 0x87, 0xD7), // SGRCube212
	RGB(0xFF, 0x87, 0xFF), // SGRCube213

	// plane 6, row 4
	RGB(0xFF, 0xAF, 0x00), // SGRCube214
	RGB(0xFF, 0xAF, 0x5F), // SGRCube215
	RGB(0xFF, 0xAF, 0x87), // SGRCube216
	RGB(0xFF, 0xAF, 0xAF), // SGRCube217
	RGB(0xFF, 0xAF, 0xD7), // SGRCube218
	RGB(0xFF, 0xAF, 0xFF), // SGRCube219

	// plane 6, row 5
	RGB(0xFF, 0xD7, 0x00), // SGRCube220
	RGB(0xFF, 0xD7, 0x5F), // SGRCube221
	RGB(0xFF, 0xD7, 0x87), // SGRCube222
	RGB(0xFF, 0xD7, 0xAF), // SGRCube223
	RGB(0xFF, 0xD7, 0xD7), // SGRCube224
	RGB(0xFF, 0xD7, 0xFF), // SGRCube225

	// plane 6, row 6
	RGB(0xFF, 0xFF, 0x00), // SGRCube226
	RGB(0xFF, 0xFF, 0x5F), // SGRCube227
	RGB(0xFF, 0xFF, 0x87), // SGRCube228
	RGB(0xFF, 0xFF, 0xAF), // SGRCube229
	RGB(0xFF, 0xFF, 0xD7), // SGRCube230
	RGB(0xFF, 0xFF, 0xFF), // SGRCube231

	// Grayscale colors
	RGB(0x08, 0x08, 0x08), // SGRGray1
	RGB(0x12, 0x12, 0x12), // SGRGray2
	RGB(0x1C, 0x1C, 0x1C), // SGRGray3
	RGB(0x26, 0x26, 0x26), // SGRGray4
	RGB(0x30, 0x30, 0x30), // SGRGray5
	RGB(0x3A, 0x3A, 0x3A), // SGRGray6
	RGB(0x44, 0x44, 0x44), // SGRGray7
	RGB(0x4E, 0x4E, 0x4E), // SGRGray8
	RGB(0x58, 0x58, 0x58), // SGRGray9
	RGB(0x62, 0x62, 0x62), // SGRGray10
	RGB(0x6C, 0x6C, 0x6C), // SGRGray11
	RGB(0x76, 0x76, 0x76), // SGRGray12
	RGB(0x80, 0x80, 0x80), // SGRGray13
	RGB(0x8A, 0x8A, 0x8A), // SGRGray14
	RGB(0x94, 0x94, 0x94), // SGRGray15
	RGB(0x9E, 0x9E, 0x9E), // SGRGray16
	RGB(0xA8, 0xA8, 0xA8), // SGRGray17
	RGB(0xB2, 0xB2, 0xB2), // SGRGray18
	RGB(0xBC, 0xBC, 0xBC), // SGRGray19
	RGB(0xC6, 0xC6, 0xC6), // SGRGray20
	RGB(0xD0, 0xD0, 0xD0), // SGRGray21
	RGB(0xDA, 0xDA, 0xDA), // SGRGray22
	RGB(0xE4, 0xE4, 0xE4), // SGRGray23
	RGB(0xEE, 0xEE, 0xEE), // SGRGray24
)

// ColorModel implements an SGR color model.
type ColorModel interface {
	Convert(c SGRColor) SGRColor
}

// ColorModelFunc is a convenient way to implement to implement simple SGR
// color models.
type ColorModelFunc func(c SGRColor) SGRColor

// Convert calls the aliased function.
func (f ColorModelFunc) Convert(c SGRColor) SGRColor { return f(c) }

// SGRColorMap implements an SGR ColorModel around a map; any colors not in the
// map are passed through.
type SGRColorMap map[SGRColor]SGRColor

// SGRColorCanon implements a canonical mapping from 24-bit colors back to
// their 3, 4, and 8-bit palette aliases.
var SGRColorCanon = make(SGRColorMap, len(Palette8))

func init() {
	for i := 0; i < len(Palette8); i++ {
		SGRColorCanon[Palette8Colors[i]] = Palette8[i]
	}
}

// Convert a color through the map, passing through any misses.
func (cm SGRColorMap) Convert(c SGRColor) SGRColor {
	if mc, def := cm[c]; def {
		return mc
	}
	return c
}

// ColorModelID is the identity color model.
var ColorModelID = ColorModelFunc(func(c SGRColor) SGRColor { return c })

// ColorModel24 upgrades colors to their 24-bit default definitions.
var ColorModel24 = ColorModelFunc(SGRColor.To24Bit)

// ColorTheme is a Palette for the first N (usually 16) colors; its conversion
// falls back to the normal 8-bit palette.
type ColorTheme Palette

// Convert returns the theme color definition, or its 8-bit palette definition,
// if the color is not already 24-bit color.
func (theme ColorTheme) Convert(c SGRColor) SGRColor {
	if c&sgrColor24 != 0 {
		return c
	}
	c &= 0xff
	if int(c) < len(theme) {
		return theme[c]
	}
	return Palette8[c]
}

// VGAPalette is the classic VGA color theme.
var VGAPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xAA, 0x00, 0x00),
	RGB(0x00, 0xAA, 0x00),
	RGB(0xAA, 0x55, 0x00),
	RGB(0x00, 0x00, 0xAA),
	RGB(0xAA, 0x00, 0xAA),
	RGB(0x00, 0xAA, 0xAA),
	RGB(0xAA, 0xAA, 0xAA),
	RGB(0x55, 0x55, 0x55),
	RGB(0xFF, 0x55, 0x55),
	RGB(0x55, 0xFF, 0x55),
	RGB(0xFF, 0xFF, 0x55),
	RGB(0x55, 0x55, 0xFF),
	RGB(0xFF, 0x55, 0xFF),
	RGB(0x55, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// CMDPalette is the color theme used by Windows cmd.exe.
var CMDPalette = ColorTheme{
	RGB(0x01, 0x01, 0x01),
	RGB(0x80, 0x00, 0x00),
	RGB(0x00, 0x80, 0x00),
	RGB(0x80, 0x80, 0x00),
	RGB(0x00, 0x00, 0x80),
	RGB(0x80, 0x00, 0x80),
	RGB(0x00, 0x80, 0x80),
	RGB(0xC0, 0xC0, 0xC0),
	RGB(0x80, 0x80, 0x80),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// TermnialAppPalette is the color theme used by Mac Terminal.App.
var TermnialAppPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xC2, 0x36, 0x21),
	RGB(0x25, 0xBC, 0x24),
	RGB(0xAD, 0xAD, 0x27),
	RGB(0x49, 0x2E, 0xE1),
	RGB(0xD3, 0x38, 0xD3),
	RGB(0x33, 0xBB, 0xC8),
	RGB(0xCB, 0xCC, 0xCD),
	RGB(0x81, 0x83, 0x83),
	RGB(0xFC, 0x39, 0x1F),
	RGB(0x31, 0xE7, 0x22),
	RGB(0xEA, 0xEC, 0x23),
	RGB(0x58, 0x33, 0xFF),
	RGB(0xF9, 0x35, 0xF8),
	RGB(0x14, 0xF0, 0xF0),
	RGB(0xE9, 0xEB, 0xEB),
}

// PuTTYPalette is the color theme used by PuTTY.
var PuTTYPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xBB, 0x00, 0x00),
	RGB(0x00, 0xBB, 0x00),
	RGB(0xBB, 0xBB, 0x00),
	RGB(0x00, 0x00, 0xBB),
	RGB(0xBB, 0x00, 0xBB),
	RGB(0x00, 0xBB, 0xBB),
	RGB(0xBB, 0xBB, 0xBB),
	RGB(0x55, 0x55, 0x55),
	RGB(0xFF, 0x55, 0x55),
	RGB(0x55, 0xFF, 0x55),
	RGB(0xFF, 0xFF, 0x55),
	RGB(0x55, 0x55, 0xFF),
	RGB(0xFF, 0x55, 0xFF),
	RGB(0x55, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// MIRCPalette is the color theme used by mIRC.
var MIRCPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0x7F, 0x00, 0x00),
	RGB(0x00, 0x93, 0x00),
	RGB(0xFC, 0x7F, 0x00),
	RGB(0x00, 0x00, 0x7F),
	RGB(0x9C, 0x00, 0x9C),
	RGB(0x00, 0x93, 0x93),
	RGB(0xD2, 0xD2, 0xD2),
	RGB(0x7F, 0x7F, 0x7F),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFC, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFC),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// XTermPalette is the color theme used by xTerm.
var XTermPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xCD, 0x00, 0x00),
	RGB(0x00, 0xCD, 0x00),
	RGB(0xCD, 0xCD, 0x00),
	RGB(0x00, 0x00, 0xEE),
	RGB(0xCD, 0x00, 0xCD),
	RGB(0x00, 0xCD, 0xCD),
	RGB(0xE5, 0xE5, 0xE5),
	RGB(0x7F, 0x7F, 0x7F),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x5C, 0x5C, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// XPalette is the color theme used by X.
var XPalette = ColorTheme{
	RGB(0x00, 0x00, 0x00),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
	RGB(0x80, 0x80, 0x80),
	RGB(0xFF, 0x00, 0x00),
	RGB(0x90, 0xEE, 0x90),
	RGB(0xFF, 0xFF, 0xE0),
	RGB(0xAD, 0xD8, 0xE6),
	RGB(0xFF, 0x00, 0xFF),
	RGB(0xE0, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

// UbuntuPalette is the color theme used by Ubuntu.
var UbuntuPalette = ColorTheme{
	RGB(0xDE, 0x38, 0x2B),
	RGB(0x39, 0xB5, 0x4A),
	RGB(0xFF, 0xC7, 0x06),
	RGB(0x00, 0x6F, 0xB8),
	RGB(0x76, 0x26, 0x71),
	RGB(0x2C, 0xB5, 0xE9),
	RGB(0xCC, 0xCC, 0xCC),
	RGB(0xFF, 0xFF, 0xFF),
	RGB(0x80, 0x80, 0x80),
	RGB(0x00, 0xFF, 0x00),
	RGB(0xFF, 0xFF, 0x00),
	RGB(0x00, 0x00, 0xFF),
	RGB(0xAD, 0xD8, 0xE6),
	RGB(0x00, 0xFF, 0xFF),
	RGB(0xE0, 0xFF, 0xFF),
	RGB(0xFF, 0xFF, 0xFF),
}

func (p Palette) concat(colors ...SGRColor) Palette {
	return append(p[:len(p):len(p)], colors...)
}

// Convert returns the palette color closest to c in Euclidean R,G,B space.
func (p Palette) Convert(c SGRColor) SGRColor {
	if len(p) == 0 {
		return SGRBlack
	}
	return p[p.Index(c)]
}

// Index returns the index of the palette color closest to c in Euclidean R,G,B
// space.
func (p Palette) Index(c SGRColor) int {
	cr, cg, cb := c.RGB()
	ret, bestSum := 0, uint32(1<<32-1)
	for i := range p {
		pr, pg, pb := p[i].RGB()
		if sum := sqDiff(cr, pr) + sqDiff(cg, pg) + sqDiff(cb, pb); sum < bestSum {
			if sum == 0 {
				return i
			}
			ret, bestSum = i, sum
		}
	}
	return ret
}

// sqDiff borrowed from image/color
func sqDiff(x, y uint8) uint32 {
	d := uint32(x - y)
	return (d * d) >> 2
}