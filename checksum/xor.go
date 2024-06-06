package checksum

func XOR(in []byte) (out byte) {
	for i := range in {
		out ^= in[i]
	}

	return out
}
