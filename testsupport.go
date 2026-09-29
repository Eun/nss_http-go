package main

/*
#include <stdlib.h>
#include <pwd.h>
#include <shadow.h>
#include <grp.h>
#include <gshadow.h>
*/
import "C"

import (
	"unsafe"

	"github.com/Eun/nss_http/types"
)

// This file exposes the cgo struct-filling helpers to the test suite.
//
// Go does not allow cgo in _test.go files ("use of cgo in test ... not
// supported"), so the bridge has to live in a regular file. Everything here is
// unexported and only referenced by helpers_test.go.

// testBuf is a C buffer of the kind glibc hands to getpwnam_r and friends.
type testBuf struct {
	ptr  *C.char
	size C.size_t
}

func newTestBuf(size int) *testBuf {
	return &testBuf{ptr: (*C.char)(C.calloc(C.size_t(size), 1)), size: C.size_t(size)}
}

func (b *testBuf) free() { C.free(unsafe.Pointer(b.ptr)) }

// base is the buffer start as an integer, used to assert layout properties.
func (b *testBuf) base() uintptr { return uintptr(unsafe.Pointer(b.ptr)) }

// readVector walks a NULL terminated char* vector into a Go slice, exactly as
// a libc caller would consume gr_mem / sg_mem.
func readVector(v **C.char) []string {
	if v == nil {
		return nil
	}
	var out []string
	for i := 0; ; i++ {
		entry := *(**C.char)(unsafe.Add(unsafe.Pointer(v), uintptr(i)*uintptr(ptrSize)))
		if entry == nil {
			return out
		}
		out = append(out, C.GoString(entry))
	}
}

// pointerVectorSize is the size of one vector slot, for alignment assertions.
const pointerVectorSize = uintptr(ptrSize)

// testCopyToBuffer wraps copyToBuffer, returning the strings read back out of
// the buffer plus the number of bytes written.
func testCopyToBuffer(b *testBuf, parts ...string) ([]string, uint64, error) {
	ptrs, used, err := copyToBuffer(b.ptr, b.size, parts...)
	if err != nil {
		return nil, uint64(used), err
	}
	out := make([]string, len(ptrs))
	for i, p := range ptrs {
		out[i] = C.GoString(p)
	}
	return out, uint64(used), nil
}

// testWritePointerVector copies parts into the buffer and then writes the
// pointer vector after them, returning the vector contents and its address.
func testWritePointerVector(b *testBuf, usedBytes uint64, parts ...string) ([]string, uintptr, error) {
	var ptrs []*C.char
	if len(parts) > 0 {
		var err error
		ptrs, _, err = copyToBuffer(b.ptr, b.size, parts...)
		if err != nil {
			return nil, 0, err
		}
	}
	v, err := writePointerVector(b.ptr, b.size, C.size_t(usedBytes), ptrs)
	if err != nil {
		return nil, 0, err
	}
	return readVector(v), uintptr(unsafe.Pointer(v)), nil
}

// testCopyThenWriteVector is the realistic sequence: strings first, then the
// vector placed after them.
func testCopyThenWriteVector(b *testBuf, parts ...string) ([]string, uintptr, uint64, error) {
	ptrs, used, err := copyToBuffer(b.ptr, b.size, parts...)
	if err != nil {
		return nil, 0, uint64(used), err
	}
	v, err := writePointerVector(b.ptr, b.size, used, ptrs)
	if err != nil {
		return nil, 0, uint64(used), err
	}
	return readVector(v), uintptr(unsafe.Pointer(v)), uint64(used), nil
}

// storedGroup is the observable result of StoreGroupInGroupStruct.
type storedGroup struct {
	Name    string
	Passwd  string
	Gid     uint
	Members []string
	MemNil  bool
}

func testStoreGroup(b *testBuf, group *types.Group) (storedGroup, error) {
	var result C.struct_group
	if err := StoreGroupInGroupStruct(group, &result, b.ptr, b.size); err != nil {
		return storedGroup{}, err
	}
	return storedGroup{
		Name:    C.GoString(result.gr_name),
		Passwd:  C.GoString(result.gr_passwd),
		Gid:     uint(result.gr_gid),
		Members: readVector(result.gr_mem),
		MemNil:  result.gr_mem == nil,
	}, nil
}

// storedGShadow is the observable result of StoreGroupInGShadowStruct.
type storedGShadow struct {
	Name    string
	Passwd  string
	Members []string
	Admins  []string
	MemNil  bool
	AdmNil  bool
}

func testStoreGShadow(b *testBuf, group *types.Group) (storedGShadow, error) {
	var result C.struct_sgrp
	if err := StoreGroupInGShadowStruct(group, &result, b.ptr, b.size); err != nil {
		return storedGShadow{}, err
	}
	return storedGShadow{
		Name:    C.GoString(result.sg_namp),
		Passwd:  C.GoString(result.sg_passwd),
		Members: readVector(result.sg_mem),
		Admins:  readVector(result.sg_adm),
		MemNil:  result.sg_mem == nil,
		AdmNil:  result.sg_adm == nil,
	}, nil
}

// storedPasswd is the observable result of StoreUserInPasswdStruct.
type storedPasswd struct {
	Name   string
	Passwd string
	Gecos  string
	Dir    string
	Shell  string
	Uid    uint
	Gid    uint
}

func testStorePasswd(b *testBuf, user *types.User) (storedPasswd, error) {
	var result C.struct_passwd
	if err := StoreUserInPasswdStruct(user, &result, b.ptr, b.size); err != nil {
		return storedPasswd{}, err
	}
	return storedPasswd{
		Name:   C.GoString(result.pw_name),
		Passwd: C.GoString(result.pw_passwd),
		Gecos:  C.GoString(result.pw_gecos),
		Dir:    C.GoString(result.pw_dir),
		Shell:  C.GoString(result.pw_shell),
		Uid:    uint(result.pw_uid),
		Gid:    uint(result.pw_gid),
	}, nil
}

// storedSpwd is the observable result of StoreUserInSpwdStruct.
type storedSpwd struct {
	Name   string
	Passwd string
	LstChg int64
	Expire int64
}

func testStoreSpwd(b *testBuf, user *types.User) (storedSpwd, error) {
	var result C.struct_spwd
	if err := StoreUserInSpwdStruct(user, &result, b.ptr, b.size); err != nil {
		return storedSpwd{}, err
	}
	return storedSpwd{
		Name:   C.GoString(result.sp_namp),
		Passwd: C.GoString(result.sp_pwdp),
		LstChg: int64(result.sp_lstchg),
		Expire: int64(result.sp_expire),
	}, nil
}
