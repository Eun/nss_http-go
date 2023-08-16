package main

// #include <stdlib.h>
// #include <nss.h>
// #include <grp.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/rs/zerolog/log"
)

var gshadowDatabase DatabaseState[types.Group]

//export _nss_http_setsgent
func _nss_http_setsgent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_setsgent")
	if gshadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	gshadowDatabase.IsOpen = true
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_endsgent
func _nss_http_endsgent() C.enum_nss_status {
	log.Debug().Msg("_nss_http_endsgent")
	if !gshadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	gshadowDatabase.IsOpen = false
	gshadowDatabase.FetchedItems = false
	gshadowDatabase.Items = nil
	gshadowDatabase.ItemIndex = 0
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getsgent_r
func _nss_http_getsgent_r(sgbuf *C.struct_sgrp, buffer *C.char, buflen C.size_t) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getsgent_r")
	if !gshadowDatabase.IsOpen {
		return C.NSS_STATUS_UNAVAIL
	}
	if !gshadowDatabase.FetchedItems {
		var err error
		gshadowDatabase.Items, err = getGroups()
		if err != nil {
			log.Err(err).Msg("unable to get groups")
			return C.NSS_STATUS_UNAVAIL
		}
	}

	if gshadowDatabase.ItemIndex+1 > len(gshadowDatabase.Items) {
		return C.NSS_STATUS_NOTFOUND
	}

	// store everything in buffer
	if err := StoreGroupInGShadowStruct(&gshadowDatabase.Items[gshadowDatabase.ItemIndex], sgbuf, buffer, buflen); err != nil {
		log.Err(err).Msg("unable to store user in buffer")
		return C.NSS_STATUS_UNAVAIL
	}
	gshadowDatabase.ItemIndex++
	return C.NSS_STATUS_SUCCESS
}

//export _nss_http_getsgnam_r
func _nss_http_getsgnam_r(name *C.char, result *C.struct_sgrp, buffer *C.char, buflen C.size_t, errnop *C.int) C.enum_nss_status {
	log.Debug().Msg("_nss_http_getsgnam_r")

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
	if err := StoreGroupInGShadowStruct(group, result, buffer, buflen); err != nil {
		log.Err(err).Str("name", groupName).Msg("unable to store group in buffer")
		return C.NSS_STATUS_UNAVAIL
	}

	return C.NSS_STATUS_SUCCESS
}
