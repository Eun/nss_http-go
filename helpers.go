package main

/*
#include <stdlib.h>
#include <shadow.h>
#include <string.h>
#include <pwd.h>
#include <grp.h>
static size_t sizeof_char = sizeof(char*);
*/
import "C"
import (
	"unsafe"

	"github.com/Eun/nss_http/types"
	"github.com/pkg/errors"
)

func copyToBuffer(buffer *C.char, buflen C.size_t, parts ...string) []*C.char {
	buf := buffer
	ptrs := make([]*C.char, len(parts))
	for i, part := range parts {
		cs := C.CString(part)
		size := C.sizeof_char * (C.strlen(cs) + 1)
		if size > buflen {
			C.free(unsafe.Pointer(cs))
			return nil
		}
		ptrs[i] = buf
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(cs), size)
		C.free(unsafe.Pointer(cs))
		buf = (*C.char)(unsafe.Add(unsafe.Pointer(buf), size))
		buflen -= size
	}
	return ptrs
}

func StoreUserInPasswdStruct(user *types.User, result *C.struct_passwd, buffer *C.char, buflen C.size_t) error {
	ptrs := copyToBuffer(buffer, buflen, user.User, user.Passwd, user.Name, user.Dir, user.Shell)
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
	ptrs := copyToBuffer(buffer, buflen, user.User, user.Passwd)
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
	ptrs := copyToBuffer(buffer, buflen, group.Name, group.Passwd)
	if len(ptrs) == 0 {
		return errors.New("out of memory")
	}
	result.gr_name = ptrs[0]
	result.gr_passwd = ptrs[1]
	result.gr_gid = group.Gid
	return nil
}
