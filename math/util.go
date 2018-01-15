package math

// Lerp performs linear interpolation between the starting and ending values,
// based on the given amount.
func Lerp(start, end, amount float32) float32 {
	return start*(1.0-amount) + end*amount
}

// Clamp restricts a value between a minimum and maximum value.
func Clamp(value, min, max float32) float32 {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
