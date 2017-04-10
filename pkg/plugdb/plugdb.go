package plugdb

/*
#include <stdlib.h>
*/
import "C"

// Random randoms
func Random() int {
	return int(C.random())
}

// Seed seeds
func Seed(i int) {
	C.srandom(C.uint(i))
}
