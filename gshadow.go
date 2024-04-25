package main

// #include <stdlib.h>
// #include <errno.h>
// #include <nss.h>
// #include <gshadow.h>
import "C"
import (
	"github.com/Eun/nss_http/types"
	"github.com/Eun/nss_http/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var gshadowDatabase utils.Database[types.Group]

//export _nss_http_setsgent
func _nss_http_setsgent() (status C.enum_nss_status) {
	log.Trace().Msg("setsgent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("setsgent")
	return errorToNSSStatus(gshadowDatabase.Open())
}

//export _nss_http_endsgent
func _nss_http_endsgent() (status C.enum_nss_status) {
	log.Trace().Msg("endsgent")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("endsgent")
	return errorToNSSStatus(gshadowDatabase.Close())
}

// //export _nss_http_getsgent
// func _nss_http_getsgent() *C.struct_sgrp {
// 	log.Trace().Msg("getsgent")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getsgent")
// 	panic("not implemented")
// }

//export _nss_http_getsgent_r
func _nss_http_getsgent_r(resultBuf *C.struct_sgrp, buffer *C.char, bufLen C.size_t, result **C.struct_sgrp) (status C.enum_nss_status) {
	log.Trace().Msg("getsgent_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getsgent_r")
	*result = nil
	status = errorToNSSStatus(
		gshadowDatabase.GetEnt(
			func() ([]types.Group, error) {
				return getGroups(ShadowRetrivalMode)
			},
			func(item *types.Group) error {
				return StoreGroupInGShadowStruct(item, resultBuf, buffer, bufLen)
			},
		),
	)
	if status == C.NSS_STATUS_SUCCESS {
		*result = resultBuf
	}
	return status
}

// //export _nss_http_getsgnam
// func _nss_http_getsgnam(name *C.char) *C.struct_sgrp {
// 	log.Trace().Msg("getsgnam")
// 	defer log.Debug().
// 		Int("errno", GetErrNo()).
// 		Msg("getsgnam")
// 	panic("not implemented")
// }

//export _nss_http_getsgnam_r
func _nss_http_getsgnam_r(name *C.char, resultBuf *C.struct_sgrp, buffer *C.char, bufLen C.size_t, result **C.struct_sgrp) (status C.enum_nss_status) {
	log.Trace().Msg("getsgnam_r")
	defer log.Debug().
		Int32("status", status).
		Int("errno", GetErrNo()).
		Msg("getsgnam_r")
	*result = nil
	groupName := C.GoString(name)
	group, err := getGroup(ShadowRetrivalMode, types.NameIdentifier(groupName))
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
	if err := StoreGroupInGShadowStruct(group, resultBuf, buffer, bufLen); err != nil {
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
