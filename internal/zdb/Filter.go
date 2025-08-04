package zdb

import (
	"fmt"
	"time"
)

type Filter struct {
	Kind          string
	PartialName   string
	PlayLevelFrom int
	PlayLevelTo   int
	FunLevelFrom  int
	FunLevelTo    int
	RehearsalFrom time.Duration
	RehearsalTo   time.Duration
	//	LearnPriority int
	FirstNote     string
	IncludeHidden bool
	Mode          string
	Fifth         int
	List          int
	SortMethod    string
}

func (f *Filter) String() string {
	return fmt.Sprintf("K: %v fltr:%v P0:%v P1:%v F0:%v F1:%v R0:%v R1:%v H:%v Fth:%d Lst:%d",
		f.Kind, f.PartialName, f.PlayLevelFrom, f.PlayLevelTo, f.FunLevelFrom, f.FunLevelTo,
		f.RehearsalFrom, f.RehearsalTo, f.IncludeHidden, f.Fifth, f.List)
}
func FilterNew() Filter {
	return Filter{
		Kind:          "",
		PartialName:   "",
		PlayLevelFrom: 0,
		PlayLevelTo:   PlayLevelMax,
		FunLevelFrom:  0,
		FunLevelTo:    FunLevelMax,
		RehearsalFrom: 0,
		RehearsalTo:   0,
		//	LearnPriority: 0,
		Fifth:         1000,
		List:          0,
		FirstNote:     "",
		IncludeHidden: false,
		Mode:          "*",
		SortMethod:    "",
	}
}
