package account

import (
	"context"
	"database/sql"
	"errors"

	"waitq/api/internal/entities/account"
	helper "waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type httpHandler struct {
	accountService service.AccountService
	logger         *zap.Logger
}

func newHTTPHandler(accountService service.AccountService, logger *zap.Logger) *httpHandler {
	return &httpHandler{
		accountService: accountService,
		logger:         logger,
	}
}

type PathUUIDParam struct {
	ID uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
}

type GetAccountByIDOutput struct {
	Body struct {
		Message string           `json:"message"`
		Account *account.Account `json:"account"`
	}
}

func (h *httpHandler) getByID(ctx context.Context, input *PathUUIDParam) (*GetAccountByIDOutput, error) {
	if user := helper.GetAuthenticatedUser(ctx); user.ID != input.ID {
		h.logger.Error("unauthorized access", zap.Any("user", user), zap.Any("account", input.ID))
		return nil, huma.Error403Forbidden("Cannot access account")
	}

	account, err := h.accountService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Account not found")
		default:
			h.logger.Error("failed to fetch account", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the account")
		}
	}

	resp := &GetAccountByIDOutput{}
	resp.Body.Message = "Account fetched successfully"
	resp.Body.Account = &account

	return resp, nil
}

type GetAccountByUserIDInput struct {
	UserID         string `path:"userId"`
	IncludeDeleted bool   `query:"includeDeleted" required:"false" default:"false"`
}

type GetAccountByUserIDOutput struct {
	Body struct {
		Message  string           `json:"message"`
		Accounts *account.Account `json:"accounts"`
	}
}

func (h *httpHandler) getByUserID(ctx context.Context, input *GetAccountByUserIDInput) (*GetAccountByUserIDOutput, error) {
	h.logger.Info("getByUserID")
	userID, err := uuid.Parse(input.UserID) // fetching and validation input
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid account ID")
	}

	if user := helper.GetAuthenticatedUser(ctx); user.ID != userID {
		h.logger.Error("unauthorized access", zap.Any("user", user), zap.Any("userID", userID))
		return nil, huma.Error403Forbidden("Cannot access accounts")
	}

	accounts, err := h.accountService.GetByUserId(ctx, userID, shared.GetManyRequest{
		IncludeDeleted: input.IncludeDeleted,
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Accounts not found")
		default:
			h.logger.Error("failed to fetch accounts", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the accounts")
		}
	}

	resp := &GetAccountByUserIDOutput{}
	resp.Body.Message = "Accounts fetched successfully"
	resp.Body.Accounts = &accounts

	return resp, nil
}

type CreateAccountInput struct {
	Body struct {
		CreateAccountFields struct {
			Name   string    `json:"name" minLength:"3" maxLength:"100"`
			Email  string    `json:"email" minLength:"3" maxLength:"100" format:"email"`
			UserID uuid.UUID `json:"userId" minLength:"36" maxLength:"36" format:"uuid"`
		} `json:"account"`
	}
}

type CreateAccountOutput struct {
	Body struct {
		Message string           `json:"message"`
		Account *account.Account `json:"account"`
	}
}

func (h *httpHandler) create(ctx context.Context, input *CreateAccountInput) (*CreateAccountOutput, error) {
	if user := helper.GetAuthenticatedUser(ctx); user.ID != input.Body.CreateAccountFields.UserID {
		h.logger.Error("unauthorized access", zap.Any("user", user), zap.Any("input", input))
		return nil, huma.Error403Forbidden("Cannot create account for another user")
	}

	account, err := h.accountService.Create(ctx, account.Account{
		Name:   input.Body.CreateAccountFields.Name,
		Email:  input.Body.CreateAccountFields.Email,
		UserID: input.Body.CreateAccountFields.UserID,
	})

	if err != nil {
		h.logger.Error("failed to create account", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while creating the account")
	}

	resp := &CreateAccountOutput{}
	resp.Body.Message = "Account created successfully"
	resp.Body.Account = &account

	return resp, nil
}

type UpdateAccountInput struct {
	PathUUIDParam
	Body struct {
		UpdateAccountFields struct {
			Name  string `json:"name" minLength:"3" maxLength:"100"`
			Email string `json:"email" minLength:"3" maxLength:"100" format:"email"`
		} `json:"account"`
	}
}

type UpdateAccountOutput struct {
	Body struct {
		Message string           `json:"message"`
		Account *account.Account `json:"account"`
	}
}

func (h *httpHandler) update(ctx context.Context, input *UpdateAccountInput) (*UpdateAccountOutput, error) {
	accountRecord, err := h.accountService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Account not found")
		default:
			h.logger.Error("failed to fetch account", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the account")
		}
	}

	if accountRecord.DeletedAt != nil {
		return nil, huma.Error404NotFound("Account not found")
	}

	if user := helper.GetAuthenticatedUser(ctx); user.ID != accountRecord.UserID {
		h.logger.Error("unauthorized access", zap.Any("user", user), zap.Any("account", accountRecord))
		return nil, huma.Error403Forbidden("Cannot update account for another user")
	}

	account, err := h.accountService.Update(ctx, account.Account{
		ID:    input.ID,
		Name:  input.Body.UpdateAccountFields.Name,
		Email: input.Body.UpdateAccountFields.Email,
	})

	if err != nil {
		h.logger.Error("failed to update account", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the account")
	}

	resp := &UpdateAccountOutput{}
	resp.Body.Message = "Account updated successfully"
	resp.Body.Account = &account

	return resp, nil
}

type DeleteAccountInput struct {
	PathUUIDParam
}

type DeleteAccountOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) delete(ctx context.Context, input *DeleteAccountInput) (*DeleteAccountOutput, error) {
	account, err := h.accountService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Account not found")
		default:
			h.logger.Error("failed to fetch account", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the account")
		}
	}

	if user := helper.GetAuthenticatedUser(ctx); user.ID != account.UserID {
		h.logger.Error("unauthorized access", zap.Any("user", user), zap.Any("account", account))
		return nil, huma.Error403Forbidden("Cannot delete account for another user")
	}

	err = h.accountService.Delete(ctx, input.ID)

	if err != nil {
		h.logger.Error("failed to delete account", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while deleting the account")
	}

	resp := &DeleteAccountOutput{}
	resp.Body.Message = "Account deleted successfully"

	return resp, nil
}
