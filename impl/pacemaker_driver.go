package impl

/*
#cgo LDFLAGS: -Wl,-unresolved-symbols=ignore-all


#include <crm/cluster.h>
#include <crm/common/util.h>
#include <crm/common/xml.h>
#include <crm/common/ipc.h>
#include <crm/common/mainloop.h>
#include <clients.h>


extern int go_nodes_get(pacemaker_client_t * client, xmlNode ** output_data);
extern pacemaker_client_t * new_pacemaker_client();
extern int destroy_pacemaker_client(pacemaker_client_t *client);
extern bool pacemaker_connect(pacemaker_client_t *client);

static void destroy(gpointer user_data){
    //crm_exit(CRM_EX_DISCONNECT);
    //return 102
}

static gint compare_node_uname(gconstpointer a, gconstpointer b)
{
    const crm_node_t *a_node = a;
    const crm_node_t *b_node = b;
    return strcmp(a_node->uname?a_node->uname:"", b_node->uname?b_node->uname:"");
}


static int dispatch(const char *buffer, ssize_t length, gpointer userdata) {
		xmlNode *msg = string2xml(buffer);

    if (msg) {
        get_nodes_context_t *ctx = userdata;
        crm_trace("dispatching %p", userdata);

        if (ctx == NULL) {
          crm_err("No ctx!");
          return 0;
        }
				crm_log_xml_trace(msg, "message");
        (ctx->data) = msg;
        (ctx->rc) = 0;
        g_main_loop_quit(ctx->loop);
        return 0;
    }

    return 0;
}

pacemaker_client_t * new_pacemaker_client() {
	pacemaker_client_t *client;
	client = calloc(1, sizeof(pacemaker_client_t));
  if (client == NULL) {
        return NULL;
   }

  GMainLoop *amainloop = g_main_loop_new(NULL, FALSE);
	get_nodes_context_t *context;
	context = calloc(1, sizeof(get_nodes_context_t));
  context->loop = amainloop;
  context->rc = 0;

 	struct ipc_client_callbacks node_callbacks = {
        .dispatch = dispatch,
        .destroy = destroy
  };
   mainloop_io_t *ipc =
                   mainloop_add_ipc_client(CRM_SYSTEM_MCP, G_PRIORITY_DEFAULT, 0, context, &node_callbacks);

   client->ctx = context;
   client->ipc = ipc;
   return client;
}

int destroy_pacemaker_client(pacemaker_client_t *client) {
	 if (client->ipc!=NULL) {
       crm_ipc_close(mainloop_get_ipc_client(client->ipc));
       crm_ipc_destroy(mainloop_get_ipc_client(client->ipc));
   }
   if (client->ctx->loop!=NULL) {
		 g_main_loop_unref(client->ctx->loop);
   }

   if (client->ctx!=NULL) {
   	 free(client->ctx);
   }

   free(client);
	 return 0;
}

int go_nodes_get(pacemaker_client_t *client, xmlNode ** output_data) {
	int rc;

	client->ctx->rc = 0;
	if (client->ipc!= NULL){
		xmlNode * msg = create_xml_node(NULL, "poke");
		rc = crm_ipc_send(mainloop_get_ipc_client(client->ipc), msg, 0, 0, NULL);
		free_xml(msg);
		if (rc < 0) {
			return rc;
		}
		g_main_loop_run(client->ctx->loop);

		if (client->ctx->rc == 0) {
			*output_data = client->ctx->data;
		}

		return client->ctx->rc;
	}
	return -1;
}

bool pacemaker_connect(pacemaker_client_t *client){
	if (client->ipc!= NULL){
		crm_info("reconnect to pacemaker");
		crm_ipc_close(mainloop_get_ipc_client(client->ipc));
		return crm_ipc_connect(mainloop_get_ipc_client(client->ipc));
	}
	return false;
}
*/
import "C"
