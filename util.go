package main

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func limit(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func overlap(a1, a2, b1, b2 int) bool {
	return min(a1, a2) <= max(b1, b2) && min(b1, b2) <= max(a1, a2)
}
