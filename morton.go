package pix

// Implementation of functions in this file based on
// https://fgiesen.wordpress.com/2009/12/13/decoding-morton-codes/

type MortonCode uint32

func smoosh2(x MortonCode) uint32 {
  x &= 0x09249249                  // x = ---- 9--8 --7- -6-- 5--4 --3- -2-- 1--0
  x = (x ^ (x >> 2)) & 0x030c30c3  // x = ---- --98 ---- 76-- --54 ---- 32-- --10
  x = (x ^ (x >> 4)) & 0x0300f00f  // x = ---- --98 ---- ---- 7654 ---- ---- 3210
  x = (x ^ (x >> 8)) & 0xff0000ff  // x = ---- --98 ---- ---- ---- ---- 7654 3210
  x = (x ^ (x >> 16)) & 0x000003ff // x = ---- ---- ---- ---- ---- --98 7654 3210
  return uint32(x)
}

// "Insert" two 0 bits after each of the 10 low bits of x
func spread2(x uint32) MortonCode {
  x &= 0x000003ff                  // x = ---- ---- ---- ---- ---- --98 7654 3210
  x = (x ^ (x << 16)) & 0xff0000ff // x = ---- --98 ---- ---- ---- ---- 7654 3210
  x = (x ^ (x << 8)) & 0x0300f00f  // x = ---- --98 ---- ---- 7654 ---- ---- 3210
  x = (x ^ (x << 4)) & 0x030c30c3  // x = ---- --98 ---- 76-- --54 ---- 32-- --10
  x = (x ^ (x << 2)) & 0x09249249  // x = ---- 9--8 --7- -6-- 5--4 --3- -2-- 1--0
  return MortonCode(x)
}

func mortonCode(x, y, z uint8) MortonCode {
  return (spread2(uint32(z)) << 2) + (spread2(uint32(y)) << 1) + spread2(uint32(x))
}

func mortonX(code MortonCode) uint8 { return uint8(smoosh2(code >> 0)) }
func mortonY(code MortonCode) uint8 { return uint8(smoosh2(code >> 1)) }
func mortonZ(code MortonCode) uint8 { return uint8(smoosh2(code >> 2)) }

// Each mask contains 0 in the slot for its component, and 1s elsewhere,
// assuming 10 bits per number and filling in the high 2 bits with 1 to be safe.
const xMask = MortonCode(0b11110110110110110110110110110110) // mortonCode(0, 1023, 1023)
const yMask = MortonCode(0b11101101101101101101101101101101) // mortonCode(1023, 0, 1023)
const zMask = MortonCode(0b11011011011011011011011011011011) // mortonCode(1023, 1023, 0)

// Less-than and greater than in morton space; the masks are used
// to equalize the values of all other coordinates to 1
func ltMortonX(a, b MortonCode) bool { return a|xMask < b|xMask } // a.x < b.x
func ltMortonY(a, b MortonCode) bool { return a|yMask < b|yMask } // a.y < b.y
func ltMortonZ(a, b MortonCode) bool { return a|zMask < b|zMask } // a.z < b.z
func gtMortonX(a, b MortonCode) bool { return a|xMask > b|xMask } // a.x > b.x
func gtMortonY(a, b MortonCode) bool { return a|yMask > b|yMask } // a.y > b.y
func gtMortonZ(a, b MortonCode) bool { return a|zMask > b|zMask } // a.z > b.z

// ---

// 2d

// "Insert" a 0 bit after each of the 16 low bits of x
func spread1(x uint32) uint32 {
  x &= 0x0000ffff                 // x = ---- ---- ---- ---- fedc ba98 7654 3210
  x = (x ^ (x << 8)) & 0x00ff00ff // x = ---- ---- fedc ba98 ---- ---- 7654 3210
  x = (x ^ (x << 4)) & 0x0f0f0f0f // x = ---- fedc ---- ba98 ---- 7654 ---- 3210
  x = (x ^ (x << 2)) & 0x33333333 // x = --fe --dc --ba --98 --76 --54 --32 --10
  x = (x ^ (x << 1)) & 0x55555555 // x = -f-e -d-c -b-a -9-8 -7-6 -5-4 -3-2 -1-0
  return x
}

func smoosh1(x uint32) uint32 {
  x &= 0x55555555                 // x = -f-e -d-c -b-a -9-8 -7-6 -5-4 -3-2 -1-0
  x = (x ^ (x >> 1)) & 0x33333333 // x = --fe --dc --ba --98 --76 --54 --32 --10
  x = (x ^ (x >> 2)) & 0x0f0f0f0f // x = ---- fedc ---- ba98 ---- 7654 ---- 3210
  x = (x ^ (x >> 4)) & 0x00ff00ff // x = ---- ---- fedc ba98 ---- ---- 7654 3210
  x = (x ^ (x >> 8)) & 0x0000ffff // x = ---- ---- ---- ---- fedc ba98 7654 3210
  return x
}

func interleave2(x, y uint32) uint32    { return (spread1(y) << 1) + spread1(x) }
func deinterleave2x(code uint32) uint16 { return uint16(smoosh1(code >> 0)) }
func deinterleave2y(code uint32) uint16 { return uint16(smoosh1(code >> 1)) }
