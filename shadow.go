package main

// #include <stdlib.h>
// #include <errno.h>
// #include <nss.h>
// #include <shadow.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var shadowDatabase utils.Database[types.User]

//export _nss_http_setspent
func _nss_http_setspent() (status C.enum_nss_status) {
	log.Trace().Msg("setspent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("setspent")
	return errorToNSSStatus(shadowDatabase.Open())
}

//export _nss_http_endspent
func _nss_http_endspent() (status C.enum_nss_status) {
	log.Trace().Msg("endspent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("endspent")
	return errorToNSSStatus(shadowDatabase.Close())
}

// //export _nss_http_getspent
// func _nss_http_getspent() *C.struct_spwd {
// 	log.Trace().Msg("getspent")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getspent")
// 	panic("not implemented")
// }

//export _nss_http_getspent_r
func _nss_http_getspent_r(resultBuf *C.struct_spwd, buffer *C.char, bufLen C.size_t, result **C.struct_spwd) (status C.enum_nss_status) {
	log.Trace().Msg("getspent_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getspent_r")
	*result = nil
	status = errorToNSSStatus(
		shadowDatabase.GetEnt(
			func() ([]types.User, error) {
				return getUsers(ShadowRetrivalMode)
			},
			func(item *types.User) error {
				return StoreUserInSpwdStruct(item, resultBuf, buffer, bufLen)
			},
		),
	)
	if status == C.NSS_STATUS_SUCCESS {
		*result = resultBuf
	}
	return status
}

// //export _nss_http_getspnam
// func _nss_http_getspnam(name *C.char) *C.struct_spwd {
// 	log.Trace().Msg("getspnam")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getspnam")
// 	panic("not implemented")
// }

//export _nss_http_getspnam_r
func _nss_http_getspnam_r(name *C.char, resultBuf *C.struct_spwd, buffer *C.char, bufLen C.size_t, result **C.struct_spwd) (status C.enum_nss_status) {
	log.Trace().Msg("getspnam_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getspnam_r")
	*result = nil
	userName := C.GoString(name)
	user, err := getUser(ShadowRetrivalMode, types.NameIdentifier(userName))
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
	if err := StoreUserInSpwdStruct(user, resultBuf, buffer, bufLen); err != nil {
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
