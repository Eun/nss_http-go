package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <nss.h>
#include <pwd.h>
#include <shadow.h>
#include <grp.h>
#include <gshadow.h>
static size_t sizeof_char = sizeof(char*);

void setErrno(int err) {
    errno = err;
}

int getErrno() {
	return errno;
}
*/
import "C"
import (
	"unsafe"

	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/pkg/errors"
)

const ptrSize = C.size_t(unsafe.Sizeof(uintptr(0)))

func errorToNSSStatus(err error) C.enum_nss_status {
	if errors.Is(err, &utils.DatabaseUnavailableError{}) {
		C.setErrno(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	if errors.Is(err, &utils.DatabaseOutOfMemoryError{}) {
		C.setErrno(C.ERANGE)
		return C.NSS_STATUS_TRYAGAIN
	}
	if errors.Is(err, &utils.DatabaseEOF{}) {
		C.setErrno(0)
		return C.NSS_STATUS_NOTFOUND
	}
	// OutOfMemoryError is returned by value and has no custom Is method, so
	// the target must be a value too. Matching against &OutOfMemoryError{}
	// never succeeded and fell through to NSS_STATUS_SUCCESS below, which
	// reports a truncated entry as a valid one.
	if errors.Is(err, utils.OutOfMemoryError{}) {
		C.setErrno(C.ERANGE)
		return C.NSS_STATUS_TRYAGAIN
	}
	C.setErrno(0)
	return C.NSS_STATUS_SUCCESS
}

func SetErrNo(n C.int) {
	C.setErrno(n)
}

func GetErrNo() int {
	return int(C.getErrno())
}

// copyToBuffer copies all strings that are passed in to the given buffer,
// it returns the written bytes and the pointers in the buffer of every string given,
// e.g. when you write
//
//	"joe", "alice"
//
// the result buffer would be filled with
//
//	6a 6f 65 00 61 6c 69 63 65 00 |joe.alice.|
//
// where the function would return the addresses pointing to 6a, and 61
// notice that each string is NULL terminated
func copyToBuffer(buffer *C.char, bufsize C.size_t, parts ...string) ([]*C.char, C.size_t, error) {
	var writtenBytes C.size_t
	ptrs := make([]*C.char, len(parts))
	for i, part := range parts {
		// create a cstring representation
		cs := C.CString(part)
		// determinate how many bytes we need to copy
		strSize := C.sizeof_char * C.strlen(cs)
		allocSize := strSize + 1 // string size + NULL terminator
		if allocSize > bufsize-writtenBytes {
			// out of memory
			C.free(unsafe.Pointer(cs))
			return nil, writtenBytes, utils.OutOfMemoryError{}
		}
		// set the result buffer, it is always the start of the buffer + the written bytes
		ptrs[i] = (*C.char)(unsafe.Add(unsafe.Pointer(buffer), int(writtenBytes)))
		// copy the string
		C.memcpy(unsafe.Pointer(ptrs[i]), unsafe.Pointer(cs), strSize)
		// write NULL terminator at end of string
		C.memset(unsafe.Add(unsafe.Pointer(ptrs[i]), int(strSize)), 0, 1)
		// free cstring representation
		C.free(unsafe.Pointer(cs))
		// add allocSize to writtenBytes so in the next iteration we can use
		// writtenBytes as the base offset
		writtenBytes += allocSize
	}
	return ptrs, writtenBytes, nil
}

func StoreUserInPasswdStruct(user *types.User, result *C.struct_passwd, buffer *C.char, bufsize C.size_t) error {
	ptrs, _, err := copyToBuffer(buffer, bufsize, user.User, user.Passwd, user.Name, user.Dir, user.Shell)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(ptrs) != 5 {
		return errors.New("fatal error: expected 5 pointers in result")
	}
	result.pw_name = ptrs[0]
	result.pw_passwd = ptrs[1]
	result.pw_gecos = ptrs[2]
	result.pw_dir = ptrs[3]
	result.pw_shell = ptrs[4]
	result.pw_uid = C.uint(user.Uid)
	result.pw_gid = C.uint(user.Gid)
	return nil
}

func StoreUserInSpwdStruct(user *types.User, result *C.struct_spwd, buffer *C.char, bufsize C.size_t) error {
	ptrs, _, err := copyToBuffer(buffer, bufsize, user.User, user.Passwd)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(ptrs) != 2 {
		return errors.New("fatal error: expected 2 pointers in result")
	}
	result.sp_namp = ptrs[0]
	result.sp_pwdp = ptrs[1]
	result.sp_lstchg = -1
	result.sp_min = -1
	result.sp_max = -1
	result.sp_warn = -1
	result.sp_inact = -1
	result.sp_expire = -1
	result.sp_flag = 0
	return nil
}

// writePointerVector writes a NULL terminated vector of char* into buffer,
// starting at the first pointer aligned offset at or after usedBytes.
//
// The vector has to live inside the caller supplied buffer: NSS callers such
// as glibc's getgrnam_r only keep the buffer alive, so pointing gr_mem at Go
// memory (or at a Go slice's backing array) would hand out a dangling pointer
// as soon as the Go GC moves or collects it.
//
// The returned pointer is suitable for struct group's gr_mem member.
func writePointerVector(buffer *C.char, bufsize, usedBytes C.size_t, ptrs []*C.char) (**C.char, error) {
	base := unsafe.Pointer(buffer)
	align := uintptr(ptrSize)

	// Align the start of the vector. Storing a pointer through a misaligned
	// address is undefined behaviour and faults on strict alignment
	// architectures (e.g. some arm/sparc configurations).
	start := unsafe.Add(base, int(usedBytes))
	if rem := uintptr(start) % align; rem != 0 {
		start = unsafe.Add(start, int(align-rem))
	}

	// len(ptrs) entries plus the NULL terminator.
	offset := uintptr(start) - uintptr(base)
	need := (uintptr(len(ptrs)) + 1) * align
	if offset+need > uintptr(bufsize) {
		return nil, utils.OutOfMemoryError{}
	}

	for i, p := range ptrs {
		slot := (**C.char)(unsafe.Add(start, int(uintptr(i)*align)))
		*slot = p
	}
	// NULL terminate the vector.
	terminator := (**C.char)(unsafe.Add(start, int(uintptr(len(ptrs))*align)))
	*terminator = nil

	return (**C.char)(start), nil
}

func StoreGroupInGroupStruct(group *types.Group, result *C.struct_group, buffer *C.char, bufsize C.size_t) error {
	strs := append([]string{group.Name, group.Passwd}, group.GroupMembers...)
	ptrs, writtenBytes, err := copyToBuffer(buffer, bufsize, strs...)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(ptrs) != len(strs) {
		return errors.Errorf("fatal error: expected %d pointers in result", len(strs))
	}

	// ptrs[0] and ptrs[1] are gr_name and gr_passwd, everything after that
	// belongs to the gr_mem vector.
	members, err := writePointerVector(buffer, bufsize, writtenBytes, ptrs[2:])
	if err != nil {
		return errors.WithStack(err)
	}

	result.gr_name = ptrs[0]
	result.gr_passwd = ptrs[1]
	result.gr_gid = C.uint(group.Gid)
	result.gr_mem = members
	return nil
}

func StoreGroupInGShadowStruct(group *types.Group, result *C.struct_sgrp, buffer *C.char, bufsize C.size_t) error {
	strs := append([]string{group.Name, group.Passwd}, group.GroupMembers...)
	ptrs, writtenBytes, err := copyToBuffer(buffer, bufsize, strs...)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(ptrs) != len(strs) {
		return errors.Errorf("fatal error: expected %d pointers in result", len(strs))
	}

	// sg_adm and sg_mem must both be valid NULL terminated vectors rather
	// than NULL: callers such as getsgnam_r iterate them unconditionally.
	members, err := writePointerVector(buffer, bufsize, writtenBytes, ptrs[2:])
	if err != nil {
		return errors.WithStack(err)
	}
	// An empty administrator vector, placed after the member vector.
	admUsed := (uintptr(unsafe.Pointer(members)) - uintptr(unsafe.Pointer(buffer))) +
		(uintptr(len(ptrs[2:]))+1)*uintptr(ptrSize)
	admins, err := writePointerVector(buffer, bufsize, C.size_t(admUsed), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	result.sg_namp = ptrs[0]
	result.sg_passwd = ptrs[1]
	result.sg_adm = admins
	result.sg_mem = members
	return nil
}
