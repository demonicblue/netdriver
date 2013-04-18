#include <stdlib.h> // malloc()
#include <stdio.h> // fprintf()
#include <sys/socket.h> // socketpair()
#include <sys/ioctl.h> // ioctl()
#include <err.h> // err()
#include <errno.h> // err
#include <string.h> // strerror()
#include <fcntl.h> // open()
#include <unistd.h> // close(), fork()
#include <netinet/in.h>

#include "nbd.h"

#define DATASIZE 1024*1024*50
#define SERVER_SOCKET 0
#define CLIENT_SOCKET 1

#ifdef WORDS_BIGENDIAN
uint64_t ntohll(uint64_t x) {
  return x;
}
#else
uint64_t ntohll(uint64_t x) {
  return __builtin_bswap64(x);
}
#endif
#define htonll ntohll

static int write_all(int fd, const void *buf, size_t count) {
  int bytes_written;
  
  while(count > 0) {
    bytes_written = write(fd, buf, count);
    if(bytes_written <= 0) {
      err(1, "Could not write to the socket");
    }
    buf = (char *)buf + bytes_written;
    count -= bytes_written;
  }
  
  return 0;
}

static int read_all(int fd, void *buf, size_t count) {
  int bytes_read;
  
  while(count > 0) {
    bytes_read = read(fd, buf, count);
    if(bytes_read <= 0) {
      err(1, "Could not read from the socket");
    }
    buf = (char *)buf + bytes_read;
    count -= bytes_read;
  }
  
  return 0;
}

int main(int argc, char *argv[]) {
  
  int nbd_fd, socket_fd; // NBD-device file descriptor
  char *dev_path;
  int fd[2]; // Inter process comm. file descriptor
  struct nbd_request request;
  struct nbd_reply reply;
  void *data, *chunk;
  uint32_t len, bytes_read;
  uint64_t offset;
  
  data = malloc(DATASIZE); // Allocate RAM-disk.
  
  dev_path = argv[1]; // Device path
  
  socketpair(AF_UNIX, SOCK_STREAM, 0, fd); // Set up inter-process communication
  
  nbd_fd = open(dev_path, O_RDWR);
  if(nbd_fd == -1) {
    err(1, "Cannot open %s", dev_path);
  }
  
  ioctl(nbd_fd, NBD_SET_SIZE, DATASIZE);
  ioctl(nbd_fd, NBD_CLEAR_SOCK);
  
  
  if(!fork()) { // Creating a child process to act as the client
    close(fd[SERVER_SOCKET]); // We do not need this anymore
    socket_fd = fd[CLIENT_SOCKET];
    
    if(ioctl(nbd_fd, NBD_SET_SOCK, socket_fd) == -1) {
      err(1, "Cannot set client socket");
    }
    
    int err = ioctl(nbd_fd, NBD_DO_IT);
    fprintf(stderr, "nbd device terminated with code %d", err);
    if(err == -1) {
      fprintf(stderr, "%s\n", strerror(errno));
    }
    
    ioctl(nbd_fd, NBD_CLEAR_QUE);
    ioctl(nbd_fd, NBD_CLEAR_SOCK);
    
    exit(0);
  }
  
  // Server code below
  
  close(fd[CLIENT_SOCKET]);
  socket_fd = fd[SERVER_SOCKET];
  
  // Setting up reply packets
  reply.magic = htonl(NBD_REPLY_MAGIC);
  reply.error = htonl(0);
  
  while(1) {
    bytes_read = read(socket_fd, &request, sizeof(request));
    
    memcpy(reply.handle, request.handle, sizeof(request.handle));
    
    len = ntohl(request.len);
    offset = ntohll(request.from);
    
    if( request.magic != htonl(NBD_REQUEST_MAGIC)) {
      // error in transfer / io error
      err(1, "Data integrity check failed");
    }
    
    switch(ntohl(request.type)) {
      case NBD_CMD_READ:
        chunk = malloc(len + sizeof(struct nbd_reply));
        memcpy(chunk+sizeof(struct nbd_reply), (char *)data + offset, len);
        memcpy(chunk, &reply, sizeof(struct nbd_reply));
        write_all(socket_fd, chunk, len + sizeof(struct nbd_reply));
        free(chunk);
        break;
      case NBD_CMD_WRITE:
        chunk = malloc(len);
        read_all(socket_fd, chunk, len);
        memcpy((char *)data + offset, chunk, len);
        free(chunk);
        write_all(socket_fd, &reply, sizeof(struct nbd_reply));
        break;
      case NBD_CMD_DISC:
        return 0;
      case NBD_CMD_FLUSH:
        break;
      case NBD_CMD_TRIM:
        break;
      default:
        err(1, "Unexpected NBD command: %d", ntohl(request.type));
    }
  }
  
  
  
  
}