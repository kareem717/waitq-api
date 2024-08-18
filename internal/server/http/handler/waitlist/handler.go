package waitlist

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"waitq/api/internal/entities/waitlist"
	helper "waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type httpHandler struct {
	waitlistService service.WaitlistService
	logger          *zap.Logger
}

func newHTTPHandler(waitlistService service.WaitlistService, logger *zap.Logger) *httpHandler {
	return &httpHandler{
		waitlistService: waitlistService,
		logger:          logger,
	}
}

type IDPathParam struct {
	ID uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
}

type WaitlistWithMessage struct {
	Body struct {
		Message  string             `json:"message"`
		Waitlist *waitlist.Waitlist `json:"waitlist"`
	}
}

func (h *httpHandler) getByID(ctx context.Context, input *IDPathParam) (*WaitlistWithMessage, error) {
	if key := helper.GetWaitlistServiceKey(ctx); key != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyId", key), zap.Any("waitlistId", input.ID))
		return nil, huma.Error403Forbidden("Cannot access waitlist")
	}

	waitlist, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	resp := &WaitlistWithMessage{}
	resp.Body.Message = "Waitlist fetched successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

func (h *httpHandler) getApiKeys(ctx context.Context, input *IDPathParam) (*WaitlistWithMessage, error) {
	waitlist, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlist.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlist))
		return nil, huma.Error403Forbidden("Cannot access waitlist detials owned by another account")
	}

	resp := &WaitlistWithMessage{}
	resp.Body.Message = "Waitlist API keys fetched successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type CreateWaitlistInput struct {
	Body struct {
		CreateWaitlistFields struct {
			Name      string    `json:"name" minLength:"3" maxLength:"100"`
			AccountID uuid.UUID `json:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
		} `json:"waitlist"`
	}
}

func (h *httpHandler) create(ctx context.Context, input *CreateWaitlistInput) (*WaitlistWithMessage, error) {
	if account := helper.GetAuthenticatedAccount(ctx); account.ID != input.Body.CreateWaitlistFields.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", account), zap.Any("input", input))
		return nil, huma.Error403Forbidden("Cannot create waitlist for another account")
	}

	waitlist, err := h.waitlistService.Create(ctx, waitlist.Waitlist{
		Name:      input.Body.CreateWaitlistFields.Name,
		AccountID: input.Body.CreateWaitlistFields.AccountID,
	})

	if err != nil {
		h.logger.Error("failed to create waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while creating the waitlist")
	}

	resp := &WaitlistWithMessage{}
	resp.Body.Message = "Waitlist created successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type UnsubscribeEmailRequest struct {
	Email string `json:"email" minLength:"3" maxLength:"320" format:"email"`
}

type GetWaitlistByAccountIDInput struct {
	AccountID string `path:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
	Body      struct {
		PaginationParams shared.PaginationRequest `json:"paginationParams"`
	}
}

type GetWaitlistByAccountIDOutput struct {
	Body struct {
		Count    int                  `json:"count"`
		Message  string               `json:"message"`
		Waitlist *[]waitlist.Waitlist `json:"waitlists"`
	}
}

func (h *httpHandler) getByAccountID(ctx context.Context, input *GetWaitlistByAccountIDInput) (*GetWaitlistByAccountIDOutput, error) {
	accountID, err := uuid.Parse(input.AccountID) // fetching and validation input
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid account ID")
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != accountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("accountID", accountID))
		return nil, huma.Error403Forbidden("Cannot access accounts")
	}

	waitlists, err := h.waitlistService.GetByAccountId(ctx, accountID, input.Body.PaginationParams)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Accounts not found")
		default:
			h.logger.Error("failed to fetch accounts", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the accounts")
		}
	}

	resp := &GetWaitlistByAccountIDOutput{}
	resp.Body.Message = "Waitlists fetched successfully"
	resp.Body.Waitlist = &waitlists
	resp.Body.Count = len(*resp.Body.Waitlist)

	return resp, nil
}

type UpdateWaitlistInput struct {
	IDPathParam
	Body struct {
		UpdateWaitlistFields struct {
			Name string `json:"name" minLength:"3" maxLength:"100"`
		} `json:"waitlist"`
	}
}

func (h *httpHandler) update(ctx context.Context, input *UpdateWaitlistInput) (*WaitlistWithMessage, error) {
	if key := helper.GetWaitlistServiceKey(ctx); key != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyId", key), zap.Any("waitlistId", input.ID))
		return nil, huma.Error403Forbidden("Cannot update waitlist for another account")
	}

	waitlistRecord, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistRecord.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	waitlist, err := h.waitlistService.Update(ctx, input.ID, waitlist.Waitlist{
		Name: input.Body.UpdateWaitlistFields.Name,
	})

	if err != nil {
		h.logger.Error("failed to update waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the waitlist")
	}

	resp := &WaitlistWithMessage{}
	resp.Body.Message = "Waitlist updated successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type UpdateWaitlistJWTSecretInput struct {
	IDPathParam
	Body struct {
		UpdateWaitlistJWTSecretFields struct {
			JWTSecret string `json:"jwtSecret" minLength:"32" maxLength:"512" required:"false"`
		} `json:"waitlist"`
	}
}

func (h *httpHandler) updateJWTSecret(ctx context.Context, input *UpdateWaitlistJWTSecretInput) (*WaitlistWithMessage, error) {
	if key := helper.GetWaitlistServiceKey(ctx); key != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyId", key), zap.Any("waitlistId", input.ID))
		return nil, huma.Error403Forbidden("Cannot update waitlist for another account")
	}

	waitlistRecord, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistRecord.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	waitlist, err := h.waitlistService.UpdateJWTSecret(ctx, input.ID, input.Body.UpdateWaitlistJWTSecretFields.JWTSecret)
	if err != nil {
		h.logger.Error("failed to update waitlist jwt secret", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the waitlist jwt secret")
	}

	resp := &WaitlistWithMessage{}
	resp.Body.Message = "Waitlist JWT secret updated successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type MessageOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) delete(ctx context.Context, input *IDPathParam) (*MessageOutput, error) {
	if key := helper.GetWaitlistServiceKey(ctx); key != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyId", key), zap.Any("waitlistId", input.ID))
		return nil, huma.Error403Forbidden("Cannot delete waitlist for another account")
	}

	waitlist, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Account not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlist.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	err = h.waitlistService.Delete(ctx, input.ID)

	if err != nil {
		h.logger.Error("failed to delete account", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while deleting the account")
	}

	resp := &MessageOutput{}
	resp.Body.Message = "Waitlist deleted successfully"

	return resp, nil
}

type AddEmailsInput struct {
	IDPathParam
	Body struct {
		Email string `json:"emails" minLength:"3" maxLength:"320" format:"email"`
	}
}

type AddEmailsOutput struct {
	Body struct {
		Message    string         `json:"message"`
		EmailAdded waitlist.Email `json:"emailAdded"`
	}
}

func (h *httpHandler) addEmails(ctx context.Context, input *AddEmailsInput) (*AddEmailsOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistResp.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	email, err := h.waitlistService.AddEmail(ctx, waitlistResp.ID, input.Body.Email)
	if err != nil {
		h.logger.Error("failed to add emails to waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while adding emails to the waitlist")
	}

	resp := &AddEmailsOutput{}
	resp.Body.Message = "Emails added to waitlist successfully"
	resp.Body.EmailAdded = email

	return resp, nil
}

type DeleteEmailInput struct {
	IDPathParam
	Email string `query:"email" minLength:"3" maxLength:"320" format:"email"`
}

func (h *httpHandler) deleteEmail(ctx context.Context, input *DeleteEmailInput) (*MessageOutput, error) {
	if serviceKey := helper.GetWaitlistServiceKey(ctx); serviceKey != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyID", serviceKey), zap.Any("waitlistID", input.ID))
		return nil, huma.Error403Forbidden("Cannot delete email from waitlist for another account")
	}

	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistResp.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	err = h.waitlistService.DeleteEmail(ctx, input.ID, input.Email)
	if err != nil {
		h.logger.Error("failed to delete email from waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while deleting email from the waitlist")
	}

	resp := &MessageOutput{}
	resp.Body.Message = "Email deleted from waitlist successfully"
	return resp, nil
}

type GetEmailsByWaitlistIDInput struct {
	IDPathParam
	Body struct {
		PaginationParams shared.EmailPaginationRequest `json:"paginationParams"`
	}
}

type GetEmailsByWaitlistIDOutput struct {
	Body struct {
		Message string           `json:"message"`
		Emails  []waitlist.Email `json:"emails"`
		Count   int              `json:"count"`
	}
}

func (h *httpHandler) getEmailsByWaitlistID(ctx context.Context, input *GetEmailsByWaitlistIDInput) (*GetEmailsByWaitlistIDOutput, error) {
	if serviceKeyId := helper.GetWaitlistServiceKey(ctx); serviceKeyId != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyID", serviceKeyId), zap.Any("waitlistID", input.ID))
		return nil, huma.Error403Forbidden("Cannot add emails to waitlist for another account")
	}

	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistResp.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	emails, err := h.waitlistService.GetEmailsByWaitlistID(ctx, input.ID, input.Body.PaginationParams)
	if err != nil {
		h.logger.Error("failed to fetch emails", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the emails")
	}

	resp := &GetEmailsByWaitlistIDOutput{}
	resp.Body.Message = "Emails added to waitlist successfully"
	resp.Body.Emails = emails
	resp.Body.Count = len(emails)

	return resp, nil
}

type UnsubscribeEmailInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		EncodedEmail string `json:"encodedEmail"`
	}
}

type UnsubscribeEmailOutput struct {
	Body struct {
		Message string         `json:"message"`
		Email   waitlist.Email `json:"email"`
	}
}

func (h *httpHandler) unsubscribeEmail(ctx context.Context, input *UnsubscribeEmailInput) (*UnsubscribeEmailOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	token, err := jwt.Parse(input.Body.EncodedEmail, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.logger.Error("unexpected signing method", zap.Any("token", token.Header["alg"]))
			return nil, huma.Error400BadRequest("Invalid encoded email")
		}

		return []byte(waitlistResp.JWTSecret), nil
	})
	if err != nil {
		h.logger.Error("failed to parse token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while parsing the encoded email")
	}

	var inputEmail, waitlistID string
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		inputEmail = claims["email"].(string)
		waitlistID = claims["ref"].(string)
	} else {
		h.logger.Error("failed to parse token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while parsing the encoded email")
	}

	if waitlistID != input.ID.String() {
		return nil, huma.Error400BadRequest("Invalid waitlist ID")
	}

	email, err := h.waitlistService.GetEmailsByWaitlistIDAndEmail(ctx, waitlistResp.ID, inputEmail)
	if err != nil {
		h.logger.Error("failed to fetch email", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the email")
	}

	if email.WaitlistID != waitlistResp.ID {
		return nil, huma.Error400BadRequest("Email does not belong to waitlist")
	}

	if email.DeletedAt != nil {
		return nil, huma.Error400BadRequest("Email has already been removed from waitlist")
	}

	if email.UnsubscribedAt != nil {
		return nil, huma.Error400BadRequest("Email has already been unsubscribed")
	}

	emailResp, err := h.waitlistService.UnsubscribeEmail(ctx, waitlistResp.ID, inputEmail)
	if err != nil {
		h.logger.Error("failed to unsubscribe email from waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while unsubscribing email from the waitlist")
	}

	resp := &UnsubscribeEmailOutput{}
	resp.Body.Message = "Email unsubscribed from waitlist successfully"
	resp.Body.Email = emailResp

	return resp, nil
}

type GetUnsubscribedEmailJWTInput struct {
	IDPathParam
	Email string `query:"email" minLength:"3" maxLength:"320" format:"email"`
}

type GetUnsubscribedEmailJWTOutput struct {
	Body struct {
		Message string `json:"message"`
		Encoded string `json:"encoded"`
	}
}

func (h *httpHandler) getUnsubscribedEmailJWT(ctx context.Context, input *GetUnsubscribedEmailJWTInput) (*GetUnsubscribedEmailJWTOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot get unsubscribed email JWT for another account")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   "waitq",
		"ref":   input.ID.String(),
		"email": input.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 365 * 10).Unix(),
		"iat":   time.Now().Unix(),
	})

	// Ensure the secret is a byte slice for HMAC
	hmacSecret := []byte(waitlistResp.JWTSecret)

	// Sign and get the complete encoded token as a string using the HMAC secret
	encoded, err := token.SignedString(hmacSecret)
	if err != nil {
		h.logger.Error("failed to sign token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while signing the token")
	}

	resp := &GetUnsubscribedEmailJWTOutput{}
	resp.Body.Message = "Unsubscribed email JWT created successfully"
	resp.Body.Encoded = encoded

	return resp, nil
}

type GetWaitlistAnalyticsOutput struct {
	Body struct {
		Message   string                     `json:"message"`
		Analytics waitlist.WaitlistAnalytics `json:"analytics"`
	}
}

func (h *httpHandler) getWaitlistAnalytics(ctx context.Context, input *IDPathParam) (*GetWaitlistAnalyticsOutput, error) {
	if serviceKeyId := helper.GetWaitlistServiceKey(ctx); serviceKeyId != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyID", serviceKeyId), zap.Any("waitlistID", input.ID))
		return nil, huma.Error403Forbidden("Cannot get analytics for another account")
	}

	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistResp.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	analytics, err := h.waitlistService.GetAnalytics(ctx, input.ID)
	if err != nil {
		h.logger.Error("failed to fetch analytics", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the analytics")
	}

	resp := &GetWaitlistAnalyticsOutput{}
	resp.Body.Message = "Waitlist analytics fetched successfully"
	resp.Body.Analytics = analytics

	return resp, nil

}

type ExportEmailsToCSVOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

func (h *httpHandler) exportEmailsToCSV(ctx context.Context, input *IDPathParam) (*ExportEmailsToCSVOutput, error) {
	if serviceKeyId := helper.GetWaitlistServiceKey(ctx); serviceKeyId != input.ID.String() {
		h.logger.Error("unauthorized access", zap.Any("serviceKeyID", serviceKeyId), zap.Any("waitlistID", input.ID))
		return nil, huma.Error403Forbidden("Cannot export emails to CSV for another account")
	}

	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistResp.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	emailsChan, errChan := h.waitlistService.ExportEmails(ctx, input.ID)

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	for email := range emailsChan {
		if err := writer.Write([]string{email}); err != nil {
			h.logger.Error("failed to write email to CSV", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while writing emails to CSV")
		}
		writer.Flush()
	}

	if err := <-errChan; err != nil {
		h.logger.Error("failed to export emails", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while exporting emails")
	}

	resp := &ExportEmailsToCSVOutput{}
	resp.ContentType = "text/csv"
	resp.ContentDisposition = fmt.Sprintf("attachment; filename=%s.csv", waitlistResp.Name)
	resp.Body = buffer.Bytes()

	return resp, nil
}
