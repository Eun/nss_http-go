package main

// #include <stdlib.h>
// #include <nss.h>
// #include <shadow.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

var shadowDatabase DatabaseState[types.User]

//export _nss_http_setspent
func _nss_http_setspent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setspent")
	if shadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	shadowDatabase.IsOpen = true
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_endspent
func _nss_http_endspent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endspent")
	if !shadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	shadowDatabase.IsOpen = false
	shadowDatabase.FetchedItems = false
	shadowDatabase.Items = nil
	shadowDatabase.ItemIndex = 0
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getspent_r
func _nss_http_getspent_r(spbuf *C.struct_spwd, buffer *C.char, buflen C.size_t) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getspent_r")
	if !shadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	if !shadowDatabase.FetchedItems {
		var err error
		shadowDatabase.Items, err = getUsers()
		if err != nil {
			log.Err(err).Msg("unable to get users")
			return C.NSS_STATUS_UNAVAIL
		}
	}

	if shadowDatabase.ItemIndex+1 > len(shadowDatabase.Items) {
		return C.NSS_STATUS_NOTFOUND
	}

	// store everything in buffer
	if err := StoreUserInSpwdStruct(&shadowDatabase.Items[shadowDatabase.ItemIndex], spbuf, buffer, buflen); err != nil {
		log.Err(err).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}
	shadowDatabase.ItemIndex++
	return C.NSS_STATUS_SUCCESS
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
