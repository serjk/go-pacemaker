
#include <crm/common/util.h>
#include <crm/common/xml.h>
#include <crm/common/ipc.h>
#include <crm/common/mainloop.h>
#include <corosync/cfg.h>

typedef struct get_nodes_context_s {
	xmlNode *data;
	GMainLoop *loop;
	int rc;
} get_nodes_context_t;

typedef struct pacemaker_client_s {
	mainloop_io_t *ipc;
	get_nodes_context_t *ctx;
} pacemaker_client_t;

typedef struct corosync_client_s {
    corosync_cfg_handle_t *cfg_handle;
}corosync_client_t;