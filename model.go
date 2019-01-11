package pacemaker

import "fmt"

type CibVersion struct {
	AdminEpoch int32
	Epoch      int32
	NumUpdates int32
}

func (ver *CibVersion) String() string {
	return fmt.Sprintf("%d:%d:%d", ver.AdminEpoch, ver.Epoch, ver.NumUpdates)
}

type Element struct {
	Type     string
	Id       string
	Attr     map[string]string
	Elements []*Element
}

type CibEvent int

const (
	UpdateEvent  CibEvent = 0
	DestroyEvent CibEvent = 1
)

//go:generate stringer -type=CibEvent
type CibEventFunc func(event CibEvent, doc *CibDocument)

type subscriptionData struct {
	Id       int
	Callback CibEventFunc
}

type NodeInfo struct {
	Id    string
	Uname string
	Ip    string
}
