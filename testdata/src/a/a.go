package a

func ex1() int { // want "ex1(): returned value is always ignored."
	return 0
}

func ex2() (int, int) { // want "ex2(): 0th returned value is always ignored."
	return 0, 0
}

func ex3() (int, int, int) { // OK
	return 0, 0, 0
}

func call() int {
	_ = ex1()

	_, a := ex2()
	_, _ = ex2()

	a, b, _ := ex3()
	_, _, a = ex3()

	_, b, c, d := ex4()
	a, b, _, _, _ = ex5()

	t := t{}
	a, _, _ = t.ex7()
	a, b, c, d = t.ex8()

	var f = func() (int, int) { return 0, 1 } // want "f: 1st returned value is always ignored."
	e, _ := f()

	// TODO: this should also be warned
	// s := s{}
	// s.f = func() (int, int, int) { return 0, 1, 2 }
	// _, b, c = s.f()

	return a + b + c + d + e
}

type s struct {
	f func() (int, int, int)
}

var _, b, c, d = ex4()
var _, _, e, f, g = ex5()

func ex4() (int, int, int, int) { // want "ex4(): 0th returned value is always ignored."
	return 0, 0, 0, 0
}

func ex5() (int, int, int, int, int) { // OK
	return 0, 0, 0, 0, 0
}

func ex6() int { // OK
	return 6
}

type t struct{}

func (t) ex7() (int, int, int) { // want "ex7(): 1st returned value is always ignored."
	return 0, 0, 0
}

func (t) ex8() (int, int, int, int) { // OK
	return 0, 0, 0, 0
}
