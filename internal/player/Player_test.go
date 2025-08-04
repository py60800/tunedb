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
	player.Play(0, d, PMPlayOnce)
	for i := 0; i < 20 && player.IsPlaying(); i++ {
		fmt.Printf("i: %v p:%2.1f\n", i, player.GetProgress())
		time.Sleep(time.Second)
		switch i {
		case 5:
			player.SetTimeRatio(1.5)
		case 10:
			player.SetTimeRatio(1.0)
		case 15:
			player.SetPitchScale(0.95)
		case 20:
			player.SetPitchScale(1.0)

		}
	}
	player.Stop()
	fmt.Println("Done")

}
