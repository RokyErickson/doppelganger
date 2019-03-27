package filesystem

func PreservesExecutabilityByPath(_ string) (bool, error) {
	return false, nil
}

func PreservesExecutability(_ *Directory) (bool, error) {
	return false, nil
}
