#ifndef WRAPPER_H
#define WRAPPER_H

#include "git2/sys/odb_backend.h"
#include "git2/sys/refdb_backend.h"
//#include "git2/odb_backend.h"
/*
typedef struct {
	git_odb_writepack parent;
	void *go_interface;
} go_odb_writepack;
*/
typedef struct{
	git_odb_backend parent;
	void *go_interface;
} go_odb_backend;

typedef struct{
	git_refdb_backend parent;
	void *go_interface;
} go_refdb_backend;

#endif
