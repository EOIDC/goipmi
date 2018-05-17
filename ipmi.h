#ifndef __IPMI_H__
#define __IPMI_H__


#define IPMI_BUF_SIZE 1024;

typedef struct ipmi_ctx {
    int fd; /* the opened fd to /dev/ipmi0 | /dev/ipmi/0 | /dev/ipmidev/0 */
} ipmi_ctx;

typedef struct ipmi_rq {
    unsigned char netfn;
    unsigned char lun;
    unsigned char cmd;
    unsigned char *data; 
    unsigned short data_len;
    unsigned char recv_timeout;
} ipmi_rq;

typedef struct ipmi_rsp {
    unsigned char data[1024];
    int data_len;
} ipmi_rsp;

int ipmi_open(ipmi_ctx *ctx);
int ipmi_send(ipmi_ctx *ctx, ipmi_rq *req, ipmi_rsp *resp); 
void ipmi_close(ipmi_ctx *ctx);



#endif

