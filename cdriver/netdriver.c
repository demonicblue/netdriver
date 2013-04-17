#include <stdlib.h> // malloc()
#include <stdio.h> // fprintf()
#include <sys/socket.h> // socketpair()
#include <sys/ioctl.h> // ioctl()
#include <err.h> // err()
#include <errno.h> // err
#include <string.h> // strerror()
#include <fcntl.h> // open()
#include <unistd.h> // close(), fork()

#include "nbd.h"

#define DATASIZE 1024*1024*50
#define SERVER_SOCKET 0
#define CLIENT_SOCKET 1

int main(int argc, char *argv[]) {
  
  int nbd_fd; // NBD-device file descriptor
  char *dev_path;
  int fd[2]; // Inter process comm. file descriptor
  struct nbd_request request;
  struct nbd_reply reply;
  void* data;
  
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
    
    if(ioctl(nbd_fd, NBD_SET_SOCK, fd[CLIENT_SOCKET]) == -1) {
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
  
  
  
  
  
  
}