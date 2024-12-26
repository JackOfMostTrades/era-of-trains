package main

func DeleteFromSliceUnordered[T any](idx int, slice []T) []T {
	slice[idx] = slice[len(slice)-1]
	slice = slice[:len(slice)-1]
	return slice
}

func DeleteFromSliceOrdered[T any](idx int, slice []T) []T {
	for i := 0; i < len(slice)-idx-1; i++ {
		slice[idx+i] = slice[idx+i+1]
	}
	slice = slice[:len(slice)-1]
	return slice
}
