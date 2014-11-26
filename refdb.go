package git

/*
#include <git2.h>
#include <git2/sys/refdb_backend.h>

extern void _go_git_refdb_backend_free(git_refdb_backend *backend);

extern git_refdb_backend *new_go_refdb_backend(void *);
extern void *refdb_backend_to_go_interface(git_refdb_backend *);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type GoRefdbBackend interface {
	Repository() *Repository

	Exists(refName string) (bool, error)
	Lookup(refName string) (*Reference, error)
	//Iterator(glob string) (<-chan *Reference, error)
	Write(ref *Reference, force bool, who *Signature, message string, oldId *Oid, oldTarget string) error
	Rename(oldName, newName string, force bool, who *Signature, message string) (*Reference, error)
	Del(refName string, oldId *Oid, oldTarget string) error
	Compress() error
	HasLog(refName string) bool
	EnsureLog(refName string) error
	Free()
	//ReflogRead() error
	//ReflogWrite() error
	//ReflogRename() error
	//ReflogDelete() error
	Lock(refName string) (interface{}, error)
	Unlock(payload interface{}, success, updateReflog bool, ref *Reference, sig *Signature, message string) error
}

func NewRefdbBackendFromGo(goRefdbBackend GoRefdbBackend) *RefdbBackend {
	p := unsafe.Pointer(&goRefdbBackend)
	doNotGC[p] = struct{}{}
	return &RefdbBackend{C.new_go_refdb_backend(p)}
}

//export _Go_refdb_backend__exists
func _Go_refdb_backend__exists(exists *C.int, _backend *C.git_refdb_backend, ref_name *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	ok, err := backend.Exists(C.GoString(ref_name))
	if err != nil {
		return handleError(err)
	}
	*exists = cbool(ok)
	return C.GIT_OK
}

//export _Go_refdb_backend__lookup
func _Go_refdb_backend__lookup(out **C.git_reference, _backend *C.git_refdb_backend, ref_name *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	ref, err := backend.Lookup(C.GoString(ref_name))
	if err != nil {
		return handleError(err)
	}
	*out = ref.ptr
	return C.GIT_OK
}

/*
//export _Go_refdb_backend__iterator
func _Go_refdb_backend__iterator(iter **C.git_reference_iterator, _backend *C.git_refdb_backend, glob *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	// newreferenceiterator is probably wrong, it could end up in an infinite loop if it calls the backend's iterator which is this function.
	iterator, err := backend.Repository().NewReferenceIterator()
	if err != nil {
		return handleError(err)
	}
	iterCh, err := backend.Iterator(C.GoString(glob))
	if err != nil {
		return handleError(err)
	}
	*iter = iterator.ptr
	go func() {
		ref, err := iterator.Next()
		if err != nil {
			if gitErr, ok := err.(*GitError); !ok || gitErr.Code != ErrIterOver {
				panic(err)
			}
			// end of iteration
			close(iterCh)
		}
		iterCh <- ref
	}()
	return C.GIT_OK
}
*/

//export _Go_refdb_backend__write
func _Go_refdb_backend__write(_backend *C.git_refdb_backend, ref *C.git_reference, force C.int, who *C.git_signature, message *C.char, old *C.git_oid, old_target *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if err := backend.Write(newReferenceFromC(ref, backend.Repository()), gobool(force), newSignatureFromC(who), C.GoString(message), newOidFromC(old), C.GoString(old_target)); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_refdb_backend__rename
func _Go_refdb_backend__rename(out **C.git_reference, _backend *C.git_refdb_backend, old_name *C.char, new_name *C.char, force C.int, who *C.git_signature, message *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	ref, err := backend.Rename(C.GoString(old_name), C.GoString(new_name), gobool(force), newSignatureFromC(who), C.GoString(message))
	if err != nil {
		return handleError(err)
	}
	*out = ref.ptr
	return C.GIT_OK
}

//export _Go_refdb_backend__del
func _Go_refdb_backend__del(_backend *C.git_refdb_backend, ref_name *C.char, old_id *C.git_oid, old_target *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if err := backend.Del(C.GoString(ref_name), newOidFromC(old_id), C.GoString(old_target)); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_refdb_backend__compress
func _Go_refdb_backend__compress(_backend *C.git_refdb_backend) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if err := backend.Compress(); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_refdb_backend__has_log
func _Go_refdb_backend__has_log(_backend *C.git_refdb_backend, refname *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if backend.HasLog(C.GoString(refname)) {
		return 1
	}
	return 0
}

//export _Go_refdb_backend__ensure_log
func _Go_refdb_backend__ensure_log(_backend *C.git_refdb_backend, refname *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if err := backend.EnsureLog(C.GoString(refname)); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_refdb_backend__free
func _Go_refdb_backend__free(_backend *C.git_refdb_backend) {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	backend.Free()
	C.free(unsafe.Pointer(_backend))
}

//export _Go_refdb_backend__lock
func _Go_refdb_backend__lock(payload_out *unsafe.Pointer, _backend *C.git_refdb_backend, refname *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	payload, err := backend.Lock(C.GoString(refname))
	if err != nil {
		return handleError(err)
	}
	*payload_out = unsafe.Pointer(&lockData{payload})
	return C.GIT_OK
}

//export _Go_refdb_backend__unlock
func _Go_refdb_backend__unlock(_backend *C.git_refdb_backend, payload unsafe.Pointer, success, update_reflog C.int, ref *C.git_reference, sig *C.git_signature, message *C.char) C.int {
	backend := *(*GoRefdbBackend)(C.refdb_backend_to_go_interface(_backend))
	if err := backend.Unlock((*lockData)(payload).payload, gobool(success), gobool(update_reflog), newReferenceFromC(ref, backend.Repository()), newSignatureFromC(sig), C.GoString(message)); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

type lockData struct {
	payload interface{}
}

type Refdb struct {
	ptr *C.git_refdb
}

type RefdbBackend struct {
	ptr *C.git_refdb_backend
}

func (v *Repository) NewRefdb() (refdb *Refdb, err error) {
	refdb = new(Refdb)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refdb_new(&refdb.ptr, v.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(refdb, (*Refdb).Free)
	return refdb, nil
}

func NewRefdbBackendFromC(ptr *C.git_refdb_backend) (backend *RefdbBackend) {
	backend = &RefdbBackend{ptr}
	return backend
}

func (v *Refdb) SetBackend(backend *RefdbBackend) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_refdb_set_backend(v.ptr, backend.ptr)
	if ret < 0 {
		backend.Free()
		return MakeGitError(ret)
	}
	return nil
}

func (v *RefdbBackend) Free() {
	C._go_git_refdb_backend_free(v.ptr)
}
