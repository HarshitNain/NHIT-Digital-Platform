package grpc

import (
	"context"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ForgotPasswordWithOTP initiates an OTP-based password reset flow
func (h *AuthHandler) ForgotPasswordWithOTP(ctx context.Context, req *authpb.ForgotPasswordOTPRequest) (*authpb.ForgotPasswordOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	// tenant_id is now optional - will be fetched from email
	if err := h.authService.ForgotPasswordWithOTPByEmail(ctx, req.Email); err != nil {
		return &authpb.ForgotPasswordOTPResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	return &authpb.ForgotPasswordOTPResponse{
		Message: "If the email exists, a password reset OTP has been sent",
		Success: true,
	}, nil
}

// VerifyOTPAndResetPassword verifies an OTP and resets the password
func (h *AuthHandler) VerifyOTPAndResetPassword(ctx context.Context, req *authpb.VerifyOTPAndResetPasswordRequest) (*authpb.VerifyOTPAndResetPasswordResponse, error) {
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Otp == "" {
		return nil, status.Errorf(codes.InvalidArgument, "otp is required")
	}
	if req.NewPassword == "" {
		return nil, status.Errorf(codes.InvalidArgument, "new_password is required")
	}

	// tenant_id is now optional - will be fetched from email if not provided
	if req.TenantId == "" {
		if err := h.authService.VerifyOTPAndResetPasswordByEmail(ctx, req.Email, req.Otp, req.NewPassword); err != nil {
			return &authpb.VerifyOTPAndResetPasswordResponse{
				Message: err.Error(),
				Success: false,
			}, nil
		}
	} else {
		tenantID, err := uuid.Parse(req.TenantId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tenant ID: %v", err)
		}

		if err := h.authService.VerifyOTPAndResetPassword(ctx, req.Email, req.Otp, req.NewPassword, tenantID); err != nil {
			return &authpb.VerifyOTPAndResetPasswordResponse{
				Message: err.Error(),
				Success: false,
			}, nil
		}
	}

	return &authpb.VerifyOTPAndResetPasswordResponse{
		Message: "Password has been reset successfully",
		Success: true,
	}, nil
}
