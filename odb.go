package git

/*
#include <git2.h>
// #include <git2/errors.h>
// for memcpy
#include <string.h>

#include "wrapper.h"

extern int _go_git_odb_foreach(git_odb *db, void *payload);
extern void _go_git_odb_backend_free(git_odb_backend *backend);

//extern git_odb_writepack *new_go_odb_writepack(git_odb_backend *backend, void *self);
extern git_odb_backend *new_go_odb_backend(void *);
extern void *odb_backend_to_go_interface(git_odb_backend *);
*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type Odb struct {
	ptr *C.git_odb
}

type OdbBackend struct {
	ptr *C.git_odb_backend
}

/*
type OdbWritePack struct {
	ptr *C.git_odb_writepack
}
*/

func NewOdb() (odb *Odb, err error) {
	odb = new(Odb)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_new(&odb.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(odb, (*Odb).Free)
	return odb, nil
}

func NewOdbBackendFromC(ptr *C.git_odb_backend) *OdbBackend {
	return &OdbBackend{ptr}
}

func NewOdbBackendFromGo(goOdbBackend GoOdbBackend) *OdbBackend {
	p := unsafe.Pointer(&goOdbBackend)
	doNotGC[p] = struct{}{}
	return &OdbBackend{C.new_go_odb_backend(p)}
}

/*
func NewOdbWritePackFromC(ptr *C.git_odb_writepack) *OdbWritePack {
	return &OdbWritePack{ptr}
}

func NewOdbWritePackFromGo(goOdbWritePack GoOdbWritePack, backend *OdbBackend) *OdbWritePack {
	return &OdbWritePack{C.new_go_odb_writepack(backend.ptr, unsafe.Pointer(&goOdbWritePack))}
}

func (v *Odb) WritePack() *OdbWritePack {
	return nil
}
*/

func (v *Odb) AddBackend(backend *OdbBackend, priority int) (err error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_add_backend(v.ptr, backend.ptr, C.int(priority))
	if ret < 0 {
		backend.Free()
		return MakeGitError(ret)
	}
	return nil
}

func (v *Odb) Exists(oid *Oid) bool {
	ret := C.git_odb_exists(v.ptr, oid.toC())
	return ret != 0
}

func (v *Odb) Write(data []byte, otype ObjectType) (oid *Oid, err error) {
	oid = new(Oid)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_write(oid.toC(), v.ptr, unsafe.Pointer(hdr.Data), C.size_t(hdr.Len), C.git_otype(otype))

	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Odb) Read(oid *Oid) (obj *OdbObject, err error) {
	obj = new(OdbObject)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_read(&obj.ptr, v.ptr, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(obj, (*OdbObject).Free)
	return obj, nil
}

type OdbForEachCallback func(id *Oid) error

type foreachData struct {
	callback OdbForEachCallback
	err      error
}

//export odbForEachCb
func odbForEachCb(id *C.git_oid, handle unsafe.Pointer) int {
	data, ok := pointerHandles.Get(handle).(*foreachData)

	if !ok {
		panic("could not retrieve handle")
	}

	err := data.callback(newOidFromC(id))
	if err != nil {
		data.err = err
		return C.GIT_EUSER
	}

	return 0
}

func (v *Odb) ForEach(callback OdbForEachCallback) error {
	data := foreachData{
		callback: callback,
		err:      nil,
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	handle := pointerHandles.Track(&data)
	defer pointerHandles.Untrack(handle)

	ret := C._go_git_odb_foreach(v.ptr, handle)
	if ret == C.GIT_EUSER {
		return data.err
	} else if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// Hash determines the object-ID (sha1) of a data buffer.
func (v *Odb) Hash(data []byte, otype ObjectType) (oid *Oid, err error) {
	oid = new(Oid)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := unsafe.Pointer(header.Data)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_hash(oid.toC(), ptr, C.size_t(header.Len), C.git_otype(otype))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return oid, nil
}

// NewReadStream opens a read stream from the ODB. Reading from it will give you the
// contents of the object.
func (v *Odb) NewReadStream(id *Oid) (*OdbReadStream, error) {
	stream := new(OdbReadStream)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_open_rstream(&stream.ptr, v.ptr, id.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(stream, (*OdbReadStream).Free)
	return stream, nil
}

// NewWriteStream opens a write stream to the ODB, which allows you to
// create a new object in the database. The size and type must be
// known in advance
func (v *Odb) NewWriteStream(size int, otype ObjectType) (*OdbWriteStream, error) {
	stream := new(OdbWriteStream)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_open_wstream(&stream.ptr, v.ptr, C.size_t(size), C.git_otype(otype))
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(stream, (*OdbWriteStream).Free)
	return stream, nil
}

func (v *OdbBackend) Free() {
	C._go_git_odb_backend_free(v.ptr)
}

type OdbObject struct {
	ptr *C.git_odb_object
}

func (v *OdbObject) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_object_free(v.ptr)
}

func (object *OdbObject) Id() (oid *Oid) {
	return newOidFromC(C.git_odb_object_id(object.ptr))
}

func (object *OdbObject) Len() (len uint64) {
	return uint64(C.git_odb_object_size(object.ptr))
}

func (object *OdbObject) Data() (data []byte) {
	var c_blob unsafe.Pointer = C.git_odb_object_data(object.ptr)
	var blob []byte

	len := int(C.git_odb_object_size(object.ptr))

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&blob)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(c_blob)

	return blob
}

type OdbReadStream struct {
	ptr *C.git_odb_stream
}

// Read reads from the stream
func (stream *OdbReadStream) Read(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := (*C.char)(unsafe.Pointer(header.Data))
	size := C.size_t(header.Cap)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_read(stream.ptr, ptr, size)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	header.Len = int(ret)

	return len(data), nil
}

// Close is a dummy function in order to implement the Closer and
// ReadCloser interfaces
func (stream *OdbReadStream) Close() error {
	return nil
}

func (stream *OdbReadStream) Free() {
	runtime.SetFinalizer(stream, nil)
	C.git_odb_stream_free(stream.ptr)
}

type OdbWriteStream struct {
	ptr *C.git_odb_stream
	Id  Oid
}

// Write writes to the stream
func (stream *OdbWriteStream) Write(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := (*C.char)(unsafe.Pointer(header.Data))
	size := C.size_t(header.Len)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_write(stream.ptr, ptr, size)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return len(data), nil
}

// Close signals that all the data has been written and stores the
// resulting object id in the stream's Id field.
func (stream *OdbWriteStream) Close() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_finalize_write(stream.Id.toC(), stream.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (stream *OdbWriteStream) Free() {
	runtime.SetFinalizer(stream, nil)
	C.git_odb_stream_free(stream.ptr)
}

type GoOdbBackend interface {
	Read(oid *Oid) ([]byte, ObjectType, error)
	ReadPrefix(shortId []byte) ([]byte, ObjectType, *Oid, error)
	ReadHeader(oid *Oid) (int, ObjectType, error)
	Write(oid *Oid, buf []byte, oType ObjectType) error
	//WriteStream(...) error
	//ReadStream(...) error
	Exists(oid *Oid) bool
	ExistsPrefix(shortId []byte) (*Oid, bool)
	Refresh() error
	ForEach(cb OdbForEachCallback) error
	//WritePack(odb *Odb) GoOdbWritePack
	Free()
}

/*
type GoOdbWritePack interface {
	Append() error
	Commit() error
	Free()
}
*/

//export _Go_odb_backend__read
func _Go_odb_backend__read(data_p *unsafe.Pointer, len_p *C.size_t, type_p *C.git_otype, _backend *C.git_odb_backend, coid *C.git_oid) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	data, oType, err := backend.Read(newOidFromC(coid))
	if err != nil {
		return handleError(err)
	}
	*len_p = (C.size_t)(len(data))
	*data_p = C.git_odb_backend_malloc(_backend, C.size_t(unsafe.Sizeof(data_p)*uintptr(len(data))))
	if len(data) > 0 {
		C.memcpy(*data_p, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	}
	*type_p = (C.git_otype)(oType)
	return C.GIT_OK
}

//export _Go_odb_backend__read_prefix
func _Go_odb_backend__read_prefix(oid *C.git_oid, data_p *unsafe.Pointer, len_p *C.size_t, type_p *C.git_otype, _backend *C.git_odb_backend, coid *C.git_oid, size C.size_t) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	data, oType, id, err := backend.ReadPrefix(C.GoBytes(unsafe.Pointer(coid), C.int(size)))
	if err != nil {
		return handleError(err)
	}
	*oid = *id.toC()
	*len_p = (C.size_t)(len(data))
	*data_p = C.git_odb_backend_malloc(_backend, C.size_t(unsafe.Sizeof(data_p)*uintptr(len(data))))
	if len(data) > 0 {
		C.memcpy(*data_p, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	}
	*type_p = (C.git_otype)(oType)
	return C.GIT_OK
}

//export _Go_odb_backend__read_header
func _Go_odb_backend__read_header(len_p *C.size_t, type_p *C.git_otype, _backend *C.git_odb_backend, coid *C.git_oid) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	size, oType, err := backend.ReadHeader(newOidFromC(coid))
	if err != nil {
		return handleError(err)
	}
	*len_p = (C.size_t)(size)
	*type_p = (C.git_otype)(oType)
	return C.GIT_OK
}

//export _Go_odb_backend__write
func _Go_odb_backend__write(_backend *C.git_odb_backend, coid *C.git_oid, cdata unsafe.Pointer, size C.size_t, otype C.git_otype) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	if err := backend.Write(newOidFromC(coid), C.GoBytes(cdata, C.int(size)), ObjectType(otype)); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_odb_backend__exists
func _Go_odb_backend__exists(_backend *C.git_odb_backend, coid *C.git_oid) C.int {
	p := (C.odb_backend_to_go_interface(_backend))
	backend := *(*GoOdbBackend)(p)
	backend.Free()
	if !backend.Exists(newOidFromC(coid)) {
		return 0
	}
	return 1
}

//export _Go_odb_backend__exists_prefix
func _Go_odb_backend__exists_prefix(oid *C.git_oid, _backend *C.git_odb_backend, coid *C.git_oid, size C.size_t) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	id, ok := backend.ExistsPrefix(C.GoBytes(unsafe.Pointer(coid), C.int(size)))
	if !ok {
		return 0
	}
	*oid = *id.toC()
	return 1
}

//export _Go_odb_backend__refresh
func _Go_odb_backend__refresh(_backend *C.git_odb_backend) C.int {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	if err := backend.Refresh(); err != nil {
		return handleError(err)
	}
	return C.GIT_OK
}

//export _Go_odb_backend__foreach
func _Go_odb_backend__foreach(_backend *C.git_odb_backend, cb C.git_odb_foreach_cb, payload unsafe.Pointer) C.int {
	data := (*foreachData)(payload)
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	if err := backend.ForEach(data.callback); err != nil {
		data.err = err
		return C.GIT_EUSER
	}
	return C.GIT_OK
}

/*
//export _Go_odb_backend__writepack
func _Go_odb_backend__writepack(out **C.git_odb_writepack, _backend *C.git_odb_backend, odb *C.git_odb, progress_cb C.git_transfer_progress_cb, progress_payload unsafe.Pointer) int {
	backend := *(*GoOdbBackend)(unsafe.Pointer(_backend))
	writepack := backend.WritePack(nil)
	*out = NewOdbWritePackFromGo(writepack, &OdbBackend{_backend}).ptr
	return C.GIT_OK
}
*/

//export _Go_odb_backend__free
func _Go_odb_backend__free(_backend *C.git_odb_backend) {
	backend := *(*GoOdbBackend)(C.odb_backend_to_go_interface(_backend))
	backend.Free()
	C.free(unsafe.Pointer(_backend))
}

/*
//export _Go_odb_backend__writepack_free
func _Go_odb_backend__writepack_free(_writepack *C.git_odb_writepack) {
	writepack := *(*GoOdbWritePack)(unsafe.Pointer(_writepack))
	writepack.Free()
	C.free(unsafe.Pointer(_writepack))
}
*/
