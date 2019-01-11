package impl

import (
	"errors"
	"fmt"
	"github.com/clbanning/mxj"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"bytes"
	. "github.com/serjk/go-pacemaker"
	"github.com/stretchr/testify/assert"
)

var (
	someRes = []byte( `<primitive id="res" class="systemd" type="some_type">
                    <operations>
                        <op name="stop" interval="0" timeout="100" id="res-stop-20"/>
                        <op name="start" interval="0" timeout="100" id="res-start-20"/>
                        <op name="monitor" interval="20" timeout="20" id="res-monitor-20"/>
                    </operations>
                </primitive>`)
	fullXmlNode = []byte("<configuration><nodes><node id=\"zzz\" uname=\"unique\" type=\"normal\"/></nodes></configuration>")
	xmlResource = []byte("<primitive id=\"Id\" class=\"ocf\" provider=\"heartbeat\" type=\"IPaddr\"><operations>" +
		"<op id=\"Id-monitor\" name=\"monitor\" interval=\"300s\"/></operations>" +
		"<instance_attributes id=\"Id-params\"><nvpair id=\"Id-ip\" name=\"ip\" value=\"localhost\"/>" +
		"</instance_attributes></primitive>")
	newXmlNode           = []byte("<node id=\"zzz\" uname=\"unique\" type=\"normal\"/>")
	existedXmlNode       = []byte("<node id=\"xxx\" uname=\"c001n01\" type=\"normal\"/>")
	updateExistedXmlNode = []byte("<node id=\"xxx\" uname=\"nounique\" type=\"normal\"/>")
	xmlResourceMetaAttr  = []byte("<nvpair id=\"Id-ip\" name=\"ip\" value=\"127.0.0.1\"/>")

	nvPair = []byte("<nvpair id=\"myAddr-ip\" name=\"ip\" value=\"127.0.0.1\"/>")
	str    = []byte("<expression attribute=\"node-type\" id=\"health-location-rule-expression\" operation=\"ne\" value=\"storage-processor-1\"/>")
)

const (
	testTempDirPrefix = "test_dir"
	fileName          = "versioned-resources.xml"
)

func TestXmlPath(t *testing.T) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if !assert.NoError(t, err) {
		return
	}
	err = cib.Connect()
	if !assert.NoError(t, err) {
		return
	}
	defer cib.Close()

	doc, err := cib.Query()
	if !assert.NoError(t, err) {
		return
	}
	path := xmlpath.MustCompile("/cib/configuration/nodes/node[@id='xxx']/@type")
	root, err := xmlpath.Parse(bytes.NewReader(doc.Xml()))
	if !assert.NoError(t, err) {
		return
	}
	value, ok := path.String(root)

	assert.True(t, ok)
	assert.Equal(t, "normal", value)
}

func TestVersion(t *testing.T) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		t.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	ver, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	if ver.AdminEpoch != 1 {
		t.Error("Expected admin_epoch == 1, got ", ver.AdminEpoch)
	}
	if ver.Epoch != 0 {
		t.Error("Expected epoch == 0, got ", ver.Epoch)
	}
}

func TestCreateObj(t *testing.T) {

	cib := initTempCibFile(t)
	defer cib.Close()

	v0, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	doc, err := mxj.NewMapXml([]byte(fullXmlNode))

	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	err = cib.CreateObjInSection("", cibDocument)
	if err != nil {
		t.Error(err)
	}

	v1, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v1.Epoch-v0.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		assert.Fail(t, msg)
	}
}

func TestCreateObjInPredefinedSection(t *testing.T) {

	cib := initTempCibFile(t)
	defer cib.Close()

	v0, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	doc, err := mxj.NewMapXmlReader(bytes.NewReader(xmlResource))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.CreateObjInSection("resources", cibDocument)
	if err != nil {
		t.Error(err)
	}

	v1, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v1.Epoch-v0.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		t.Error(errors.New(msg))
	}
}

func TestReplaceObjInPredefinedSection(t *testing.T) {

	cib := initTempCibFile(t)
	defer cib.Close()

	v0, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	doc, err := mxj.NewMapXmlReader(bytes.NewReader(nvPair))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.ReplaceObjInSection("resources", cibDocument)
	if err != nil {
		t.Error(err)
	}

	v1, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v1.Epoch-v0.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		t.Error(errors.New(msg))
	}
}

func TestDeleteNode(t *testing.T) {
	cib := initTempCibFile(t)
	defer cib.Close()

	doc, err := mxj.NewMapXmlReader(bytes.NewReader(existedXmlNode))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.DeleteObjInSection("nodes", cibDocument)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cib.QueryXPathNoChildren("/cib/configuration/nodes/node[@id='xxx']")

	assert.Error(t, err)
	assert.Equal(t, reflect.TypeOf(&NotFoundObject{}), reflect.TypeOf(err))
}

func TestCreateUpdateDeleteNode(t *testing.T) {
	cib := initTempCibFile(t)
	defer cib.Close()

	doc, err := mxj.NewMapXmlReader(bytes.NewReader(existedXmlNode))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc1, err := mxj.NewMapXmlReader(bytes.NewReader(updateExistedXmlNode))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument1, err := NewCibDocument(doc1)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.DeleteObjInSection("nodes", cibDocument)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err = cib.CreateObjInSection("nodes", cibDocument)
		if err != nil {
			t.Fatal(err)
		}
		err = cib.UpdateObjInSection("nodes", cibDocument1)
		if err != nil {
			t.Fatal(err)
		}

		err = cib.DeleteObjInSection("nodes", cibDocument1)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = cib.QueryXPathNoChildren("/cib/configuration/nodes/node[@id='xxx']")

	assert.Error(t, err)
	assert.Equal(t, reflect.TypeOf(&NotFoundObject{}), reflect.TypeOf(err))
}

func TestUpdateMetaAttribute(t *testing.T) {

	cib := initTempCibFile(t)
	defer cib.Close()

	v0, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	doc, err := mxj.NewMapXmlReader(bytes.NewReader(xmlResource))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.CreateObjInSection("resources", cibDocument)
	if err != nil {
		t.Error(err)
	}

	v1, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v1.Epoch-v0.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		t.Error(errors.New(msg))
	}

	doc1, err := mxj.NewMapXmlReader(bytes.NewReader(xmlResourceMetaAttr))
	if err != nil {
		t.Fatal(err)
	}

	cibDocument1, err := NewCibDocument(doc1)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.UpdateObjInSection("resources", cibDocument1)
	if err != nil {
		t.Error(err)
	}

	v2, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v2.Epoch-v1.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		t.Error(errors.New(msg))
	}
}

func _TestCreateNewNodePredefinedSection(t *testing.T) {

	cib := initTempCibFile(t)
	defer cib.Close()

	v0, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	doc, err := mxj.NewMapXml([]byte(newXmlNode))

	if err != nil {
		t.Fatal(err)
	}

	cibDocument, err := NewCibDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	err = cib.CreateObjInSection("nodes", cibDocument)
	if err != nil {
		t.Error(err)
	}

	v1, err := cib.Version()
	if err != nil {
		t.Error(err)
	}
	if v1.Epoch-v0.Epoch != 1 {
		msg := fmt.Sprintf("cib v0: %v & cib v1: %v", v0, v1)
		t.Error(errors.New(msg))
	}
}

func TestObjectNotFound(t *testing.T) {

	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	_, err = cib.QueryXPath("//resources/primitive[@id='myId']")

	assert.Error(t, err)
	assert.Equal(t, reflect.TypeOf(&NotFoundObject{}), reflect.TypeOf(err))
}

func TestGetLocalNode(t *testing.T) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	name, err := cib.GetLocalNodeName()
	if err != nil {
		log.Fatal(err)
	}

	name1, err := cib.GetLocalNodeName()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, name, name1)
}

func TestGetNodeInfo(t *testing.T) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	//name, err := cib.GetLocalNodeName()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%s\n", name)

	info, err := cib.GetNodesInfo()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		//return
	}
	fmt.Printf("%v\n", info)

	info1, err := cib.GetNodesInfo()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%v\n", info1)

	err = cib.Close()
	if err != nil {
		t.Error("Error: %v\n", err)
		return
	}
}

func _TestGetNodeIp(t *testing.T) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	info, err := cib.GetNodeIp(0)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		//return
	}
	fmt.Printf("%v\n", info)

	info1, err := cib.GetNodeIp(0)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		//return
	}
	fmt.Printf("%v\n", info1)
}

func BenchmarkGetManually(b *testing.B) {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		b.Fatal(err)
	}
	defer cib.Close()

	for j := 0; j < b.N; j++ {
		_, err := cib.QueryXPathNoChildren("//nodes/node[@id=\"xxx\"]")
		if err != nil {
			b.Fail()
		}
	}
}

func ExampleQuery() {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.Query()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", doc.Xml()[0:4])
	// Output: <cib
}

func ExampleQueryXPath() {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.QueryXPath("//nodes/node[@id=\"xxx\"]")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", doc.Xml())
	// Output: <node id="xxx" type="normal" uname="c001n01"/>
}

func ExampleSimple() {
	cib, err := NewCibClientImpl(FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	err = cib.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.Query()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", string(doc.Xml())[0:4])
	// Output: <cib
}

func initTempCibFile(t *testing.T) CibClient {
	tempDir, err := ioutil.TempDir("", testTempDirPrefix)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadFile("testdata/simple.xml")
	if err != nil {
		t.Error(err)
	}
	filePath := tempDir + "/" + fileName
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		t.Error(err)
	}

	cib, err := NewCibClientImpl(FromFile(filePath), ForCommand)
	if err != nil {
		log.Fatal(err)
	}

	err = cib.Connect()
	if err != nil {
		t.Fatal(err)
	}

	return cib
}
