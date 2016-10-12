package dgraph

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	m := &Mutation{}
	m.Set("<_new_:qn> <name> \"Name\" .")
	m.Set("<_new_:qn> <text> \"Question text\" .")
	m.Del("<_uid_:uid> <question.correct> <_uid_:uid> .")
	fmt.Println(m.String())
}
