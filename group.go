package main

// #include <stdlib.h>
// #include <nss.h>
// #include <grp.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

var groupDatabase DatabaseState[types.Group]

//export _nss_http_setgrent
func _nss_http_setgrent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setgrent")
	if groupDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	groupDatabase.IsOpen = true
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_endgrent
func _nss_http_endgrent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endgrent")
	if !groupDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	groupDatabase.IsOpen = false
	groupDatabase.FetchedItems = false
	groupDatabase.Items = nil
	groupDatabase.ItemIndex = 0
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getgrent_r
func _nss_http_getgrent_r(grbuf *C.struct_group, buffer *C.char, buflen C.size_t) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getgrent_r")
	if !groupDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	if !groupDatabase.FetchedItems {
		var err error
		groupDatabase.Items, err = getGroups()
		if err != nil {
			log.Err(err).Msg("unable to get groups")
			return C.NSS_STATUS_UNAVAIL
		}
	}

	if groupDatabase.ItemIndex+1 > len(groupDatabase.Items) {
		return C.NSS_STATUS_NOTFOUND
	}

	// store everything in buffer
	if err := StoreGroupInGroupStruct(&groupDatabase.Items[groupDatabase.ItemIndex], grbuf, buffer, buflen); err != nil {
		log.Err(err).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}
	groupDatabase.ItemIndex++
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getgrnam_r
func _nss_http_getgrnam_r(name *C.char, result *C.struct_group, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getgrnam_r")

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

//export _nss_http_getgrgid_r
func _nss_http_getgrgid_r(gid C.uint, result *C.struct_group, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getgrgid_r")
	goGid := uint(gid)
	group, err := getGroup(types.GIDIdentifier(goGid))
	if err != nil {
		log.Err(err).Uint("gid", goGid).Msg("unable to get group by gid")
		return C.NSS_STATUS_UNAVAIL
	}
	if group == nil {
		log.Debug().Uint("gid", goGid).Msg("group not found")
		return C.NSS_STATUS_NOTFOUND
	}

	log.Debug().Any("group", group).Msg("group found")

	// store everything in buffer
	if err := StoreGroupInGroupStruct(group, result, buffer, buflen); err != nil {
		log.Err(err).Uint("gid", goGid).Msg("unable to store group in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}
