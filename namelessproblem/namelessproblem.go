package main

func AddElement(numbers *[]int, element int) {
	*numbers = append(*numbers, element)
}

func FindMin(numbers *[]int) int {
	if *numbers == nil || len(*numbers) == 0 {
		return 0
	} else {
		m := (*numbers)[0]
		for _, n := range *numbers {
			if n < m {
				m = n
			}
		}
		return m
	}
}

func ReverseSlice(numbers *[]int) {
	for i := 0; i < len(*numbers)/2; i++ {
		j := len(*numbers) - i - 1
		(*numbers)[i], (*numbers)[j] = (*numbers)[j], (*numbers)[i]
	}
}

func SwapElements(numbers *[]int, i, j int) {
	if i >= 0 && i < len(*numbers) && j >= 0 && j < len(*numbers) {
		(*numbers)[i], (*numbers)[j] = (*numbers)[j], (*numbers)[i]
	}
}
