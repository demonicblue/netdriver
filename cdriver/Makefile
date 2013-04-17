CC=c99 			#Compiler 
CFLAGS=-Wall -O3	#Flags	

ODIR=obj/%.o #Directory for .o-files
BDIR=bin

PROG = netdriver #Program-name

_OBJS = netdriver.c #Files to compile
OBJS = $(patsubst %,$(ODIR)%,$(_OBJS))

$(ODIR): 
	$(CC) -c -o $@ $< $(CFLAGS)
	
$(PROG): $(_OBJS)
	$(CC) -o $@ $^ $(CFLAGS)

.PHONY: clean

clean:
	rm -f $(ODIR)/*.o $(BDIR)
	
