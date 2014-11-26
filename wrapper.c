#include "_cgo_export.h"
#include <git2.h>
#include <git2/sys/odb_backend.h>
#include <git2/sys/refdb_backend.h>
#include "wrapper.h"
// #include "git2/submodule.h"
// #include "git2/pack.h"

typedef int (*gogit_submodule_cbk)(git_submodule *sm, const char *name, void *payload);

int _go_git_visit_submodule(git_repository *repo, void *fct)
{
	  return git_submodule_foreach(repo, (gogit_submodule_cbk)&SubmoduleVisitor, fct);
}

int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr)
{
	return git_tree_walk(tree, mode, (git_treewalk_cb)&CallbackGitTreeWalk, ptr);
}

int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload)
{
    return git_packbuilder_foreach(pb, (git_packbuilder_foreach_cb)&packbuilderForEachCb, payload);
}

int _go_git_odb_foreach(git_odb *db, void *payload)
{
    return git_odb_foreach(db, (git_odb_foreach_cb)&odbForEachCb, payload);
}

void _go_git_odb_backend_free(git_odb_backend *backend)
{
    if (backend->free)
      backend->free(backend);

    return;
}

void _go_git_refdb_backend_free(git_refdb_backend *backend)
{
    if (backend->free)
      backend->free(backend);

    return;
}

int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload)
{
	git_diff_file_cb fcb = NULL;
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile) {
		fcb = (git_diff_file_cb)&diffForEachFileCb;
	}

	if (eachHunk) {
		hcb = (git_diff_hunk_cb)&diffForEachHunkCb;
	}

	if (eachLine) {
		lcb = (git_diff_line_cb)&diffForEachLineCb;
	}

	return git_diff_foreach(diff, fcb, hcb, lcb, payload);
}

int _go_git_diff_blobs(git_blob *old, const char *old_path, git_blob *new, const char *new_path, git_diff_options *opts, int eachFile, int eachHunk, int eachLine, void *payload)
{
	git_diff_file_cb fcb = NULL;
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile) {
		fcb = (git_diff_file_cb)&diffForEachFileCb;
	}

	if (eachHunk) {
		hcb = (git_diff_hunk_cb)&diffForEachHunkCb;
	}

	if (eachLine) {
		lcb = (git_diff_line_cb)&diffForEachLineCb;
	}

	return git_diff_blobs(old, old_path, new, new_path, opts, fcb, hcb, lcb, payload);
}

void _go_git_setup_diff_notify_callbacks(git_diff_options *opts) {
	opts->notify_cb = (git_diff_notify_cb)diffNotifyCb;
}

void _go_git_setup_callbacks(git_remote_callbacks *callbacks) {
	typedef int (*completion_cb)(git_remote_completion_type type, void *data);
	typedef int (*update_tips_cb)(const char *refname, const git_oid *a, const git_oid *b, void *data);
	typedef int (*push_update_reference_cb)(const char *refname, const char *status, void *data);

	callbacks->sideband_progress = (git_transport_message_cb)sidebandProgressCallback;
	callbacks->completion = (completion_cb)completionCallback;
	callbacks->credentials = (git_cred_acquire_cb)credentialsCallback;
	callbacks->transfer_progress = (git_transfer_progress_cb)transferProgressCallback;
	callbacks->update_tips = (update_tips_cb)updateTipsCallback;
	callbacks->certificate_check = (git_transport_certificate_check_cb) certificateCheckCallback;
	callbacks->pack_progress = (git_packbuilder_progress) packProgressCallback;
	callbacks->push_transfer_progress = (git_push_transfer_progress) pushTransferProgressCallback;
	callbacks->push_update_reference = (push_update_reference_cb) pushUpdateReferenceCallback;
}

int _go_blob_chunk_cb(char *buffer, size_t maxLen, void *payload)
{
    return blobChunkCb(buffer, maxLen, payload);
}

int _go_git_blob_create_fromchunks(git_oid *id,
	git_repository *repo,
	const char *hintpath,
	void *payload)
{
    return git_blob_create_fromchunks(id, repo, hintpath, _go_blob_chunk_cb, payload);
}

int _go_git_index_add_all(git_index *index, const git_strarray *pathspec, unsigned int flags, void *callback) {
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb) &indexMatchedPathCallback : NULL;
	return git_index_add_all(index, pathspec, flags, cb, callback);
}

int _go_git_index_update_all(git_index *index, const git_strarray *pathspec, void *callback) {
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb) &indexMatchedPathCallback : NULL;
	return git_index_update_all(index, pathspec, cb, callback);
}

int _go_git_index_remove_all(git_index *index, const git_strarray *pathspec, void *callback) {
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb) &indexMatchedPathCallback : NULL;
	return git_index_remove_all(index, pathspec, cb, callback);
}

/*
 * Go Odb Backend
 */

typedef int (*odb_backend__read)(void **, size_t *, git_otype *, git_odb_backend *, const git_oid *);
typedef int (*odb_backend__read_prefix)(git_oid *, void **, size_t *, git_otype *, git_odb_backend *, const git_oid *, size_t);
typedef int (*odb_backend__read_header)(size_t *, git_otype *, git_odb_backend *, const git_oid *);
typedef int (*odb_backend__write)(git_odb_backend *, const git_oid *, const void *, size_t, git_otype);
//typedef int (*odb_backend__writestream)(git_odb_stream **, git_odb_backend *, size_t, git_otype);
//typedef int (*odb_backend__readstream)(git_odb_stream **, git_odb_backend *, const git_oid *);
typedef int (*odb_backend__exists)(git_odb_backend *, const git_oid *);
typedef int (*odb_backend__exists_prefix)(git_oid *, git_odb_backend *, const git_oid *, size_t);
typedef int (*odb_backend__refresh)(git_odb_backend *);
typedef int (*odb_backend__foreach)(git_odb_backend *, git_odb_foreach_cb, void*);
//typedef int (*odb_backend__writepack)(git_odb_writepack **, git_odb_backend *, git_odb *, git_transfer_progress_cb, void *);
typedef void (*odb_backend__free)(git_odb_backend *);

git_odb_backend *new_go_odb_backend(void *interface) {
	go_odb_backend *backend = calloc(1, sizeof(go_odb_backend));
	backend->parent.version = 1;
	backend->parent.read = (odb_backend__read) &_Go_odb_backend__read;
	backend->parent.read_prefix = (odb_backend__read_prefix) &_Go_odb_backend__read_prefix;
	backend->parent.read_header = (odb_backend__read_header) &_Go_odb_backend__read_header;
	backend->parent.write = (odb_backend__write) &_Go_odb_backend__write;
	//backend->parent.writestream = (odb_backend__writestream) &_Go_odb_backend__writestream;
	//backend->parent.readstream = (odb_backend__readstream) &_Go_odb_backend__readstream;
	backend->parent.exists = (odb_backend__exists) &_Go_odb_backend__exists;
	backend->parent.exists_prefix = (odb_backend__exists_prefix) &_Go_odb_backend__exists_prefix;
	backend->parent.refresh = (odb_backend__refresh) &_Go_odb_backend__refresh;
	backend->parent.foreach = (odb_backend__foreach) &_Go_odb_backend__foreach;
	//backend->parent.writepack = (odb_backend__writepack)&_Go_odb_backend__writepack;
	backend->parent.free = (odb_backend__free)&_Go_odb_backend__free;
	backend->go_interface = interface;
	return (git_odb_backend*)backend;
}

void *odb_backend_to_go_interface(git_odb_backend *backend) {
	return ((go_odb_backend*)backend)->go_interface;
}

/*
git_odb_writepack *new_go_odb_writepack(git_odb_backend *backend, void *self) {
	go_odb_writepack *writepack = calloc(1, sizeof(go_odb_writepack));
	writepack->parent.backend = backend;
	writepack->parent.free = &_Go_odb_backend__writepack_free;
	writepack->go_interface = self;
	return (git_odb_writepack*)writepack;
}
*/
/*
 * Go Refdb Backend
 */
typedef int (*refdb_backend__exists)(int*, git_refdb_backend*, const char*);
typedef int (*refdb_backend__lookup)(git_reference**, git_refdb_backend*, const char*);
//typedef int (*refdb_backend__iterator)(git_reference_iterator**, git_refdb_backend*, const char*);
typedef int (*refdb_backend__write)(git_refdb_backend*, const git_reference*, int, const git_signature*, const char*, const git_oid*, const char*);
typedef int (*refdb_backend__rename)(git_reference**, git_refdb_backend*, const char*, const char*, int, const git_signature*, const char*);
typedef int (*refdb_backend__del)(git_refdb_backend*, const char*, const git_oid*, const char*);
typedef int (*refdb_backend__compress)(git_refdb_backend*);
typedef int (*refdb_backend__has_log)(git_refdb_backend*, const char*);
typedef int (*refdb_backend__ensure_log)(git_refdb_backend*, const char*);
typedef void (*refdb_backend__free)(git_refdb_backend*);
//typedef int (*refdb_backend__reflog_read)(git_reflog**, git_refdb_backend*, const char*);
//typedef int (*refdb_backend__reflog_write)(git_refdb_backend*, git_reflog*);
//typedef int (*refdb_backend__reflog_rename)(git_refdb_backend*, const char*, const char*);
//typedef int (*refdb_backend__reflog_delete)(git_refdb_backend*, const char*);
typedef int (*refdb_backend__lock)(void**, git_refdb_backend*, const char*);
typedef int (*refdb_backend__unlock)(git_refdb_backend*, void*, int, int, const git_reference*, const git_signature*, const char*);

git_refdb_backend *new_go_refdb_backend(void *interface) {
	go_refdb_backend *backend = calloc(1, sizeof(go_refdb_backend));
	backend->parent.version = 1;
	backend->parent.exists = (refdb_backend__exists) &_Go_refdb_backend__exists;
	backend->parent.lookup = (refdb_backend__lookup) &_Go_refdb_backend__lookup;
	//backend->parent.iterator = (refdb_backend__iterator) &_Go_refdb_backend__iterator;
	backend->parent.write = (refdb_backend__write) &_Go_refdb_backend__write;
	backend->parent.rename = (refdb_backend__rename) &_Go_refdb_backend__rename;
	backend->parent.del = (refdb_backend__del) &_Go_refdb_backend__del;
	backend->parent.compress = (refdb_backend__compress) &_Go_refdb_backend__compress;
	backend->parent.has_log = (refdb_backend__has_log) &_Go_refdb_backend__has_log;
	backend->parent.ensure_log = (refdb_backend__ensure_log) &_Go_refdb_backend__ensure_log;
	backend->parent.free = (refdb_backend__free) &_Go_refdb_backend__free;
	//backend->parent.reflog_read = (refdb_backend__reflog_read) &_Go_refdb_backend__reflog_read;
	//backend->parent.reflog_write = (refdb_backend__reflog_write) &_Go_refdb_backend__reflog_write;
	//backend->parent.reflog_rename = (refdb_backend__reflog_rename) &_Go_refdb_backend__reflog_rename;
	//backend->parent.reflog_delete = (refdb_backend__reflog_delete) &_Go_refdb_backend__reflog_delete;
	backend->parent.lock = (refdb_backend__lock) &_Go_refdb_backend__lock;
	backend->parent.unlock = (refdb_backend__unlock) &_Go_refdb_backend__unlock;
	backend->go_interface = interface;
	return (git_refdb_backend*)backend;
}

void *refdb_backend_to_go_interface(git_refdb_backend *backend) {
	return ((go_refdb_backend*)backend)->go_interface;
}
/* EOF */
