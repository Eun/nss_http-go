package main

// #include <stdlib.h>
// #include <nss.h>
// #include <grp.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

//export _nss_http_setgrent
func _nss_http_setgrent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setgrent")
	return C.NSS_STATUS_NOTFOUND
}

//export _nss_http_endgrent
func _nss_http_endgrent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endgrent")
	return C.NSS_STATUS_NOTFOUND
}

//
//export _nss_http_getgrent_r
func _nss_http_getgrent_r(resultbuf *C.struct_group, buffer *C.char, buflen C.size_t, result **C.struct_group) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getgrent_r")
	return C.NSS_STATUS_NOTFOUND
}

//export _nss_http_getgrnam_r
func _nss_http_getgrnam_r(name *C.char, result *C.struct_group, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getspnam_r")

	groupName := C.GoString(name)
	group, err := getGroup(types.NameIdentifier(groupName))
	if err != nil {
		log.Err(err).Str("name", groupName).Msg("unable to get group by name")
		return C.NSS_STATUS_UNAVAIL
	}
	if group == nil {
		log.Debug().Str("name", groupName).Msg("group not found")
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("group", group).Msg("group found")

	// store everything in buffer
	if err := StoreGroupInGroupStruct(group, result, buffer, buflen); err != nil {
		log.Err(err).Str("name", groupName).Msg("unable to store group in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}
