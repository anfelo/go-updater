package helpers

type SliceType interface {
	int | string
}

func Contains[T SliceType](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
