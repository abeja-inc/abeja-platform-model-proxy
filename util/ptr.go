package util

func ToStringValue(target *string, defValue string) string {
	if target == nil {
		return defValue
	}
	return *target
}

func ToIntValue(target *int, defValue int) int {
	if target == nil {
		return defValue
	}
	return *target
}
