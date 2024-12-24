package main

func DeleteFromSliceUnordered[T any](idx int, slice []T) []T {
	slice[idx] = slice[len(slice)-1]
	slice = slice[:len(slice)-1]
	return slice
}
