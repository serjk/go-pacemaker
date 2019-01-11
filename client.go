package pacemaker

//go:generate mockgen -package pacemaker -destination client_mock.go -source $GOFILE

type CibClient interface {
	CreateObjInSection(section string, doc *CibDocument) error
	UpdateObjInSection(section string, doc *CibDocument) error
	ReplaceObjInSection(section string, doc *CibDocument) error
	DeleteObjInSection(section string, doc *CibDocument) error

	Query() (*CibDocument, error)
	QueryXPathNoChildren(xpath string) (*CibDocument, error)
	QueryXPath(xpath string) (*CibDocument, error)
	Version() (*CibVersion, error)

	GetLocalNodeName() (string, error)
	GetNodesInfo() (*CibDocument, error)
	GetNodeIp(uint) (string, error)

	Close() error
	Connect() error

	Subscribe(callback CibEventFunc) (uint, error)
	Subscribers() map[int]CibEventFunc
}
