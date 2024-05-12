package json

import (
	"encoding/json"
	"io"

	"github.com/JasonYao3/gnp/ch12_Data_Serialization/housework"
)

func Load(r io.Reader) ([]*housework.Chore, error) {
	var chores []*housework.Chore

	return chores, json.NewDecoder(r).Decode(&chores)
}

func Flush(w io.Writer, chores []*housework.Chore) error {
	return json.NewEncoder(w).Encode(chores)
}
