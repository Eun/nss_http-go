package main

// #include <stdlib.h>
// #include <errno.h>
// #include <nss.h>
// #include <grp.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var groupDatabase utils.Database[types.Group]

//export _nss_http_setgrent
func _nss_http_setgrent() (status C.enum_nss_status) {
	log.Trace().Msg("setgrent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("setgrent")
	return errorToNSSStatus(groupDatabase.Open())
}

//export _nss_http_endgrent
func _nss_http_endgrent() (status C.enum_nss_status) {
	log.Trace().Msg("endgrent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("endgrent")
	return errorToNSSStatus(groupDatabase.Close())
}

//
// //export _nss_http_getgrent
// func _nss_http_getgrent() *C.struct_group {
// 	log.Trace().Msg("getgrent")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getgrent")
// 	panic("not implemented")
// }
//

//export _nss_http_getgrent_r
func _nss_http_getgrent_r(resultBuf *C.struct_group, buffer *C.char, bufLen C.size_t, result **C.struct_group) (status C.enum_nss_status) {
	log.Trace().Msg("getgrent_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getgrent_r")
	*result = nil
	status = errorToNSSStatus(
		groupDatabase.GetEnt(
			func() ([]types.Group, error) {
				return getGroups(PlainRetrivalMode)
			},
			func(item *types.Group) error {
				return StoreGroupInGroupStruct(item, resultBuf, buffer, bufLen)
			},
		),
	)
	if status == C.NSS_STATUS_SUCCESS {
		*result = resultBuf
	}
	return status
}

// //export _nss_http_getgrgid
// func _nss_http_getgrgid(gid C.uint) *C.struct_group {
// 	log.Trace().Msg("getgrgid")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getgrgid")
// 	panic("not implemented")
// }

//export _nss_http_getgrgid_r
func _nss_http_getgrgid_r(gid C.uint, resultBuf *C.struct_group, buffer *C.char, bufLen C.size_t, result **C.struct_group) (status C.enum_nss_status) {
	log.Trace().Msg("getgrgid_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getgrgid_r")
	*result = nil
	goGid := uint(gid)
	group, err := getGroup(PlainRetrivalMode, types.GIDIdentifier(goGid))
	if err != nil {
		log.Err(err).Uint("gid", goGid).Msg("unable to get group by gid")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	if group == nil {
		log.Debug().Uint("gid", goGid).Msg("group not found")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("group", group).Msg("group found")

	// store everything in buffer
	if err := StoreGroupInGroupStruct(group, resultBuf, buffer, bufLen); err != nil {
		log.Err(err).Uint("gid", goGid).Msg("unable to store group in buffer")
		if errors.Is(err, utils.OutOfMemoryError{}) {
			SetErrNo(C.ERANGE)
			return C.NSS_STATUS_UNAVAIL
		}
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}

	*result = resultBuf
	SetErrNo(0)
	return C.NSS_STATUS_SUCCESS
}

// //export _nss_http_getgrnam
// func _nss_http_getgrnam(name *C.char) *C.struct_group {
// 	log.Trace().Msg("getgrnam")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getgrnam")
// 	panic("not implemented")
// }

//export _nss_http_getgrnam_r
func _nss_http_getgrnam_r(name *C.char, resultBuf *C.struct_group, buffer *C.char, bufLen C.size_t, result **C.struct_group) (status C.enum_nss_status) {
	log.Trace().Msg("getgrnam_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getgrnam_r")
	*result = nil
	groupName := C.GoString(name)
	group, err := getGroup(PlainRetrivalMode, types.NameIdentifier(groupName))
	if err != nil {
		log.Err(err).Str("name", groupName).Msg("unable to get group by name")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	if group == nil {
		log.Debug().Str("name", groupName).Msg("group not found")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("group", group).Msg("group found")

	// store everything in buffer
	if err := StoreGroupInGroupStruct(group, resultBuf, buffer, bufLen); err != nil {
		log.Err(err).Str("name", groupName).Msg("unable to store group in buffer")
		if errors.Is(err, utils.OutOfMemoryError{}) {
			SetErrNo(C.ERANGE)
			return C.NSS_STATUS_UNAVAIL
		}
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	*result = resultBuf
	SetErrNo(0)
	return C.NSS_STATUS_SUCCESS
}
