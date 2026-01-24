# Makefile for iclip_client

CC = gcc
CFLAGS = -Wall -Wextra -O2
TARGET = iclip_client
SRC = iclip_client.c

# For static linking (more portable)
STATIC_FLAGS = -static

.PHONY: all clean static

all: $(TARGET)

$(TARGET): $(SRC)
	$(CC) $(CFLAGS) -o $@ $<

# Build static binary (no dynamic library dependencies except libc)
static: $(SRC)
	$(CC) $(CFLAGS) $(STATIC_FLAGS) -o $(TARGET) $<

clean:
	rm -f $(TARGET)

install: $(TARGET)
	cp $(TARGET) /usr/local/bin/
	cp config.ini /usr/local/etc/iclip_client.ini

.PHONY: install
