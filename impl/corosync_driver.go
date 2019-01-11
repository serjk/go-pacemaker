package impl

/*
#cgo LDFLAGS: -Wl,-unresolved-symbols=ignore-all

#include <corosync/cfg.h>
#include <crm/common/util.h>

#include <clients.h>

#define INTERFACE_MAX		8
#define CS_MAX_NAME_LENGTH 256

extern corosync_client_t * new_c_client();
extern int connect_cfg(corosync_client_t * client);
extern int get_node_addr(corosync_client_t *client, uint32_t nodeid, char ** addr);


corosync_client_t * new_c_client() {
	corosync_cfg_handle_t * t = calloc(1, sizeof(corosync_cfg_handle_t));
	corosync_client_t * client = calloc(1, sizeof(corosync_client_t));
	client->cfg_handle = t;
	return client;
}

int connect_cfg(corosync_client_t *client) {
	int err;
	static corosync_cfg_callbacks_t c_callbacks = {
	.corosync_cfg_shutdown_callback = NULL
	};

	err = corosync_cfg_initialize(client->cfg_handle, &c_callbacks);
	if (err!= CS_OK){
		crm_err("got error while initializing: %d", err);
		return err;
	}
	return CS_OK;
}

int get_node_addr(corosync_client_t *client, uint32_t nodeid, char ** addr){
	int err;
	int numaddrs;
	corosync_cfg_node_address_t addrs[INTERFACE_MAX];
	int i;
	socklen_t addrlen;

	err = corosync_cfg_get_node_addrs(*client->cfg_handle, nodeid, INTERFACE_MAX, &numaddrs, addrs);
	if (err != CS_OK) {
		crm_err("got error: %d", err);
		return err;
	}

	for (i=0; i<numaddrs; i++) {
		*addr = calloc(INET6_ADDRSTRLEN, sizeof(char));
		struct sockaddr_storage *ss = (struct sockaddr_storage *)addrs[i].address;
		struct sockaddr_in *sin = (struct sockaddr_in *)addrs[i].address;
		struct sockaddr_in6 *sin6 = (struct sockaddr_in6 *)addrs[i].address;
		void *saddr;

		if (!ss->ss_family) {
			continue;
		}

		if (ss->ss_family == AF_INET6) {
			saddr = &sin6->sin6_addr;
		} else {
			saddr = &sin->sin_addr;
		}

		inet_ntop(ss->ss_family, saddr, *addr, INET6_ADDRSTRLEN*sizeof(char));
		crm_trace("Go ip add %s  for node: %s", addr, nodeid);
		return CS_OK;
	}

	return CS_ERR_NOT_EXIST;
}

*/
import "C"
