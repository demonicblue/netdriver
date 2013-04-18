CC=c99 			#Compiler 
CFLAGS=-Wall -O3	#Flags	

vpath %.c src
vpath %.h src
vpath %.o obj

ODIR = obj

PROG = bin/netdriver #Program-name

CLIBS = nbd.h

_OBJS = $(ODIR)/netdriver.o #Files to compile
OBJS = $(patsubst %,$(ODIR)%/,$(_OBJS))

CFILES = netdriver.c

%.o: $(CFILES) $(CLIBS)
	$(CC) -c -o $@ $< $(CFLAGS)
	
$(PROG): $(_OBJS)
	$(CC) -o $@ $^ $(CFLAGS)

.PHONY: clean

clean:
	rm -f $(ODIR)/*.o $(PROG)
	
