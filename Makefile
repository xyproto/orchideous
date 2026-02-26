.PHONY: clean examples examples-clean install test

PROJECT ?= orchideous
GOFLAGS ?= -mod=vendor -trimpath -v -ldflags "-s -w" -buildvcs=false
GOBUILD := go build
SRCFILES := $(wildcard go.* *.go)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  PREFIX ?= /usr/local
  MAKE ?= make
  EXE_EXT :=
else ifneq (,$(findstring BSD,$(UNAME_S)))
  PREFIX ?= /usr/local
  MAKE ?= gmake
  EXE_EXT :=
else ifneq (,$(findstring MINGW,$(UNAME_S)))
  PREFIX ?= /usr
  MAKE ?= make
  EXE_EXT := .exe
else ifneq (,$(findstring MSYS,$(UNAME_S)))
  PREFIX ?= /usr
  MAKE ?= make
  EXE_EXT := .exe
else ifneq (,$(findstring CYGWIN,$(UNAME_S)))
  PREFIX ?= /usr
  MAKE ?= make
  EXE_EXT := .exe
else
  PREFIX ?= /usr
  MAKE ?= make
  EXE_EXT :=
endif

UNAME_R ?= $(shell uname -r)
ifneq (,$(findstring arch,$(UNAME_R)))
# Arch Linux
LDFLAGS ?= -Wl,-O2,--as-needed,-z,relro,-z,now
GOFLAGS += -buildmode=pie
BUILDFLAGS ?= -ldflags "-s -w -linkmode=external -extldflags $(LDFLAGS)"
endif

EXAMPLE_DIRS := $(sort $(dir $(wildcard examples/*/main.c examples/*/main.cpp examples/*/main.cc)))

orchideous: $(SRCFILES)
	$(GOBUILD) $(GOFLAGS) $(BUILDFLAGS) -o $@$(EXE_EXT) || $(GOBUILD) -o $@$(EXE_EXT)

examples: orchideous
	@failed=""; \
	for d in $(EXAMPLE_DIRS); do \
		printf "=== %-30s" "$$d"; \
		cd "$$d" && ../../orchideous$(EXE_EXT) > /dev/null 2>&1 && echo "OK" || { echo "FAIL"; failed="$$failed $$d"; }; \
		cd "$(CURDIR)"; \
	done; \
	if [ -n "$$failed" ]; then \
		echo ""; echo "Failed:$$failed"; exit 1; \
	fi

test:
	go test $(GOFLAGS) ./...

install: orchideous
	install -Dm755 orchideous$(EXE_EXT) "$(DESTDIR)$(PREFIX)/bin/orchideous$(EXE_EXT)"
	ln -sf orchideous$(EXE_EXT) "$(DESTDIR)$(PREFIX)/bin/oh$(EXE_EXT)"

examples-clean:
	@for d in $(EXAMPLE_DIRS); do \
		name=$$(basename "$$d"); \
		rm -f "$$d/$$name" "$$d/$$name.exe" "$$d"/*.o "$$d"/*.d; \
	done

clean: examples-clean
	-rm -f orchideous$(EXE_EXT)
