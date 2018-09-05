package asset

import (
	"fmt"
	"testing"
)

func Test_insertSort(t *testing.T) {
	var s []Asset
	as1 := Asset{
		Value: 30,
	}
	as2 := Asset{
		Value: 20,
	}

	af := ascInsert(&s, as1)
	fmt.Println(af)
	af = ascInsert(&af, as2)
	af = ascInsert(&af, as2)
	af = ascInsert(&af, as1)
	// af = insertSort(af, as2)
	fmt.Println(af)
	t.Fail()
	// type args struct {
	// 	assets *[]Asset
	// 	asset  Asset
	// }
	// tests := []struct {
	// 	name string
	// 	args args
	// 	want *[]Asset
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if got := insertSort(tt.args.assets, tt.args.asset); !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("insertSort() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }
}
