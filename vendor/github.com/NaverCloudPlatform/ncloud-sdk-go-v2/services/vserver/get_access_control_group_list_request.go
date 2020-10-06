/*
 * vserver
 *
 * VPC Compute 관련 API<br/>https://ncloud.apigw.ntruss.com/vserver/v2
 *
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package vserver

type GetAccessControlGroupListRequest struct {

	// REGION코드
RegionCode *string `json:"regionCode,omitempty"`

	// ACG번호리스트
AccessControlGroupNoList []*string `json:"accessControlGroupNoList,omitempty"`

	// ACG이름
AccessControlGroupName *string `json:"accessControlGroupName,omitempty"`

	// ACG상태코드
AccessControlGroupStatusCode *string `json:"accessControlGroupStatusCode,omitempty"`

	// 페이지번호
PageNo *int32 `json:"pageNo,omitempty"`

	// 페이지사이즈
PageSize *int32 `json:"pageSize,omitempty"`

	// VPC번호
VpcNo *string `json:"vpcNo,omitempty"`
}
