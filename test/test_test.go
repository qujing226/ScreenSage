package test

import (
	"fmt"
	"testing"
)

func TestAlgo(t *testing.T) {
	fmt.Println(letterCombinations("23"))
}

func letterCombinations(digits string) (res []string) {
	length := len(digits)
	path := ""
	hash := get()
	var backtrack func(idx int)
	backtrack = func(idx int) {
		if len(path) == length {
			res = append(res, path)
			return
		}
		ch := string(digits[idx])
		//fmt.Printf("%T\n", ch)
		sets := hash[ch]
		for _, val := range sets {
			p := path
			path += val
			backtrack(idx + 1)
			path = p
		}
	}
	backtrack(0)
	return
}

func get() (hash map[string][]string) {
	hash = map[string][]string{
		"2": []string{"a", "b", "c"},
		"3": []string{"d", "e", "f"},
		"4": []string{"g", "h", "i"},
		"5": []string{"j", "k", "l"},
		"6": []string{"m", "n", "o"},
		"7": []string{"p", "q", "r", "s"},
		"8": []string{"t", "u", "v"},
		"9": []string{"w", "x", "y", "z"},
	}
	return hash
}
