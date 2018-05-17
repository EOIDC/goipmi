#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/select.h>
#include <sys/ioctl.h>
#include <linux/ipmi.h>

#include "ipmi.h"

static int curr_seq = 0;

int ipmi_open(ipmi_ctx *ctx) {
    char ipmi_dev[16];
    char ipmi_devfs[16];
    char ipmi_devfs2[16];
    int fd, i, rv;

    if ( ctx == NULL )
        return -1;

    sprintf(ipmi_dev, "/dev/ipmi%d", 0);
    sprintf(ipmi_devfs, "/dev/ipmi/%d", 0);
    sprintf(ipmi_devfs2, "/dev/ipmidev/%d", 0);

    fd = open(ipmi_dev, O_RDWR);
    if ( fd < 0 ) {
        fd = open(ipmi_devfs, O_RDWR);
	if ( fd < 0 ) {
	     fd = open(ipmi_devfs2, O_RDWR);
	}
    }

    if ( fd < 0 ) {
        return errno;
    }


    rv = ioctl(fd, IPMICTL_SET_GETS_EVENTS_CMD, &i);
    if ( rv < 0 ) {
        return errno;
    }

    ctx->fd = fd;
    return 0;
}

int ipmi_send(ipmi_ctx *ctx, ipmi_rq *req, ipmi_rsp *resp) {
    int fd, rv;

    struct ipmi_system_interface_addr bmc_addr;
    struct ipmi_req _req;
    struct ipmi_recv recv;
    struct ipmi_addr addr;

    if ( ctx == NULL || req == NULL || resp == NULL ) {
        return -1;
    }

    if ( ctx->fd <= 0 ) {
        return -1;
    }

    bmc_addr.addr_type = IPMI_SYSTEM_INTERFACE_ADDR_TYPE;
    bmc_addr.channel = IPMI_BMC_CHANNEL;
    bmc_addr.lun = 0;
    
    memset(&_req, 0, sizeof(struct ipmi_req));
    _req.addr = (unsigned char *) &bmc_addr;
    _req.addr_len = sizeof(bmc_addr);
    _req.msgid = curr_seq++;
    _req.msg.netfn = (unsigned char)req->netfn;
    _req.msg.cmd = (unsigned char)req->cmd;
    _req.msg.data = (unsigned char *)req->data;
    _req.msg.data_len = (unsigned short)req->data_len;
    rv = ioctl(ctx->fd, IPMICTL_SEND_COMMAND, &_req);
    if ( rv < 0 ) {
        printf("IPMICTL_SEND_COMMAND Failed\n");
        return errno;
    }

    recv.addr = (unsigned char *) &addr;
    recv.addr_len = sizeof(addr);
    recv.msg.data =  resp->data;
    recv.msg.data_len = sizeof(resp->data);


    {
        fd_set rset;
        struct timeval rtimeout;
        FD_ZERO(&rset);
        FD_SET(ctx->fd, &rset);
        rtimeout.tv_sec = req->recv_timeout == 0 ? 2 : (int)req->recv_timeout;
        rtimeout.tv_usec = 0;
        rv = select(ctx->fd+1, &rset, NULL, NULL, &rtimeout);
        if ( rv < 0 ) {
            return errno;
        } else if ( rv == 0 ) {
            return -2;
        }

        if ( FD_ISSET(ctx->fd, &rset) == 0 ) {
            return -2;
        }
    }


    rv = ioctl(ctx->fd, IPMICTL_RECEIVE_MSG_TRUNC, &recv);
    if ( rv < 0 ) {
        printf("IPMICTL_RECEIVE_MSG_TRUNC Failed\n");
        return errno;
    }

    resp->data_len = (int)recv.msg.data_len;
    return 0;
}

void ipmi_close(ipmi_ctx *ctx) {
    if ( ctx == NULL ) return;
    if ( ctx->fd > 0 ) close(ctx->fd);
    ctx->fd = -1;
    return;
}


/*
void test_in_c() {
    ipmi_ctx ctx;
    ipmi_rq request;
    ipmi_rsp response;
    ipmi_open(&ctx);
    request.netfn = 0x0a;
    request.cmd = 0x20;
    request.lun = 0x00;
    request.data = NULL;
    request.data_len = 0;
    ipmi_send(&ctx, &request, &response); 
    ipmi_close(&ctx);
}
*/
