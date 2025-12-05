package player

import (
	"fmt"
	"testing"
	"time"
)

/*
func TestPlayer(t *testing.T) {

	player := Mp3PlayerNew()
	d, err := player.LoadFile("test.mp3")
	if err != nil {
		panic(err)
	}
	fmt.Println("Duration:", d)
	player.Play(0, d, PMPlayOnce)
	time.Sleep(time.Second)
	for i := 0; i < 10 && player.IsPlaying(); i++ {
		fmt.Println(player.GetProgress())
		time.Sleep(time.Second)
	}
	player.Stop()
	fmt.Println("Done")

}
*/
func TestPlayer2(t *testing.T) {

	player := Mp3PlayerNew()
	d, err := player.LoadFile("test.mp3")
	if err != nil {
		panic(err)
	}
	fmt.Println("Duration:", d)
	start := time.Now()
	player.Play(0, d, PMPlayOnce)
	for i := 1; i <= 35 && player.IsPlaying(); i++ {
		//		fmt.Printf("i: %v p:%2.1f\n", i, player.GetProgress())
		time.Sleep(time.Second)
		switch i {
		case 5:
			//			player.SetTimeRatio(1.5)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())
		case 10:
			//			player.SetTimeRatio(1.5)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())
		case 15:
			player.SetTimeRatio(1.5)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())
		case 20:
			player.SetTimeRatio(1.0)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())
		case 25:
			player.SetPitchScale(0.95)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())
		case 30:
			player.SetPitchScale(1.0)
			fmt.Printf("i=%v t:%v av:%v\n", i, time.Now().Sub(start), player.GetProgress())

		}
	}
	fmt.Printf("End t:%v av:%v\n", time.Now().Sub(start), player.GetProgress())

	player.Stop()
	fmt.Println("Done")

}
