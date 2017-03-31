all: hello

GIT_REV=$(shell git describe --tags --always --dirty)
build_info.c: hello.c
	@echo  'const char *build_git_sha = "$(GIT_REV)";' > $@
	@date +'const char *build_ts = "%Y-%m-%d %H:%M:%S";' >> $@

hello: hello.c build_info.c
	gcc $^ -o $@

clean:
	rm -rf hello build_info.c
