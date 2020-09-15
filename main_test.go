package main_test

import (
	"fmt"
	"testing"
	"time"
)

func TestSlowPrint(t *testing.T) {
	t.Logf("first line")
	time.Sleep(time.Second)
	t.Logf("first line")
	time.Sleep(time.Second)
	t.Logf("first line")
	time.Sleep(time.Second)
	t.Logf("first line")
	time.Sleep(time.Second)
}

func TestBasic(t *testing.T) {
	fmt.Printf("running basic test\n")
	t.Logf("running log comment in basic test\n")
}

func TestFail(t *testing.T) {
	fmt.Printf("running fail test\n")
	t.Fail()
}

func TestSlow(t *testing.T) {
	time.Sleep(time.Millisecond * 10)
	fmt.Printf("running fail test\n")
}

func TestNotSlow(t *testing.T) {
	time.Sleep(time.Millisecond * 4)
	fmt.Printf("running fail test\n")
}

func TestHanging(t *testing.T) {
	fmt.Printf("running hanging test\n")
	time.Sleep(time.Millisecond * 110)
	fmt.Printf("done running hanging test\n")
}

func TestHier(t *testing.T) {
	t.Run("section 1", func(t *testing.T) {
		fmt.Printf("running in section 1\n")
		time.Sleep(time.Millisecond * 4)
	})

	t.Run("section 2", func(t *testing.T) {
		fmt.Printf("running in section 2\n")
		time.Sleep(time.Millisecond * 4)
	})

	t.Run("section 3", func(t *testing.T) {
		fmt.Printf("running in section 3\n")
		t.Fail()
		time.Sleep(time.Millisecond * 4)
	})

	t.Run("section 4", func(t *testing.T) {
		fmt.Printf("running in section 4\n")
		time.Sleep(time.Millisecond * 4)
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

func TestTheSkip(t *testing.T) {
	t.Logf("Skip because of reason...")
	t.Skip()
}

func TestAlright(t *testing.T) {
}

func TestAlrightAgain(t *testing.T) {
}
