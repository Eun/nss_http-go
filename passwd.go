package main

// #include <stdlib.h>
// #include <errno.h>
// #include <nss.h>
// #include <pwd.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var passwdDatabase utils.Database[types.User]

//export _nss_http_setpwent
func _nss_http_setpwent() (status C.enum_nss_status) {
	log.Trace().Msg("setpwent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("setpwent")
	return errorToNSSStatus(passwdDatabase.Open())
}

//export _nss_http_endpwent
func _nss_http_endpwent() (status C.enum_nss_status) {
	log.Trace().Msg("endpwent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("endpwent")
	return errorToNSSStatus(passwdDatabase.Close())
}

//
// //export _nss_http_getpwent
// func _nss_http_getpwent() *C.struct_passwd {
// 	log.Trace().Msg("getpwent")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getpwent")
// 	panic("not implemented")
// }
//

//export _nss_http_getpwent_r
func _nss_http_getpwent_r(resultBuf *C.struct_passwd, buffer *C.char, bufLen C.size_t, result **C.struct_passwd) (status C.enum_nss_status) {
	log.Trace().Msg("getpwent_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getpwent_r")
	*result = nil
	status = errorToNSSStatus(
		passwdDatabase.GetEnt(
			func() ([]types.User, error) {
				return getUsers(PlainRetrivalMode)
			},
			func(item *types.User) error {
				return StoreUserInPasswdStruct(item, resultBuf, buffer, bufLen)
			},
		),
	)
	if status == C.NSS_STATUS_SUCCESS {
		*result = resultBuf
	}
	return status
}

// //export _nss_http_getpwuid
// func _nss_http_getpwuid(uid C.uint) *C.struct_passwd {
// 	log.Trace().Msg("getpwuid")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getpwuid")
// 	panic("not implemented")
// }

//export _nss_http_getpwuid_r
func _nss_http_getpwuid_r(uid C.uint, resultBuf *C.struct_passwd, buffer *C.char, bufLen C.size_t, result **C.struct_passwd) (status C.enum_nss_status) {
	log.Trace().Msg("getpwuid_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getpwuid_r")
	*result = nil
	goUid := uint(uid)
	user, err := getUser(PlainRetrivalMode, types.UIDIdentifier(goUid))
	if err != nil {
		log.Err(err).Uint("uid", goUid).Msg("unable to get user by uid")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	if user == nil {
		log.Debug().Uint("uid", goUid).Msg("user not found")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("user", user).Msg("user found")

	// store everything in buffer
	if err := StoreUserInPasswdStruct(user, resultBuf, buffer, bufLen); err != nil {
		log.Err(err).Uint("uid", goUid).Msg("unable to store user in buffer")
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

// //export _nss_http_getpwnam
// func _nss_http_getpwnam(name *C.char) *C.struct_passwd {
// 	log.Trace().Msg("getpwnam")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getpwnam")
// 	panic("not implemented")
// }

//export _nss_http_getpwnam_r
func _nss_http_getpwnam_r(name *C.char, resultBuf *C.struct_passwd, buffer *C.char, bufLen C.size_t, result **C.struct_passwd) (status C.enum_nss_status) {
	log.Trace().Msg("getpwnam_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getpwnam_r")
	*result = nil
	userName := C.GoString(name)
	user, err := getUser(PlainRetrivalMode, types.NameIdentifier(userName))
	if err != nil {
		log.Err(err).Str("name", userName).Msg("unable to get user by name")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_UNAVAIL
	}
	if user == nil {
		log.Debug().Str("name", userName).Msg("user not found")
		SetErrNo(C.ENOENT)
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("user", user).Msg("user found")

	// store everything in buffer
	if err := StoreUserInPasswdStruct(user, resultBuf, buffer, bufLen); err != nil {
		log.Err(err).Str("name", userName).Msg("unable to store user in buffer")
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
