package main_test

import (
	"fmt"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	fmt.Printf("running basic test\n")
}

func TestFail(t *testing.T) {
	fmt.Printf("running fail test\n")
	t.Fail()
}

func TestSlow(t *testing.T) {
	time.Sleep(time.Millisecond * 1500)
	fmt.Printf("running fail test\n")
}

func TestHier(t *testing.T) {
	t.Run("section 1", func(t *testing.T) {
		fmt.Printf("running in section 1\n")
	})

	t.Run("section 2", func(t *testing.T) {
		fmt.Printf("running in section 2\n")
	})

	t.Run("section 3", func(t *testing.T) {
		fmt.Printf("running in section 3\n")
		t.Fail()
	})

	t.Run("section 4", func(t *testing.T) {
		fmt.Printf("running in section 4\n")
	})

}

func TestHierla(t *testing.T) {
	fmt.Printf("running basic test\n")
}

func TestMaybeSkip(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
}

func BenchmarkHello(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.Logf("hello\n")
	}
}
