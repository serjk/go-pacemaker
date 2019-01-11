package impl

import (
	"fmt"
	. "github.com/serjk/go-pacemaker"
	"log"
	"runtime"
	"strings"
	"unsafe"
)

/*
#cgo LDFLAGS: -Wl,-unresolved-symbols=ignore-all
#cgo pkg-config: libxml-2.0 glib-2.0 libqb pacemaker pacemaker-cib pacemaker-cluster libcfg

#include <stdio.h>
#include <crm/cib.h>
#include <crm/cluster.h>
#include <crm/services.h>
#include <crm/common/util.h>
#include <crm/common/xml.h>
#include <crm/common/mainloop.h>
#include <crm/common/logging.h>
#include <clients.h>
#include <corosync/cfg.h>

// Flags returned by go_cib_register_notify_callbacks
// indicating which notifications were actually
// available to register (different connection types
// enable different sets of notifications)
#define GO_CIB_NOTIFY_DESTROY 0x1
#define GO_CIB_NOTIFY_ADDREMOVE 0x2

extern int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type);
extern int go_cib_signoff(cib_t* cib);
extern int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options);
extern int go_cib_create(cib_t * cib, const char *section, xmlNode * data, int call_options);
extern int go_cib_update(cib_t * cib, const char *section, xmlNode * data, int call_options);
extern int go_cib_replace(cib_t * cib, const char *section, xmlNode * data, int call_options);
extern int go_cib_delete(cib_t * cib, const char *section, xmlNode * data, int call_options);

extern int go_nodes_get(pacemaker_client_t *client, xmlNode ** output_data);
extern pacemaker_client_t * new_pacemaker_client();
extern int destroy_pacemaker_client(pacemaker_client_t *client);
extern bool pacemaker_connect(pacemaker_client_t *client);

extern corosync_client_t * new_c_client();
extern int connect_cfg(corosync_client_t *client);
extern int get_node_addr(corosync_client_t *client, uint32_t nodeid, char ** addr);

extern unsigned int go_cib_register_notify_callbacks(cib_t * cib);
extern void go_add_idle_scheduler(GMainLoop* loop);
*/
import "C"

// When connecting to Pacemaker, we have
// to declare which type of connection to
// use. Since the API is read-only at the
// moment, it only really makes sense to
// pass Query to functions that take a
// CibConnection parameter.
type CibConnection int

func NewCibOpenConfig() CibOpenConfig {
	return CibOpenConfig{}
}

type CibOpenConfig struct {
	connection CibConnection
	file       string
	shadow     string
	server     string
	user       string
	passwd     string
	port       int
	encrypted  bool
}

const (
	Query              = C.cib_query
	Command            = C.cib_command
	NoConnection       = C.cib_no_connection
	CommandNonBlocking = C.cib_command_nonblocking

	ENXIO = C.ENXIO
	// Connection
	ENOTCONN     = C.ENOTCONN
	ECONNABORTED = C.ECONNABORTED
	ECONNREFUSED = C.ECONNREFUSED
	ECONNRESET   = C.ECONNRESET
	ENOTUNIQ     = C.ENOTUNIQ
	ECOMM        = C.ECOMM
	EOPNOTSUPP   = C.EOPNOTSUPP

	CS_ERR_LIBRARY   = C.CS_ERR_LIBRARY
	CS_ERR_NOT_EXIST = C.CS_ERR_NOT_EXIST
	CS_OK            = C.CS_OK
)

type cibOpType uint8

const (
	opCreate = cibOpType(iota)
	opUpdate
	opReplace
	opDelete
)

// Root entity representing the CIB. Can be
// populated with CIB data if the Decode
// method is used.
type CibClientImpl struct {
	cCib          *C.cib_t
	pClient       *C.pacemaker_client_t
	corosync      *C.corosync_client_t
	subscribers   map[int]CibEventFunc
	notifications uint
	conf          CibOpenConfig
}

type NewCibClient func(options ...func(*CibOpenConfig)) (CibClient, error)

func NewCibClientImpl(options ...func(*CibOpenConfig)) (CibClient, error) {
	var cib CibClientImpl
	conf := NewCibOpenConfig()
	for _, opt := range options {
		opt(&conf)
	}
	if conf.connection != Query && conf.connection != Command {
		conf.connection = Query
	}
	cib.conf = conf

	if conf.file != "" {
		s := C.CString(conf.file)
		defer C.free(unsafe.Pointer(s))
		cib.cCib = C.cib_file_new(s)
	} else if conf.shadow != "" {
		s := C.CString(conf.shadow)
		defer C.free(unsafe.Pointer(s))
		cib.cCib = C.cib_shadow_new(s)
	} else if conf.server != "" {
		s := C.CString(conf.server)
		defer C.free(unsafe.Pointer(s))

		u := C.CString(conf.user)
		defer C.free(unsafe.Pointer(u))

		p := C.CString(conf.passwd)
		defer C.free(unsafe.Pointer(p))

		var e = 0
		if conf.encrypted {
			e = 1
		}
		cib.cCib = C.cib_remote_new(s, u, p, (C.int)(conf.port), (C.gboolean)(e))
	} else {
		cib.cCib = C.cib_new()
	}

	cib.pClient = C.new_pacemaker_client()
	cib.corosync = C.new_c_client()

	return &cib, nil
}

func GetShadowFile(name string) string {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))

	shadow := C.get_shadow_file(s)
	defer C.free(unsafe.Pointer(shadow))

	return C.GoString(shadow)
}

func (cib *CibClientImpl) Connect() error {
	rc := C.go_cib_signon(cib.cCib, C.crm_system_name, (uint32)(cib.conf.connection))

	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}

	b := C.pacemaker_connect(cib.pClient)
	if !(bool)(b) {
		log.Printf("Unable to connect tp pacemaker")
	}

	rc = C.connect_cfg(cib.corosync)
	if rc != CS_OK {
		log.Printf("unable to init cfg, rc=%d", rc)
	}

	return nil
}

func (cib *CibClientImpl) Close() error {
	rc := C.go_cib_signoff(cib.cCib)
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	C.cib_delete(cib.cCib)
	C.destroy_pacemaker_client(cib.pClient)
	cib.cCib = nil
	cib.pClient = nil
	return nil
}

func (cib *CibClientImpl) Version() (*CibVersion, error) {
	var admin_epoch C.int
	var epoch C.int
	var num_updates C.int

	root, err := cib.queryImpl("/cib", true)
	if err != nil {
		return nil, err
	} else {
		defer C.free_xml(root)
	}
	if ok := C.cib_version_details(root,
		(*C.int)(unsafe.Pointer(&admin_epoch)),
		(*C.int)(unsafe.Pointer(&epoch)),
		(*C.int)(unsafe.Pointer(&num_updates))); ok == 1 {
		return &CibVersion{
			AdminEpoch: (int32)(admin_epoch),
			Epoch:      (int32)(epoch),
			NumUpdates: (int32)(num_updates)}, nil
	}
	return nil, NewCibError("failed to get CIB version details")
}

func (cib *CibClientImpl) Query() (*CibDocument, error) {
	root, err := cib.queryImpl("", false)
	if err != nil {
		return nil, err
	} else {
		defer C.free_xml(root)
		return dumpXmlToCibDoc(root)
	}
}

func (cib *CibClientImpl) QueryNoChildren() (*CibDocument, error) {
	root, err := cib.queryImpl("", true)
	if err != nil {
		return nil, err
	} else {
		defer C.free_xml(root)
		return dumpXmlToCibDoc(root)
	}
}

func (cib *CibClientImpl) QueryXPath(xpath string) (*CibDocument, error) {
	root, err := cib.queryImpl(xpath, false)
	if err != nil {
		return nil, err
	} else {
		defer C.free_xml(root)
		return dumpXmlToCibDoc(root)
	}
}

func (cib *CibClientImpl) QueryXPathNoChildren(xpath string) (*CibDocument, error) {
	root, err := cib.queryImpl(xpath, true)
	if err != nil {
		return nil, err
	} else {
		defer C.free_xml(root)
		return dumpXmlToCibDoc(root)
	}
}

func (cib *CibClientImpl) CreateObjInSection(section string, doc *CibDocument) (error) {
	return cib.updateSection(opCreate, section, doc)
}

func (cib *CibClientImpl) UpdateObjInSection(section string, doc *CibDocument) (error) {
	return cib.updateSection(opUpdate, section, doc)
}

func (cib *CibClientImpl) ReplaceObjInSection(section string, doc *CibDocument) (error) {
	return cib.updateSection(opReplace, section, doc)
}

func (cib *CibClientImpl) DeleteObjInSection(section string, doc *CibDocument) (error) {
	return cib.updateSection(opDelete, section, doc)
}

func (cib *CibClientImpl) GetLocalNodeName() (string, error) {
	return C.GoString(C.get_local_node_name()), nil
}

func (cib *CibClientImpl) GetNodesInfo() (*CibDocument, error) {
	var root *C.xmlNode

	rc := C.go_nodes_get(cib.pClient, (**C.xmlNode)(unsafe.Pointer(&root)))
	defer C.free_xml(root)

	if rc < 0 {
		msg := fmt.Sprintf("Got rc:= %d", rc)
		log.Printf(msg)
		return nil, convertPMCodeToError((int)(rc), msg)
	} else {
		return dumpXmlToCibDoc(root)
	}
}
func (cib *CibClientImpl) GetNodeIp(id uint) (string, error) {
	var ip *C.char

	rc := C.get_node_addr(cib.corosync, (C.uint32_t)(id), (**C.char)(unsafe.Pointer(&ip)))
	defer C.free(unsafe.Pointer(ip))
	if rc != CS_OK {
		return "", formatCSErrorRc((int)(rc))
	}
	ipStr := C.GoString(ip)
	return ipStr, nil
}

//trace=8,debug=7,info=6
func (cib *CibClientImpl) setLogLevel(level int) {
	C.set_crm_log_level(C.uint(level))
}

func (cib *CibClientImpl) updateSection(action cibOpType, section string, doc *CibDocument) (error) {
	var rc C.int
	var opts C.int

	docBytes := doc.Xml()
	docCStrPtr := (*C.char)(unsafe.Pointer(&docBytes[0]))

	root := C.string2xml(docCStrPtr)
	defer C.free_xml(root)

	opts = C.cib_sync_call
	if action == opCreate {
		opts |= C.cib_can_create
	}

	if section != "" {
		s := C.CString(section)
		defer C.free(unsafe.Pointer(s))
		rc = cib.cibFuncChoice(action, s, root, opts)
	} else {
		rc = cib.cibFuncChoice(action, nil, root, opts)
	}
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	return nil
}

func (cib *CibClientImpl) cibFuncChoice(action cibOpType,
	section *C.char,
	data *C.xmlNode,
	callOptions C.int) C.int {
	var rc C.int

	switch action {
	case opCreate:
		rc = C.go_cib_create(cib.cCib, section, data, callOptions)
	case opDelete:
		rc = C.go_cib_delete(cib.cCib, section, data, callOptions)
	case opReplace:
		rc = C.go_cib_replace(cib.cCib, section, data, callOptions)
	case opUpdate:
		rc = C.go_cib_update(cib.cCib, section, data, callOptions)
	default:
		rc = -EOPNOTSUPP
	}

	return rc
}

func (cib *CibClientImpl) queryImpl(xpath string, nochildren bool) (*C.xmlNode, error) {
	var root *C.xmlNode
	var rc C.int

	var opts C.int

	opts = C.cib_sync_call | C.cib_scope_local

	if xpath != "" {
		opts |= C.cib_xpath
	}

	if nochildren {
		opts |= C.cib_no_children
	}

	if xpath != "" {
		xp := C.CString(xpath)
		defer C.free(unsafe.Pointer(xp))
		rc = C.go_cib_query(cib.cCib, xp, (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	} else {
		rc = C.go_cib_query(cib.cCib, nil, (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	}
	if rc != C.pcmk_ok {
		defer C.free_xml(root)
		return nil, formatErrorRc((int)(rc))
	}
	return root, nil
}

func init() {
	C.crm_peer_init()
	s := C.CString("go-pacemaker")
	defer C.free(unsafe.Pointer(s))
	C.crm_log_init(s, C.LOG_INFO, 0, 0, 0, nil, 1)
}

func IsTrue(bstr string) bool {
	sl := strings.ToLower(bstr)
	return sl == "true" || sl == "on" || sl == "yes" || sl == "y" || sl == "1"
}

var the_cib *CibClientImpl

func (cib *CibClientImpl) Subscribers() map[int]CibEventFunc {
	return cib.subscribers
}

func (cib *CibClientImpl) Subscribe(callback CibEventFunc) (uint, error) {
	the_cib = cib
	if cib.subscribers == nil {
		cib.subscribers = make(map[int]CibEventFunc)
		flags := C.go_cib_register_notify_callbacks(cib.cCib)
		cib.notifications = uint(flags)
	}
	id := len(cib.subscribers)
	cib.subscribers[id] = callback
	return cib.notifications, nil
}

//export diffNotifyCallback
func diffNotifyCallback(current_cib *C.xmlNode) {
	for _, callback := range the_cib.subscribers {
		if doc, err := dumpXmlToCibDoc(current_cib); err != nil {
			return
		} else {
			callback(UpdateEvent, doc)
		}
	}
}

//export destroyNotifyCallback
func destroyNotifyCallback() {
	for _, callback := range the_cib.subscribers {
		callback(DestroyEvent, nil)
	}
}

//export goMainloopSched
func goMainloopSched() {
	runtime.Gosched()
}

func Mainloop() {
	mainloop := C.g_main_loop_new(nil, C.FALSE)
	C.go_add_idle_scheduler(mainloop)
	C.g_main_loop_run(mainloop)
	C.g_main_loop_unref(mainloop)
}

// Internal function used to create a CibError instance
// from a pacemaker return code.
func formatErrorRc(rc int) error {
	errorname := C.pcmk_errorname(C.int(rc))
	if errorname == nil {
		errorname = C.CString("")
		defer C.free(unsafe.Pointer(errorname))
	}

	strerror := C.pcmk_strerror(C.int(rc))
	if strerror == nil {
		strerror = C.CString("")
		defer C.free(unsafe.Pointer(strerror))
	}

	return convertPMCodeToError(rc, fmt.Sprintf("%d: %s %s", rc, C.GoString(errorname), C.GoString(strerror)))
}

// Internal function used to create a CibError instance
// from a pacemaker return code.
func formatCSErrorRc(rc int) error {
	errorname := C.cs_strerror(C.qb_to_cs_error(C.int(rc)))
	if errorname == nil {
		errorname = C.CString("")
		defer C.free(unsafe.Pointer(errorname))
	}
	return convertCSCodeToError(rc, fmt.Sprintf("%d: %s", rc, C.GoString(errorname)))
}

func dumpXmlToCibDoc(node *C.xmlNode) (*CibDocument, error) {
	buffer := C.dump_xml_unformatted(node)
	if buffer == nil {
		return nil, NewCibError("couldn't dump xml node")
	} else {
		defer C.free(unsafe.Pointer(buffer))
		bufLen := (C.int)(C.strlen(buffer))
		buffBytes := C.GoBytes(unsafe.Pointer(buffer), bufLen)
		return NewCibDocumentFromBytes(buffBytes)
	}
}
