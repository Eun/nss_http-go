package main

// #include <stdlib.h>
// #include <nss.h>
// #include <shadow.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

//export _nss_http_setspent
func _nss_http_setspent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setspent")
	return C.NSS_STATUS_NOTFOUND
}

//export _nss_http_endspent
func _nss_http_endspent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endspent")
	return C.NSS_STATUS_NOTFOUND
}

//
//export _nss_http_getspent_r
func _nss_http_getspent_r(resultbuf *C.struct_spwd, buffer *C.char, buflen C.size_t, result **C.struct_spwd) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getspent_r")
	return C.NSS_STATUS_NOTFOUND
}

//export _nss_http_getspnam_r
func _nss_http_getspnam_r(name *C.char, result *C.struct_spwd, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getspnam_r")

	userName := C.GoString(name)
	user, err := getUser(types.NameIdentifier(userName))
	if err != nil {
		log.Err(err).Str("name", userName).Msg("unable to get user by name")
		return C.NSS_STATUS_UNAVAIL
	}
	if user == nil {
		log.Debug().Str("name", userName).Msg("user not found")
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("user", user).Msg("user found")

	// store everything in buffer
	if err := StoreUserInSpwdStruct(user, result, buffer, buflen); err != nil {
		log.Err(err).Str("name", userName).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}
