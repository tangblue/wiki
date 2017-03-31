#include <stdio.h>

#include "build_info.h"

int main()
{
	printf("Rev: %s, build at %s\n", build_git_sha, build_ts);

	return 0;
}
