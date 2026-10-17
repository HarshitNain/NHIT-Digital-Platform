package grpc

import (
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// toProtoVendorAccount converts domain vendor account to protobuf
func toProtoVendorAccount(acc *domain.VendorAccount) *vendorpb.VendorAccount {
	return &vendorpb.VendorAccount{
		Id:            acc.ID.String(),
		VendorId:      acc.VendorID.String(),
		AccountName:   acc.AccountName,
		AccountNumber: acc.AccountNumber,
		AccountType:   acc.AccountType,
		NameOfBank:    acc.NameOfBank,
		BranchName:    acc.BranchName,
		IfscCode:      acc.IFSCCode,
		SwiftCode:     acc.SwiftCode,
		IsPrimary:     acc.IsPrimary,
		IsActive:      acc.IsActive,
		Remarks:       acc.Remarks,
		CreatedBy:     acc.CreatedBy.String(),
		CreatedAt:     timestamppb.New(acc.CreatedAt),
		UpdatedAt:     timestamppb.New(acc.UpdatedAt),
	}
}
