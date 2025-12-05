package zique

import (
	"fmt"
	//"os"
	"testing"
	"time"
)

var tunes = []string{
	"/home/yvon/Music/Mscz/Reel/xml/hills of kaitoke, the.musicxml",
	//"/home/yvon/Music/Mscz/Jig/xml/hills of glenorchy, the.musicxml",
	"/home/yvon/Music/Mscz/Reel/xml/Charlie_Mulvihill's.musicxml",
	"/home/yvon/Music/Mscz/Reel/xml/Heather_Breeze,_The.musicxml",
}
var coeff = map[int]int{
	1:  16,
	2:  8,
	4:  4,
	8:  2,
	16: 1,
}

func FeedBack(zique *ZiquePlayer) {
	p := time.Now()
	var evtP Tick
	for i := 0; ; i++ {
		evt := <-zique.TickBack
		now := time.Now()
		if i > 0 {
			expected := evtP.TickTime * time.Duration(evtP.Beats*coeff[evtP.BeatType]*evtP.XmlDivisions/4)
			delta := now.Sub(p)
			fmt.Println(expected, delta)
		} else {
			beatTime := evt.MeasureLengthTune / evt.Beats
			beatCount := float64(evt.MeasureLength) / float64(beatTime)
			fmt.Printf("%v/%v\n", beatTime, beatCount)
		}
		evtP = evt
		p = now
	}
}

func TestZique(t *testing.T) {
	player, msg := ZiquePlayerNew(".", "Synth")

	if msg != "" {
		t.Fatal(msg)
	}

	go FeedBack(player)

	/*	player.Play(tunes[0])

		for player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
		}
		player.Stop()*/

	player.PlaySet([]SetElem{
		{
			File:  tunes[0],
			Count: 2,
		},
		{
			File:  tunes[1],
			Count: 2,
		},
	})
	time.Sleep(60 * time.Second)
	player.Stop()

	/*
	   player.Play(tunes[1])
	   time.Sleep(5 * time.Second)
	   player.SetTempo(180)

	   time.Sleep(5 * time.Second)
	   player.SetTempo(60)
	   time.Sleep(5 * time.Second)
	   player.Stop()
	*/
}
