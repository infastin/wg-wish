package queries

// KV meta information.
//
//	0.......4.........5........7
//	┌───────┬─────────┬────────┐
//	│version│collision│reserved│
//	└───────┴─────────┴────────┘
//
// version   (4 bits)  - binary version of the key value pair.
// collision (1 bit)   - indicates whether key has collisions or not.
// reserved  (3 bits)  - reserved for future.
// NOTE: Since versions starts from 1, the maximum version is 16.
type Meta byte

func (m Meta) Append(b []byte) []byte {
	return append(b, byte(m))
}

func (m Meta) Version() int {
	return int(m&0xF) + 1
}

func (m Meta) SetVersion(v int) Meta {
	return (m & 0xF0) | Meta(v-1)
}

func (m Meta) Collision() bool {
	return (m & 0x10) != 0
}

func (m Meta) SetCollision(c bool) Meta {
	if c {
		return m | 0x10
	}
	return m & 0xEF
}
