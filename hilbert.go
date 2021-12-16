package pix

// Translated from this public domain C++ library:
// https://github.com/rawrunprotected/hilbert_curves

// These are multiplication tables of the alternating group A4,
// preconvolved with the mapping between Morton and Hilbert curves.
var mortonToHilbertTable = []uint8{
	48, 33, 27, 34, 47, 78, 28, 77,
	66, 29, 51, 52, 65, 30, 72, 63,
	76, 95, 75, 24, 53, 54, 82, 81,
	18, 3, 17, 80, 61, 4, 62, 15,
	0, 59, 71, 60, 49, 50, 86, 85,
	84, 83, 5, 90, 79, 56, 6, 89,
	32, 23, 1, 94, 11, 12, 2, 93,
	42, 41, 13, 14, 35, 88, 36, 31,
	92, 37, 87, 38, 91, 74, 8, 73,
	46, 45, 9, 10, 7, 20, 64, 19,
	70, 25, 39, 16, 69, 26, 44, 43,
	22, 55, 21, 68, 57, 40, 58, 67,
}

var hilbertToMortonTable = []uint8{
	48, 33, 35, 26, 30, 79, 77, 44,
	78, 68, 64, 50, 51, 25, 29, 63,
	27, 87, 86, 74, 72, 52, 53, 89,
	83, 18, 16, 1, 5, 60, 62, 15,
	0, 52, 53, 57, 59, 87, 86, 66,
	61, 95, 91, 81, 80, 2, 6, 76,
	32, 2, 6, 12, 13, 95, 91, 17,
	93, 41, 40, 36, 38, 10, 11, 31,
	14, 79, 77, 92, 88, 33, 35, 82,
	70, 10, 11, 23, 21, 41, 40, 4,
	19, 25, 29, 47, 46, 68, 64, 34,
	45, 60, 62, 71, 67, 18, 16, 49,
}

func transformCurve(in, bits uint32, lookupTable []uint8) uint32 {
	var transform, out uint32
	for i := int32(3 * (bits - 1)); i >= 0; i -= 3 {
		transform = uint32(lookupTable[transform|((in>>i)&7)])
		out = (out << 3) | (transform & 7)
		transform &= ^uint32(7)
	}
	return out
}

// The 'bits' parameter in the Morton/Hilbert conversion functions refers to the extent of the encoded space.
// For instance a bits value of 1 indicates a 2x2(x2) space, 2 is a 4x4(x4) space, 3 is a 8x8(x8) space and so on.
func mortonToHilbert3D(mortonIndex uint32, bits uint32) uint32 {
	return transformCurve(uint32(mortonIndex), bits, mortonToHilbertTable)
}

func hilbertToMorton3D(hilbertIndex, bits uint32) uint32 {
	return transformCurve(hilbertIndex, bits, hilbertToMortonTable)
}

func hilbertCode(x, y, z uint8) uint32 {
	return mortonToHilbert3D(uint32(mortonCode(x, y, z)), 8)
}

func prefixScan(x uint32) uint32 {
	x = (x >> 8) ^ x
	x = (x >> 4) ^ x
	x = (x >> 2) ^ x
	x = (x >> 1) ^ x
	return x
}

func hilbertToXY(i, bits uint32) (uint32, uint32) {
	i = i << (32 - 2*bits)

	i0 := smoosh1(i)
	i1 := smoosh1(i >> 1)

	t0 := (i0 | i1) ^ 0xFFFF
	t1 := i0 & i1

	prefixT0 := prefixScan(t0)
	prefixT1 := prefixScan(t1)

	a := (((i0 ^ 0xFFFF) & prefixT1) | (i0 & prefixT0))

	return (a ^ i1) >> (16 - bits), (a ^ i0 ^ i1) >> (16 - bits)
}

func xyToHilbert(x, y, bits uint32) uint32 {
	x = x << (16 - bits)
	y = y << (16 - bits)

	var A, B, C, D uint32

	// Initial prefix scan round, prime with x and y
	{
		a := x ^ y
		b := 0xFFFF ^ a
		c := 0xFFFF ^ (x | y)
		d := x & (y ^ 0xFFFF)

		A = a | (b >> 1)
		B = (a >> 1) ^ a

		C = ((c >> 1) ^ (b & (d >> 1))) ^ c
		D = ((a & (c >> 1)) ^ (d >> 1)) ^ d
	}

	{
		a := A
		b := B
		c := C
		d := D

		A = ((a & (a >> 2)) ^ (b & (b >> 2)))
		B = ((a & (b >> 2)) ^ (b & ((a ^ b) >> 2)))

		C ^= ((a & (c >> 2)) ^ (b & (d >> 2)))
		D ^= ((b & (c >> 2)) ^ ((a ^ b) & (d >> 2)))
	}

	{
		a := A
		b := B
		c := C
		d := D

		A = ((a & (a >> 4)) ^ (b & (b >> 4)))
		B = ((a & (b >> 4)) ^ (b & ((a ^ b) >> 4)))

		C ^= ((a & (c >> 4)) ^ (b & (d >> 4)))
		D ^= ((b & (c >> 4)) ^ ((a ^ b) & (d >> 4)))
	}

	// Final round and projection
	{
		a := A
		b := B
		c := C
		d := D

		C ^= ((a & (c >> 8)) ^ (b & (d >> 8)))
		D ^= ((b & (c >> 8)) ^ ((a ^ b) & (d >> 8)))
	}

	// Undo transformation prefix scan
	a := C ^ (C >> 1)
	b := D ^ (D >> 1)

	// Recover index bits
	i0 := x ^ y
	i1 := b | (0xFFFF ^ (i0 | a))

	return ((spread1(i1) << 1) | spread1(i0)) >> (32 - 2*bits)
}

// the following statements return 0–7 in order:
// hilbertCode(0, 0, 0) // => 0
// hilbertCode(0, 1, 0) // => 1
// hilbertCode(0, 1, 1) // => 2
// hilbertCode(0, 0, 1) // => 3
// hilbertCode(1, 0, 1) // => 4
// hilbertCode(1, 1, 1) // => 5
// hilbertCode(1, 1, 0) // => 6
// hilbertCode(1, 0, 0) // => 7
