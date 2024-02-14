package dns

func CloneSlice[E any, S ~[]E](s S) S {
	if s == nil {
		return nil
	}

	return append(S(nil), s...)
}
