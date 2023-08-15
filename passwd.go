package main

// #include <stdlib.h>
// #include <nss.h>
// #include <shadow.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

var passwdDatabase DatabaseState[types.User]

//export _nss_http_setpwent
func _nss_http_setpwent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setpwent")
	if passwdDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	passwdDatabase.IsOpen = true
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_endpwent
func _nss_http_endpwent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endpwent")
	if !passwdDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	passwdDatabase.IsOpen = false
	passwdDatabase.FetchedItems = false
	passwdDatabase.Items = nil
	passwdDatabase.ItemIndex = 0
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getpwent_r
func _nss_http_getpwent_r(pwbuf *C.struct_passwd, buffer *C.char, buflen C.size_t, pwbufp **C.struct_passwd) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getpwent_r")
	if !passwdDatabase.IsOpen {
		*pwbufp = nil
		return C.NSS_STATUS_UNAVAIL
	}
	if !passwdDatabase.FetchedItems {
		var err error
		passwdDatabase.Items, err = getUsers()
		if err != nil {
			log.Err(err).Msg("unable to get users")
			*pwbufp = nil
			return C.NSS_STATUS_UNAVAIL
		}
	}

	if passwdDatabase.ItemIndex+1 >= len(passwdDatabase.Items) {
		*pwbufp = nil
		return C.NSS_STATUS_NOTFOUND
	}

	// store everything in buffer
	if err := StoreUserInPasswdStruct(&passwdDatabase.Items[passwdDatabase.ItemIndex], pwbuf, buffer, buflen); err != nil {
		log.Err(err).Msg("unable to store user in buffer")
		*pwbufp = nil
		return C.NSS_STATUS_UNAVAIL
	}
	passwdDatabase.ItemIndex++
	*pwbufp = pwbuf
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getpwnam_r
func _nss_http_getpwnam_r(name *C.char, result *C.struct_passwd, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getpwnam_r")

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
	if err := StoreUserInPasswdStruct(user, result, buffer, buflen); err != nil {
		log.Err(err).Str("name", userName).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getpwuid_r
func _nss_http_getpwuid_r(uid C.uint, result *C.struct_passwd, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("nss_http_getpwuid_r")
	goUid := uint(uid)
	user, err := getUser(types.UIDIdentifier(goUid))
	if err != nil {
		log.Err(err).Uint("uid", goUid).Msg("unable to get user by uid")
		return C.NSS_STATUS_UNAVAIL
	}
	if user == nil {
		log.Debug().Uint("uid", goUid).Msg("user not found")
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("user", user).Msg("user found")

	// store everything in buffer
	if err := StoreUserInPasswdStruct(user, result, buffer, buflen); err != nil {
		log.Err(err).Uint("uid", goUid).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}
