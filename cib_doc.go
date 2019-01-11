package pacemaker

import (
	"github.com/clbanning/mxj"
	"log"
)

func NewCibDocumentFromBytes(body []byte) (*CibDocument, error) {
	mv, err := mxj.NewMapXml(body)
	if err != nil {
		log.Printf("Failed parse xml %s", err)
		return nil, err
	}

	return &CibDocument{mv}, nil
}

func NewCibDocument(body mxj.Map) (*CibDocument, error) {
	return &CibDocument{body}, nil
}

type CibDocument struct {
	MV mxj.Map
}

func (doc *CibDocument) Json() []byte {
	v, _ := doc.MV.Json()
	return v
}

func (doc *CibDocument) Xml() []byte {
	v, _ := doc.MV.XmlIndent("", "  ")
	return v
}
