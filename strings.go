package ptk

// UniqueSlice overwrites in
func UniqueSlice(in []string) (out []string) {
	set := make(Set, len(in))
	out = in[:0]

	for _, s := range in {
		if !set.Has(s) {
			out = append(out, s)
			set.Set(s)
		}
	}

	return
}
