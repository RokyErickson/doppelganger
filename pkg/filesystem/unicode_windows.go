package filesystem

func IsUnicodeProbeFileName(_ string) bool {
	return false
}

func DecomposesUnicodeByPath(_ string) (bool, error) {
	return false, nil
}

func DecomposesUnicode(_ *Directory) (bool, error) {
	return false, nil
}
