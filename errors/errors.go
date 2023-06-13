// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"fmt"
)

const (
	ADD                          = 100
	ADD_IPV6_FORMAT              = 101
	ADD_VM_NAME_ERR              = 102
	ADD_VM_LPM4                  = 104
	ADD_VM_LPM6                  = 105
	ADD_VM_ADD_ROUT4             = 106
	ADD_VM_ADD_ROUT6             = 107
	ADD_VM_NO_VFS                = 108
	ALREADY_ALLOCATED            = 109
	CANT_GET_NAME                = 110
	ADD_VM_VNF_ERROR             = 111
	DEL                          = 150
	DEL_VM_NOT_FND               = 151
	GET_VM_NOT_FND               = 171
	LIST                         = 200
	ADD_RT                       = 250
	ADD_RT_FAIL4                 = 251
	ADD_RT_FAIL6                 = 252
	ADD_RT_NO_VM                 = 253
	DEL_RT                       = 300
	GET_NETNAT_ITER_ERROR        = 349
	ADD_NAT                      = 350
	ADD_NAT_IP_EXISTS            = 351
	ADD_NAT_ALLOC                = 352
	ADD_NAT_ADD_KEY              = 353
	ADD_NET_NAT_DATA             = 354
	ADD_NETWORK_NAT              = 355
	DEL_NETWORK_NAT              = 356
	ADD_NETNAT_NONLOCAL          = 357
	ADD_NETNAT_INVALID_PORT      = 358
	ADD_NETNAT_DATA_NOT_FOUND    = 359
	DEL_NETNAT_NONLOCAL          = 360
	DEL_NETNAT_INVALID_PORT      = 361
	DEL_NETNAT_ENTRY_NOT_FOUND   = 362
	ADD_NETNAT_IP_EXISTS         = 363
	ADD_NETNAT_KEY               = 364
	ADD_NETNAT_ALLO_DATA         = 365
	ADD_NETNAT_ADD_DATA          = 366
	DEL_NETNAT_KEY_DELETED       = 367
	GET_NETNAT_IPV6_UNSUPPORTED  = 369
	ADD_NEIGHNAT_WRONGTYPE       = 370
	DEL_NEIGHNAT_WRONGTYPE       = 371
	ADD_NEIGHNAT_ENTRY_EXIST     = 372
	ADD_NEIGHNAT_ALLOC           = 373
	DEL_NEIGHNAT_ENTRY_NOFOUND   = 374
	GET_NEIGHNAT_UNDER_IPV6      = 375
	GET_NETNAT_INFO_TYPE_UNKNOWN = 376
	ADD_NAT_VNF_ERR              = 377
	ADD_DNAT                     = 400
	ADD_DNAT_IP_EXISTS           = 401
	ADD_DNAT_ALLOC               = 402
	ADD_DNAT_ADD_KEY             = 403
	ADD_DNAT_ADD_DATA            = 404
	DEL_NAT                      = 450
	DEL_NAT_NO_SNAT              = 451
	GET_NAT                      = 500
	GET_NAT_NO_IP_SET            = 501
	ADD_LB_VIP                   = 550
	ADD_LB_NO_VNI_EXIST          = 551
	ADD_LB_UNSUPP_IP             = 552
	DEL_LB_VIP                   = 600
	DEL_LB_NO_VNI_EXIST          = 601
	DEL_LB_UNSUPP_IP             = 602
	ADD_PFX                      = 650
	ADD_PFX_NO_VM                = 651
	ADD_PFX_ROUTE                = 652
	ADD_PFX_VNF_ERR              = 653
	DEL_PFX                      = 700
	DEL_PFX_NO_VM                = 701
	CREATE_LB_UNSUPP_IP          = 750
	CREATE_LB_ERR                = 751
	CREATE_LB_VNF_ERR            = 752
	DEL_LB_ID_ERR                = 755
	DEL_LB_BACK_IP_ERR           = 756
	GET_LB_ID_ERR                = 760
	GET_LB_BACK_IP_ERR           = 761
	ADD_FWALL_ERR                = 800
	ADD_FWALL_RULE_ERR           = 801
	ADD_FWALL_NO_DROP_SUPPORT    = 802
	ADD_FWALL_ID_EXISTS          = 803
	GET_FWALL_ERR                = 810
	GET_NO_FWALL_RULE_ERR        = 811
	DEL_FWALL_ERR                = 820
	DEL_NO_FWALL_RULE_ERR        = 821

	// os.Exit value
	CLIENT_ERROR = 1
	SERVER_ERROR = 2
)

var ErrServerError = fmt.Errorf("server error")

type StatusError struct {
	errorCode int32
	message   string
}

func (s *StatusError) Message() string {
	return s.message
}

func (s *StatusError) ErrorCode() int32 {
	return s.errorCode
}

func (s *StatusError) Error() string {
	if s.message != "" {
		return fmt.Sprintf("[error code %d] %s", s.errorCode, s.message)
	}
	return fmt.Sprintf("error code %d", s.errorCode)
}

func NewStatusError(errorCode int32, message string) *StatusError {
	return &StatusError{
		errorCode: errorCode,
		message:   message,
	}
}

func IsStatusErrorCode(err error, errorCodes ...int32) bool {
	statusError := &StatusError{}
	if !errors.As(err, &statusError) {
		return false
	}

	for _, errorCode := range errorCodes {
		if statusError.ErrorCode() == errorCode {
			return true
		}
	}
	return false
}

func IgnoreStatusErrorCode(err error, errorCode int32) error {
	if IsStatusErrorCode(err, errorCode) {
		return nil
	}
	return err
}
