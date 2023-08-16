package main

/*
#include <stdlib.h>
#include <string.h>
#include <pwd.h>
#include <shadow.h>
#include <grp.h>
#include <gshadow.h>
static size_t sizeof_char = sizeof(char*);
*/
import "C"
import (
	"unsafe"

	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

func copyToBuffer(buffer *C.char, buflen C.size_t, parts ...string) ([]*C.char, C.size_t) {
	var writtenBytes C.size_t
	ptrs := make([]*C.char, len(parts))
	for i, part := range parts {
		cs := C.CString(part)
		strSize := C.sizeof_char * C.strlen(cs)
		allocSize := strSize + 1 // + NULL terminator
		if allocSize > buflen-writtenBytes {
			C.free(unsafe.Pointer(cs))
			return nil, writtenBytes
		}
		ptrs[i] = (*C.char)(unsafe.Add(unsafe.Pointer(buffer), int(writtenBytes)))
		C.memcpy(unsafe.Pointer(ptrs[i]), unsafe.Pointer(cs), strSize)    // write string
		C.memset(unsafe.Add(unsafe.Pointer(ptrs[i]), int(strSize)), 0, 1) // write NULL terminator
		C.free(unsafe.Pointer(cs))
		writtenBytes += allocSize
	}
	return ptrs, writtenBytes
}

func StoreUserInPasswdStruct(user *types.User, result *C.struct_passwd, buffer *C.char, buflen C.size_t) error {
	ptrs, _ := copyToBuffer(buffer, buflen, user.User, user.Passwd, user.Name, user.Dir, user.Shell)
	if len(ptrs) == 0 {
		return errors.New("out of memory")
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

func StoreUserInSpwdStruct(user *types.User, result *C.struct_spwd, buffer *C.char, buflen C.size_t) error {
	ptrs, _ := copyToBuffer(buffer, buflen, user.User, user.Passwd)
	if len(ptrs) == 0 {
		return errors.New("out of memory")
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

func StoreGroupInGroupStruct(group *types.Group, result *C.struct_group, buffer *C.char, buflen C.size_t) error {
	ptrs, writtenBytes := copyToBuffer(buffer, buflen, append([]string{group.Name, group.Passwd}, group.GroupMembers...)...)
	if len(ptrs) == 0 {
		return errors.New("out of memory")
	}
	// we also need to write a NULL terminator for the gr_mem vector
	if buflen-writtenBytes <= 0 {
		return errors.New("out of memory")
	}
	buf := (*C.char)(unsafe.Add(unsafe.Pointer(buffer), int(writtenBytes)))
	C.memset(unsafe.Pointer(buf), 0, 1) // write NULL

	result.gr_name = ptrs[0]
	result.gr_passwd = ptrs[1]
	result.gr_gid = C.uint(group.Gid)
	if len(group.GroupMembers) == 0 {
		result.gr_mem = &buf
	} else {
		result.gr_mem = &ptrs[2]
	}
	return nil
}

func StoreGroupInGShadowStruct(group *types.Group, result *C.struct_sgrp, buffer *C.char, buflen C.size_t) error {
	ptrs, _ := copyToBuffer(buffer, buflen, group.Name, group.Passwd)
	if len(ptrs) == 0 {
		return errors.New("out of memory")
	}
	result.sg_namp = ptrs[0]
	result.sg_passwd = ptrs[1]
	result.sg_adm = nil
	result.sg_mem = nil
	return nil
}
