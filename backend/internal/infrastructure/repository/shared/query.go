package infrastructurerepositoryshared

func NormalizeLimit(limit int) uint64 {
	if limit < 0 {
		return 0
	}

	return uint64(limit)
}

func NormalizeOffset(page int, limit int) uint64 {
	if page <= 1 || limit <= 0 {
		return 0
	}

	return uint64((page - 1) * limit)
}

func SearchPattern(search *string) (string, bool) {
	if search == nil || *search == "" {
		return "", false
	}

	return "%" + *search + "%", true
}
